//go:build windows

package agent

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"syscall"
)

// newCmdExec creates a cmd.exe command using SysProcAttr.CmdLine to bypass
// Go's argument escaping (EscapeArg) which mangles inner quotes.
// /S strips only the outermost quotes, preserving any inner quotes
// in the command (e.g. registry paths with spaces).
// CREATE_NEW_PROCESS_GROUP ensures child processes can be killed as a tree on timeout.
func newCmdExec(_ context.Context, command string) *exec.Cmd {
	cmd := exec.Command("cmd")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CmdLine:       `cmd /S /C "` + command + `"`,
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
	return cmd
}

// runWithTimeout starts cmd, waits for completion or context expiry.
// On timeout, kills the entire process tree before returning.
func runWithTimeout(cmd *exec.Cmd, ctx context.Context) (string, error) {
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
			exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(cmd.Process.Pid)).Run()
		}
		<-done
		if ctx.Err() == context.DeadlineExceeded {
			return buf.String(), fmt.Errorf("command timeout exceeded: %w", ctx.Err())
		}
		return buf.String(), fmt.Errorf("command cancelled: %w", ctx.Err())
	}
}
