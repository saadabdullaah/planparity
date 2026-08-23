package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/saadabdullaah/planparity/internal/model"
)

func BenchmarkWriteJSONTenThousandResources(b *testing.B) {
	resources := make([]model.ResourceInstance, 10_000)
	for index := range resources {
		resources[index] = model.ResourceInstance{
			Address:         "module.inventory.aws_instance.node[\"benchmark\"]",
			ModuleAddress:   "module.inventory",
			Mode:            model.ResourceManaged,
			ResourceType:    "aws_instance",
			ResourceName:    "node",
			ProviderAddress: "registry.terraform.io/hashicorp/aws",
			Status:          model.ResourceNormal,
			Dependencies:    []string{"module.network.aws_subnet.private"},
			Structure: model.AttributeStructure{
				TypeFingerprint: "object",
				Keys:            []string{"ami", "instance_type", "subnet_id"},
			},
		}
	}
	result := model.InspectionResult{
		SchemaVersion: "1",
		Snapshot: model.StateSnapshot{
			SchemaVersion: "1",
			FormatVersion: "1.0",
			Engine:        model.EngineMetadata{Kind: model.EngineTerraform, Version: "1.9.0"},
			Resources:     resources,
			Providers:     []model.ProviderReference{},
			Modules:       []model.ModuleReference{},
			Outputs:       []model.OutputMetadata{},
		},
		Findings: []model.Finding{},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		recorder := httptest.NewRecorder()
		writeJSON(recorder, http.StatusOK, result)
		if recorder.Code != http.StatusOK {
			b.Fatalf("writeJSON status = %d, want %d", recorder.Code, http.StatusOK)
		}
	}
}
