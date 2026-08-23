package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/saadabdullaah/planparity/internal/app"
	"github.com/saadabdullaah/planparity/internal/buildinfo"
	"github.com/saadabdullaah/planparity/internal/model"
)

func (server *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", server.handleHealth)
	mux.HandleFunc("GET /api/v1/system", server.handleSystem)
	mux.HandleFunc("GET /api/v1/engines", server.handleEngines)
	mux.HandleFunc("POST /api/v1/inspect", server.handleInspect)
	mux.HandleFunc("POST /api/v1/compare", server.handleCompare)
	mux.HandleFunc("POST /api/v1/rehearsals/plan", server.handlePlan)
	mux.HandleFunc("POST /api/v1/rehearsals/upgrade", server.handleUpgrade)
	mux.HandleFunc("POST /api/v1/rehearsals/migrate", server.handleMigrate)
	mux.HandleFunc("GET /api/v1/jobs/{id}", server.handleJob)
	mux.HandleFunc("GET /api/v1/jobs/{id}/events", server.handleEvents)
	mux.HandleFunc("DELETE /api/v1/jobs/{id}", server.handleCancelJob)
	mux.HandleFunc("GET /api/v1/results/{id}", server.handleResult)
	mux.HandleFunc("GET /api/v1/results/{id}/export", server.handleExport)
	mux.Handle("/", server.assets())
	return mux
}

func (server *Server) handleHealth(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{"status": "ok", "schema_version": model.SchemaV1})
}

func (server *Server) handleSystem(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{
		"schema_version": model.SchemaV1, "build": buildinfo.Current(), "os": runtime.GOOS, "arch": runtime.GOARCH,
		"execution_mode": model.LocalRestricted, "result_persistence": "memory-only",
	})
}

func (server *Server) handleEngines(writer http.ResponseWriter, request *http.Request) {
	ctx, cancel := contextWithTimeout(request, 20*time.Second)
	defer cancel()
	statuses := server.service.EngineStatuses(ctx, server.options.TerraformPath, server.options.OpenTofuPath)
	writeJSON(writer, http.StatusOK, map[string]any{"schema_version": model.SchemaV1, "engines": statuses})
}

type inspectAPIRequest struct {
	Engine     string `json:"engine"`
	EnginePath string `json:"engine_path,omitempty"`
	State      string `json:"state"`
}

type compareAPIRequest struct {
	Engine                string  `json:"engine"`
	EnginePath            string  `json:"engine_path,omitempty"`
	Before                string  `json:"before"`
	After                 string  `json:"after"`
	ProbableMoveThreshold float64 `json:"probable_move_threshold,omitempty"`
}

type planAPIRequest struct {
	Engine              string `json:"engine"`
	EnginePath          string `json:"engine_path,omitempty"`
	Config              string `json:"config"`
	State               string `json:"state"`
	InheritProxy        bool   `json:"inherit_proxy,omitempty"`
	AllowRegistryAccess bool   `json:"allow_registry_access"`
}

type upgradeAPIRequest struct {
	FromEngine          string `json:"from_engine"`
	FromEnginePath      string `json:"from_engine_path"`
	ToEngine            string `json:"to_engine"`
	ToEnginePath        string `json:"to_engine_path"`
	Config              string `json:"config"`
	State               string `json:"state"`
	InheritProxy        bool   `json:"inherit_proxy,omitempty"`
	AllowRegistryAccess bool   `json:"allow_registry_access"`
}

type migrateAPIRequest struct {
	TerraformPath       string `json:"terraform_path"`
	OpenTofuPath        string `json:"opentofu_path"`
	Config              string `json:"config"`
	State               string `json:"state"`
	InheritProxy        bool   `json:"inherit_proxy,omitempty"`
	AllowRegistryAccess bool   `json:"allow_registry_access"`
}

func (server *Server) handleInspect(writer http.ResponseWriter, request *http.Request) {
	var input inspectAPIRequest
	if !server.decode(writer, request, &input) {
		return
	}
	kind, err := parseKind(input.Engine)
	if err != nil || input.State == "" {
		writeAPIError(writer, http.StatusBadRequest, "INVALID_REQUEST", "engine and state are required.")
		return
	}
	statePath, ok := server.authorizeAPIPath(writer, input.State)
	if !ok {
		return
	}
	server.startJob(writer, "inspect", func(ctx context.Context, emit func(string, string)) (any, error) {
		emit("engine.detecting", "Detecting engine")
		emit("state.normalizing", "Normalizing state")
		result, err := server.service.Inspect(ctx, app.InspectRequest{Engine: kind, EnginePath: input.EnginePath, StatePath: statePath})
		if err == nil {
			emit("report.finalizing", "Finalizing sanitized report")
		}
		return result, err
	})
}

