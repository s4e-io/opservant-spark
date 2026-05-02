//go:build !windows

package agent

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"syscall"
)

// newCmdExec creates a cmd.exe command for non-Windows platforms (cross-compilation).
// This path is never reached at runtime since cmd is only used on Windows.
func newCmdExec(ctx context.Context, command string) *exec.Cmd {
	return exec.CommandContext(ctx, "cmd", "/C", command)
}

// runWithTimeout starts cmd, waits for completion or context expiry.
// Setpgid puts sh and all children in one process group so SIGKILL kills the full tree.
func runWithTimeout(cmd *exec.Cmd, ctx context.Context) (string, error) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("command failed to start: %w", err)
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		if err != nil {
			return buf.String(), fmt.Errorf("command execution failed: %w", err)
		}
		return buf.String(), nil
	case <-ctx.Done():
		if cmd.Process != nil {
			syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		<-done
		if ctx.Err() == context.DeadlineExceeded {
			return buf.String(), fmt.Errorf("command timeout exceeded: %w", ctx.Err())
		}
		return buf.String(), fmt.Errorf("command cancelled: %w", ctx.Err())
	}
}
