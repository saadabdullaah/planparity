package runner

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	switch os.Getenv("PLANPARITY_TEST_HELPER_MODE") {
	case "parent":
		child := exec.Command(os.Args[0])
		child.Env = append(os.Environ(), "PLANPARITY_TEST_HELPER_MODE=child")
		_ = child.Start()
		for {
			time.Sleep(time.Second)
		}
	case "child":
		path := os.Getenv("PLANPARITY_TEST_HEARTBEAT")
		for index := 0; ; index++ {
			_ = os.WriteFile(path, []byte(strconv.Itoa(index)), 0o600)
			time.Sleep(40 * time.Millisecond)
		}
	default:
		os.Exit(m.Run())
	}
}

func TestLimitedBufferTruncates(t *testing.T) {
	t.Parallel()
	buffer := newLimitedBuffer(4)
	written, err := buffer.Write([]byte("abcdefgh"))
	if err != nil {
		t.Fatal(err)
	}
	if written != 8 {
		t.Fatalf("reported write=%d want=8", written)
	}
	if got := string(buffer.Bytes()); got != "abcd\n[output truncated by PlanParity]\n" {
		t.Fatalf("unexpected buffer: %q", got)
	}
}

func TestProcessExecutorTerminatesAProcessTreeOnTimeout(t *testing.T) {
	heartbeat := filepath.Join(t.TempDir(), "heartbeat")
	environment := append(os.Environ(),
		"PLANPARITY_TEST_HELPER_MODE=parent",
		"PLANPARITY_TEST_HEARTBEAT="+heartbeat,
	)
	executor := ProcessExecutor{Timeout: 2 * time.Second, OutputLimit: 1024}
	started := time.Now()
	_, err := executor.Execute(context.Background(), Request{
		Binary: os.Args[0], Operation: OperationVersion, Environment: environment,
	})
	var executionError *ExecutionError
	if !errors.As(err, &executionError) || !executionError.TimedOut {
		t.Fatalf("expected timeout, got %v", err)
	}
	if time.Since(started) > 5*time.Second {
		t.Fatalf("termination took too long: %s", time.Since(started))
	}
	time.Sleep(150 * time.Millisecond)
	before, readErr := os.ReadFile(heartbeat)
	if readErr != nil {
		t.Fatalf("child helper did not create heartbeat: %v", readErr)
	}
	time.Sleep(250 * time.Millisecond)
	after, readErr := os.ReadFile(heartbeat)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(before) != string(after) {
		t.Fatalf("child process survived parent timeout: %q then %q", before, after)
	}
}
