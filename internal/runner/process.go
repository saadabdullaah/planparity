package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

const defaultOutputLimit = 16 << 20 // 16 MiB per stream.

// Result contains bounded process output and execution metadata.
type Result struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
	Duration time.Duration
}

// ExecutionError is a sanitized process failure. Output is bounded and kept internal.
type ExecutionError struct {
	Operation Operation
	ExitCode  int
	TimedOut  bool
	Stderr    []byte
	Cause     error
}

func (e *ExecutionError) Error() string {
	if e.TimedOut {
		return fmt.Sprintf("%s timed out", e.Operation)
	}
	return fmt.Sprintf("%s failed with exit code %d", e.Operation, e.ExitCode)
}

func (e *ExecutionError) Unwrap() error { return e.Cause }

// Executor runs a typed request.
type Executor interface {
	Execute(context.Context, Request) (Result, error)
}

// ProcessExecutor invokes external engines with deadlines and bounded output.
type ProcessExecutor struct {
	Timeout     time.Duration
	OutputLimit int
}

// Execute runs an allowlisted command and never invokes a shell.
func (executor ProcessExecutor) Execute(parent context.Context, request Request) (Result, error) {
	arguments, err := Arguments(request)
	if err != nil {
		return Result{}, err
	}
	if request.Binary == "" {
		return Result{}, errors.New("engine binary path is required")
	}
	timeout := executor.Timeout
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	limit := executor.OutputLimit
	if limit <= 0 {
		limit = defaultOutputLimit
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	command := exec.Command(request.Binary, arguments...)
	command.Dir = request.WorkingDirectory
	command.Env = request.Environment
	configureProcess(command)
	stdout := newLimitedBuffer(limit)
	stderr := newLimitedBuffer(limit)
	command.Stdout = stdout
	command.Stderr = stderr

	started := time.Now()
	if err = command.Start(); err == nil {
		processTree, _ := attachProcessTree(command)
		defer closeProcessTree(processTree)
		done := make(chan error, 1)
		go func() { done <- command.Wait() }()
		select {
		case err = <-done:
		case <-ctx.Done():
			_ = terminateProcessTree(command, processTree)
			err = <-done
		}
	}
	result := Result{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), Duration: time.Since(started)}
	if command.ProcessState != nil {
		result.ExitCode = command.ProcessState.ExitCode()
	}
	if err == nil {
		return result, nil
	}
	executionError := &ExecutionError{
		Operation: request.Operation,
		ExitCode:  result.ExitCode,
		TimedOut:  errors.Is(ctx.Err(), context.DeadlineExceeded),
		Stderr:    append([]byte(nil), result.Stderr...),
		Cause:     err,
	}
	return result, executionError
}

type limitedBuffer struct {
	buffer bytes.Buffer
	limit  int
	full   bool
}

func newLimitedBuffer(limit int) *limitedBuffer { return &limitedBuffer{limit: limit} }

func (buffer *limitedBuffer) Write(data []byte) (int, error) {
	originalLength := len(data)
	remaining := buffer.limit - buffer.buffer.Len()
	if remaining <= 0 {
		buffer.full = true
		return originalLength, nil
	}
	if len(data) > remaining {
		data = data[:remaining]
		buffer.full = true
	}
	_, _ = buffer.buffer.Write(data)
	return originalLength, nil
}

func (buffer *limitedBuffer) Bytes() []byte {
	result := append([]byte(nil), buffer.buffer.Bytes()...)
	if buffer.full {
		result = append(result, []byte("\n[output truncated by PlanParity]\n")...)
	}
	return result
}
