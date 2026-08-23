package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/saadabdullaah/planparity/internal/app"
)

func newTestServer(t *testing.T, maxBytes int64) *Server {
	t.Helper()
	instance, err := New(app.NewService(nil, 100*time.Millisecond), Options{Host: "127.0.0.1", Port: 8765, MaxRequestBytes: maxBytes, AllowedPathRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(instance.jobs.Close)
	return instance
}

func request(instance *Server, method, target, token string, body io.Reader) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(method, "http://127.0.0.1:8765"+target, body)
	req.Host = "127.0.0.1:8765"
	if token != "" {
		req.Header.Set("X-PlanParity-Token", token)
	}
	instance.Handler().ServeHTTP(recorder, req)
	return recorder
}

func TestHealthHasSecurityHeaders(t *testing.T) {
	instance := newTestServer(t, 0)
	recorder := request(instance, http.MethodGet, "/api/v1/health", "", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	for _, header := range []string{"Content-Security-Policy", "Cache-Control", "X-Content-Type-Options", "X-Frame-Options", "Referrer-Policy"} {
		if recorder.Header().Get(header) == "" {
			t.Errorf("missing %s", header)
		}
	}
	if strings.Contains(recorder.Body.String(), instance.SessionToken()) {
		t.Fatal("health response exposed the session token")
	}
}

func TestRejectsHostOriginAndMissingToken(t *testing.T) {
	instance := newTestServer(t, 0)

	evilHost := httptest.NewRequest(http.MethodGet, "http://example.com/api/v1/health", nil)
	evilHost.Host = "example.com"
	hostRecorder := httptest.NewRecorder()
	instance.Handler().ServeHTTP(hostRecorder, evilHost)
	if hostRecorder.Code != http.StatusBadRequest {
		t.Fatalf("evil host status = %d", hostRecorder.Code)
	}

	originRequest := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:8765/api/v1/health", nil)
	originRequest.Host = "127.0.0.1:8765"
	originRequest.Header.Set("Origin", "http://evil.example")
	originRecorder := httptest.NewRecorder()
	instance.Handler().ServeHTTP(originRecorder, originRequest)
	if originRecorder.Code != http.StatusForbidden {
		t.Fatalf("evil origin status = %d", originRecorder.Code)
	}

	unauthorized := request(instance, http.MethodPost, "/api/v1/inspect", "", strings.NewReader(`{"engine":"terraform","state":"state.tfstate"}`))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("missing token status = %d", unauthorized.Code)
	}
}

func TestStrictRequestDecoding(t *testing.T) {
	instance := newTestServer(t, 80)
	unknown := request(instance, http.MethodPost, "/api/v1/inspect", instance.SessionToken(), strings.NewReader(`{"engine":"terraform","state":"state.tfstate","raw_state":"secret"}`))
	if unknown.Code != http.StatusBadRequest {
		t.Fatalf("unknown field status = %d", unknown.Code)
	}
	oversized := request(instance, http.MethodPost, "/api/v1/inspect", instance.SessionToken(), bytes.NewBufferString(`{"engine":"terraform","state":"`+strings.Repeat("x", 200)+`"}`))
	if oversized.Code != http.StatusBadRequest {
		t.Fatalf("oversized status = %d", oversized.Code)
	}
}

func TestFailedJobNeverReturnsInternalDetail(t *testing.T) {
	instance := newTestServer(t, 0)
	canary := "PP_SECRET_CANARY_NEVER_EXPOSE_7e53f9"
	statePath := filepath.Join(instance.PathRoot(), canary+".tfstate")
	if err := os.WriteFile(statePath, []byte(`{"version":4}`), 0o600); err != nil {
		t.Fatal(err)
	}
	payloadBytes, err := json.Marshal(map[string]string{"engine": "terraform", "engine_path": `C:\missing\terraform.exe`, "state": statePath})
	if err != nil {
		t.Fatal(err)
	}
	payload := string(payloadBytes)
	accepted := request(instance, http.MethodPost, "/api/v1/inspect", instance.SessionToken(), strings.NewReader(payload))
	if accepted.Code != http.StatusAccepted {
		t.Fatalf("start status = %d: %s", accepted.Code, accepted.Body.String())
	}
	var response struct {
		JobID string `json:"job_id"`
	}
	if err := json.Unmarshal(accepted.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	var statusBody string
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		status := request(instance, http.MethodGet, "/api/v1/jobs/"+response.JobID, instance.SessionToken(), nil)
		statusBody = status.Body.String()
		if strings.Contains(statusBody, `"status":"FAILED"`) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !strings.Contains(statusBody, `"status":"FAILED"`) {
		t.Fatalf("job did not fail: %s", statusBody)
	}
	if strings.Contains(statusBody, canary) || strings.Contains(statusBody, "C:\\missing") {
		t.Fatalf("job status leaked request detail: %s", statusBody)
	}
	result := request(instance, http.MethodGet, "/api/v1/results/"+response.JobID, instance.SessionToken(), nil)
	if result.Code != http.StatusConflict || strings.Contains(result.Body.String(), canary) {
		t.Fatalf("unexpected failed result response: %d %s", result.Code, result.Body.String())
	}
}

func TestRejectsStatePathsOutsideTheAuthorizedRoot(t *testing.T) {
	outer := t.TempDir()
	allowed := filepath.Join(outer, "allowed")
	if err := os.Mkdir(allowed, 0o700); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(outer, "outside.tfstate")
	if err := os.WriteFile(outside, []byte(`{"version":4}`), 0o600); err != nil {
		t.Fatal(err)
	}
	instance, err := New(app.NewService(nil, 100*time.Millisecond), Options{Host: "127.0.0.1", Port: 8765, AllowedPathRoot: allowed})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(instance.jobs.Close)

	payload, err := json.Marshal(map[string]string{"engine": "terraform", "state": outside})
	if err != nil {
		t.Fatal(err)
	}
	rejected := request(instance, http.MethodPost, "/api/v1/inspect", instance.SessionToken(), bytes.NewReader(payload))
	if rejected.Code != http.StatusBadRequest || !strings.Contains(rejected.Body.String(), `"code":"INVALID_PATH"`) {
		t.Fatalf("outside path status = %d: %s", rejected.Code, rejected.Body.String())
	}
	if strings.Contains(rejected.Body.String(), outside) {
		t.Fatal("path rejection exposed the supplied path")
	}

	link := filepath.Join(allowed, "linked.tfstate")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symbolic links are unavailable: %v", err)
	}
	linkPayload, err := json.Marshal(map[string]string{"engine": "terraform", "state": link})
	if err != nil {
		t.Fatal(err)
	}
	linked := request(instance, http.MethodPost, "/api/v1/inspect", instance.SessionToken(), bytes.NewReader(linkPayload))
	if linked.Code != http.StatusBadRequest {
		t.Fatalf("linked outside path status = %d: %s", linked.Code, linked.Body.String())
	}
}

func TestSSEReplaysTerminalEvents(t *testing.T) {
	instance := newTestServer(t, 0)
	id, err := instance.jobs.Start("test", func(context.Context, func(string, string)) (any, error) {
		return map[string]string{"status": "safe"}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		job, _ := instance.jobs.Get(id)
		if job.Status == app.JobCompleted {
			break
		}
		time.Sleep(time.Millisecond)
	}
	recorder := request(instance, http.MethodGet, "/api/v1/jobs/"+id+"/events", instance.SessionToken(), nil)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "event: job.completed") {
		t.Fatalf("unexpected SSE response: %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestAssetsCannotReadWorkspaceFiles(t *testing.T) {
	instance := newTestServer(t, 0)
	recorder := request(instance, http.MethodGet, "/../../testdata/security/canary.txt", "", nil)
	if (recorder.Code != http.StatusOK && recorder.Code != http.StatusTemporaryRedirect) || strings.Contains(recorder.Body.String(), "PP_SECRET_CANARY") {
		t.Fatalf("asset handler exposed workspace data: %d %s", recorder.Code, recorder.Body.String())
	}
}