func (server *Server) handleCompare(writer http.ResponseWriter, request *http.Request) {
	var input compareAPIRequest
	if !server.decode(writer, request, &input) {
		return
	}
	kind, err := parseKind(input.Engine)
	if err != nil || input.Before == "" || input.After == "" {
		writeAPIError(writer, http.StatusBadRequest, "INVALID_REQUEST", "engine, before, and after are required.")
		return
	}
	threshold := input.ProbableMoveThreshold
	if threshold == 0 {
		threshold = 0.85
	}
	if threshold < 0.5 || threshold > 1 {
		writeAPIError(writer, http.StatusBadRequest, "INVALID_REQUEST", "probable_move_threshold must be between 0.5 and 1.")
		return
	}
	beforePath, ok := server.authorizeAPIPath(writer, input.Before)
	if !ok {
		return
	}
	afterPath, ok := server.authorizeAPIPath(writer, input.After)
	if !ok {
		return
	}
	server.startJob(writer, "compare", func(ctx context.Context, emit func(string, string)) (any, error) {
		emit("engine.detecting", "Detecting engine")
		emit("state.normalizing", "Normalizing snapshots")
		emit("comparison.running", "Comparing resources")
		result, err := server.service.Compare(ctx, app.CompareRequest{Engine: kind, EnginePath: input.EnginePath, BeforePath: beforePath, AfterPath: afterPath, ProbableMoveThreshold: threshold})
		if err == nil {
			emit("report.finalizing", "Finalizing sanitized report")
		}
		return result, err
	})
}

func (server *Server) handlePlan(writer http.ResponseWriter, request *http.Request) {
	var input planAPIRequest
	if !server.decode(writer, request, &input) {
		return
	}
	kind, err := parseKind(input.Engine)
	if err != nil || input.Config == "" || input.State == "" {
		writeAPIError(writer, http.StatusBadRequest, "INVALID_REQUEST", "engine, config, and state are required.")
		return
	}
	configPath, ok := server.authorizeAPIPath(writer, input.Config)
	if !ok {
		return
	}
	statePath, ok := server.authorizeAPIPath(writer, input.State)
	if !ok {
		return
	}
	server.startJob(writer, "rehearsal.plan", func(ctx context.Context, emit func(string, string)) (any, error) {
		emit("workspace.preparing", "Preparing copied workspace")
		emit("engine.initializing", "Initializing engine without backend")
		emit("plan.generating", "Generating refresh-disabled plan")
		return server.service.RehearsePlan(ctx, app.PlanRehearsalRequest{Engine: kind, EnginePath: input.EnginePath, ConfigPath: configPath, StatePath: statePath, InheritProxy: input.InheritProxy, AllowRegistryAccess: input.AllowRegistryAccess, PathRoot: server.PathRoot()})
	})
}

func (server *Server) handleUpgrade(writer http.ResponseWriter, request *http.Request) {
	var input upgradeAPIRequest
	if !server.decode(writer, request, &input) {
		return
	}
	fromKind, fromErr := parseKind(input.FromEngine)
	toKind, toErr := parseKind(input.ToEngine)
	if fromErr != nil || toErr != nil || input.FromEnginePath == "" || input.ToEnginePath == "" || input.Config == "" || input.State == "" {
		writeAPIError(writer, http.StatusBadRequest, "INVALID_REQUEST", "upgrade request is incomplete.")
		return
	}
	configPath, ok := server.authorizeAPIPath(writer, input.Config)
	if !ok {
		return
	}
	statePath, ok := server.authorizeAPIPath(writer, input.State)
	if !ok {
		return
	}
	server.startJob(writer, "rehearsal.upgrade", func(ctx context.Context, emit func(string, string)) (any, error) {
		emit("workspace.preparing", "Preparing isolated baseline and candidate workspaces")
		emit("plan.generating", "Generating baseline and candidate plans")
		emit("comparison.running", "Comparing engine behavior")
		return server.service.RehearseUpgrade(ctx, app.UpgradeRehearsalRequest{FromEngine: fromKind, FromEnginePath: input.FromEnginePath, ToEngine: toKind, ToEnginePath: input.ToEnginePath, ConfigPath: configPath, StatePath: statePath, InheritProxy: input.InheritProxy, AllowRegistryAccess: input.AllowRegistryAccess, PathRoot: server.PathRoot()})
	})
}

