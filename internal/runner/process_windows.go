//go:build windows

package runner

import (
	"os/exec"
	"strconv"
	"syscall"
	"unsafe"
)

const (
	jobObjectExtendedLimitInformation = 9
	jobObjectLimitKillOnJobClose      = 0x00002000
	processSetQuota                   = 0x0100
	processTerminate                  = 0x0001
)

var (
	kernel32                 = syscall.NewLazyDLL("kernel32.dll")
	createJobObject          = kernel32.NewProc("CreateJobObjectW")
	setInformationJobObject  = kernel32.NewProc("SetInformationJobObject")
	assignProcessToJobObject = kernel32.NewProc("AssignProcessToJobObject")
	terminateJobObject       = kernel32.NewProc("TerminateJobObject")
	openProcess              = kernel32.NewProc("OpenProcess")
	closeHandle              = kernel32.NewProc("CloseHandle")
)

type jobObjectBasicLimitInformation struct {
	PerProcessUserTimeLimit int64
	PerJobUserTimeLimit     int64
	LimitFlags              uint32
	MinimumWorkingSetSize   uintptr
	MaximumWorkingSetSize   uintptr
	ActiveProcessLimit      uint32
	Affinity                uintptr
	PriorityClass           uint32
	SchedulingClass         uint32
}

type ioCounters struct {
	ReadOperationCount  uint64
	WriteOperationCount uint64
	OtherOperationCount uint64
	ReadTransferCount   uint64
	WriteTransferCount  uint64
	OtherTransferCount  uint64
}

type jobObjectExtendedLimit struct {
	BasicLimitInformation jobObjectBasicLimitInformation
	IoInfo                ioCounters
	ProcessMemoryLimit    uintptr
	JobMemoryLimit        uintptr
	PeakProcessMemoryUsed uintptr
	PeakJobMemoryUsed     uintptr
}

type processTree struct {
	job syscall.Handle
}

func configureProcess(*exec.Cmd) {}

func attachProcessTree(command *exec.Cmd) (*processTree, error) {
	job, _, createErr := createJobObject.Call(0, 0)
	if job == 0 {
		return nil, createErr
	}
	information := jobObjectExtendedLimit{}
	information.BasicLimitInformation.LimitFlags = jobObjectLimitKillOnJobClose
	set, _, setErr := setInformationJobObject.Call(job, jobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&information)), unsafe.Sizeof(information))
	if set == 0 {
		closeHandle.Call(job)
		return nil, setErr
	}
	process, _, openErr := openProcess.Call(processSetQuota|processTerminate, 0, uintptr(command.Process.Pid))
	if process == 0 {
		closeHandle.Call(job)
		return nil, openErr
	}
	assigned, _, assignErr := assignProcessToJobObject.Call(job, process)
	closeHandle.Call(process)
	if assigned == 0 {
		closeHandle.Call(job)
		return nil, assignErr
	}
	return &processTree{job: syscall.Handle(job)}, nil
}

func terminateProcessTree(command *exec.Cmd, tree *processTree) error {
	if command.Process != nil {
		// taskkill walks the current descendant tree, including a child that may
		// have started in the small interval before the parent joined the Job.
		if err := exec.Command("taskkill.exe", "/PID", strconv.Itoa(command.Process.Pid), "/T", "/F").Run(); err == nil {
			return nil
		}
	}
	if tree != nil && tree.job != 0 {
		terminated, _, terminateErr := terminateJobObject.Call(uintptr(tree.job), 1)
		if terminated != 0 {
			return nil
		}
		if terminateErr != syscall.Errno(0) {
			return terminateErr
		}
	}
	if command.Process == nil {
		return nil
	}
	return command.Process.Kill()
}

func closeProcessTree(tree *processTree) {
	if tree != nil && tree.job != 0 {
		closeHandle.Call(uintptr(tree.job))
		tree.job = 0
	}
}
