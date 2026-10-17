package Process

import (
	"fmt"
	"os"
	"strings"
	"syscall"
)

const LAUNCHER = "go"

func bindToParent(parent int) bool {
	raw, e := os.ReadFile(fmt.Sprintf("/proc/%d/comm", parent))

	if e != nil || strings.TrimSpace(string(raw)) != LAUNCHER {
		return false
	}

	syscall.RawSyscall(syscall.SYS_PRCTL, syscall.PR_SET_PDEATHSIG, uintptr(syscall.SIGTERM), 0)

	return true
}