func (server *Server) handleMigrate(writer http.ResponseWriter, request *http.Request) {
	var input migrateAPIRequest
	if !server.decode(writer, request, &input) {
		return
	}
	if input.TerraformPath == "" || input.OpenTofuPath == "" || input.Config == "" || input.State == "" {
		writeAPIError(writer, http.StatusBadRequest, "INVALID_REQUEST", "migration request is incomplete.")
		return
	}
	configPath, ok := server.authorizeAPIPath(writer, input.Config)
	if !ok {
		return
	}
	statePath, ok := server.authorizeAPIPath(writer, input.State)
	if !ok {
		return
	}
	server.startJob(writer, "rehearsal.migrate", func(ctx context.Context, emit func(string, string)) (any, error) {
		emit("workspace.preparing", "Preparing Terraform and OpenTofu workspaces")
		emit("plan.generating", "Generating engine plans")
		emit("comparison.running", "Comparing Terraform and OpenTofu")
		return server.service.RehearseMigrate(ctx, app.MigrateRehearsalRequest{TerraformPath: input.TerraformPath, OpenTofuPath: input.OpenTofuPath, ConfigPath: configPath, StatePath: statePath, InheritProxy: input.InheritProxy, AllowRegistryAccess: input.AllowRegistryAccess, PathRoot: server.PathRoot()})
	})
}

func (server *Server) authorizeAPIPath(writer http.ResponseWriter, input string) (string, bool) {
	path, err := server.authorizeLocalPath(input)
	if err != nil {
		writeAPIError(writer, http.StatusBadRequest, "INVALID_PATH", "State and configuration paths must exist within the UI path root printed in the terminal.")
		return "", false
	}
	return path, true
}

func (server *Server) startJob(writer http.ResponseWriter, kind string, function app.JobFunc) {
	id, err := server.jobs.Start(kind, function)
	if err != nil {
		writeAPIError(writer, http.StatusInternalServerError, "JOB_START_FAILED", "Analysis could not be started.")
		return
	}
	writeJSON(writer, http.StatusAccepted, map[string]string{
		"job_id": id, "status_url": "/api/v1/jobs/" + id, "events_url": "/api/v1/jobs/" + id + "/events", "result_url": "/api/v1/results/" + id,
	})
}

func (server *Server) handleJob(writer http.ResponseWriter, request *http.Request) {
	snapshot, exists := server.jobs.Get(request.PathValue("id"))
	if !exists {
		writeAPIError(writer, http.StatusNotFound, "JOB_NOT_FOUND", "Job was not found.")
		return
	}
	writeJSON(writer, http.StatusOK, snapshot)
}

func (server *Server) handleCancelJob(writer http.ResponseWriter, request *http.Request) {
	if !server.jobs.Cancel(request.PathValue("id")) {
		writeAPIError(writer, http.StatusNotFound, "JOB_NOT_FOUND", "Job was not found.")
		return
	}
	writeJSON(writer, http.StatusAccepted, map[string]string{"status": "cancellation_requested"})
}

func (server *Server) handleResult(writer http.ResponseWriter, request *http.Request) {
	server.writeResult(writer, request.PathValue("id"), false)
}

func (server *Server) handleExport(writer http.ResponseWriter, request *http.Request) {
	server.writeResult(writer, request.PathValue("id"), true)
}

func (server *Server) writeResult(writer http.ResponseWriter, id string, download bool) {
	result, status, exists := server.jobs.Result(id)
	if !exists {
		writeAPIError(writer, http.StatusNotFound, "RESULT_NOT_FOUND", "Result was not found.")
		return
	}
	if status != app.JobCompleted {
		writeAPIError(writer, http.StatusConflict, "RESULT_NOT_READY", "Result is not available for this job status.")
		return
	}
	if download {
		writer.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="planparity-%s.json"`, id))
	}
	writeJSON(writer, http.StatusOK, result)
}

func (server *Server) decode(writer http.ResponseWriter, request *http.Request, destination any) bool {
	request.Body = http.MaxBytesReader(writer, request.Body, server.options.MaxRequestBytes)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		writeAPIError(writer, http.StatusBadRequest, "INVALID_JSON", "Request body is invalid or too large.")
		return false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeAPIError(writer, http.StatusBadRequest, "INVALID_JSON", "Request body must contain one JSON object.")
		return false
	}
	return true
}

func parseKind(value string) (model.EngineKind, error) {
	switch strings.ToLower(value) {
	case "terraform":
		return model.EngineTerraform, nil
	case "tofu", "opentofu":
		return model.EngineOpenTofu, nil
	default:
		return "", errors.New("unsupported engine")
	}
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}

func writeAPIError(writer http.ResponseWriter, status int, code, message string) {
	writeJSON(writer, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}

func contextWithTimeout(request *http.Request, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(request.Context(), timeout)
}
