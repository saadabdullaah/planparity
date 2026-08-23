package runner

import (
	"strings"
	"testing"
)

func TestAllowlistContainsNoMutatingCommands(t *testing.T) {
	t.Parallel()
	requests := []Request{
		{Operation: OperationVersion},
		{Operation: OperationValidate},
		{Operation: OperationShow, TargetPath: "state.json"},
		{Operation: OperationProviderSchemas},
		{Operation: OperationInit},
		{Operation: OperationPlan, StatePath: "state.json", OutputPath: "plan.bin"},
	}
	for _, request := range requests {
		arguments, err := Arguments(request)
		if err != nil {
			t.Fatalf("%s: %v", request.Operation, err)
		}
		joined := " " + strings.ToLower(strings.Join(arguments, " ")) + " "
		for _, forbidden := range []string{" apply ", " destroy ", " state push ", " state rm ", " import ", " taint ", " untaint ", " force-unlock "} {
			if strings.Contains(joined, forbidden) {
				t.Fatalf("%s constructed forbidden arguments %q", request.Operation, arguments)
			}
		}
	}
}

func TestUnknownOperationIsRejected(t *testing.T) {
	t.Parallel()
	if _, err := Arguments(Request{Operation: Operation("APPLY")}); err == nil {
		t.Fatal("unknown operation was accepted")
	}
}

func FuzzArgumentsRejectUnknown(f *testing.F) {
	f.Add("APPLY")
	f.Add("state push")
	f.Fuzz(func(t *testing.T, operation string) {
		known := map[Operation]bool{
			OperationVersion: true, OperationValidate: true, OperationShow: true,
			OperationProviderSchemas: true, OperationInit: true, OperationPlan: true,
		}
		_, err := Arguments(Request{Operation: Operation(operation)})
		if !known[Operation(operation)] && err == nil {
			t.Fatalf("unknown operation %q accepted", operation)
		}
	})
}
