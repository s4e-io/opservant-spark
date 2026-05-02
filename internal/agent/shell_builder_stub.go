//go:build !windows

package agent

import (
	"context"
	"os/exec"
)

func (e *PlaybookExecutor) detectAndExtractShell(_ context.Context, _ string) (*exec.Cmd, bool) {
	return nil, false
}
