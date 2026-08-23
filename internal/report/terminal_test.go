package report

import (
	"bytes"
	"strings"
	"testing"

	"github.com/saadabdullaah/planparity/internal/model"
)

func TestInspectionTerminalStripsControlSequences(t *testing.T) {
	t.Parallel()
	result := model.InspectionResult{
		Snapshot: model.StateSnapshot{
			Engine:    model.EngineMetadata{Name: "Terraform\x1b[31m", Version: "1.0\rspoof"},
			Resources: []model.ResourceInstance{{Address: "safe\x1b]0;owned\a.address", Status: model.ResourceNormal, Mode: model.ResourceManaged}},
		},
		Findings: []model.Finding{{ID: "PP-TEST\x1b[2J", Title: "Synthetic\nTitle", Explanation: "Explanation\u202e"}},
	}
	var output bytes.Buffer
	if err := InspectionTerminal(&output, result); err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"\x1b", "\a", "\r", "\u202e"} {
		if strings.Contains(output.String(), forbidden) {
			t.Fatalf("terminal output retained control data %q: %q", forbidden, output.String())
		}
	}
}
