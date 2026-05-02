//go:build windows

package agent

import (
	"context"
	"os/exec"
	"strings"

	"github.com/s4e-io/opservant-spark/internal/logger"
)

// detectAndExtractShell detects if command starts with a shell executable
// and returns the appropriate exec.Cmd for direct execution with context support.
func (e *PlaybookExecutor) detectAndExtractShell(ctx context.Context, command string) (*exec.Cmd, bool) {
	trimmedCmd := strings.TrimSpace(command)
	lowerCmd := strings.ToLower(trimmedCmd)

	if strings.HasPrefix(lowerCmd, "powershell.exe ") ||
		strings.HasPrefix(lowerCmd, "powershell ") ||
		strings.HasPrefix(lowerCmd, "pwsh.exe ") ||
		strings.HasPrefix(lowerCmd, "pwsh ") {
		return e.buildPowerShellCommand(ctx, trimmedCmd), true
	}

	if strings.HasPrefix(lowerCmd, "bash.exe ") || strings.HasPrefix(lowerCmd, "bash ") {
		return e.buildBashCommand(ctx, trimmedCmd), true
	}

	if strings.HasPrefix(lowerCmd, "wsl.exe ") ||
		strings.HasPrefix(lowerCmd, "wsl ") ||
		lowerCmd == "wsl.exe" ||
		lowerCmd == "wsl" {
		return e.buildWSLCommand(ctx, trimmedCmd), true
	}

	return nil, false
}

// buildPowerShellCommand builds a PowerShell command for direct execution with context support.
func (e *PlaybookExecutor) buildPowerShellCommand(_ context.Context, command string) *exec.Cmd {
	trimmedCmd := strings.TrimSpace(command)
	if trimmedCmd == "" {
		return nil
	}

	shellExe := "powershell.exe"
	lowerCmd := strings.ToLower(trimmedCmd)
	if strings.HasPrefix(lowerCmd, "pwsh.exe") || strings.HasPrefix(lowerCmd, "pwsh ") {
		shellExe = "pwsh.exe"
	}

	args := []string{"-NoProfile", "-NonInteractive"}

	shellPrefixes := []string{"powershell.exe ", "powershell ", "pwsh.exe ", "pwsh "}
	var restOfCommand string
	for _, prefix := range shellPrefixes {
		if strings.HasPrefix(lowerCmd, prefix) {
			restOfCommand = strings.TrimSpace(trimmedCmd[len(prefix):])
			break
		}
	}

	if restOfCommand == "" {
		e.logger.LogWithCategory(logger.DEBUG, "execution", "PowerShell command detected: %s", shellExe)
		return exec.Command(shellExe, args...)
	}

	lowerRest := strings.ToLower(restOfCommand)
	switch {
	case strings.HasPrefix(lowerRest, "-c "):
		args = append(args, "-Command", stripPowerShellQuotes(strings.TrimSpace(restOfCommand[3:])))
	case strings.HasPrefix(lowerRest, "-command "):
		args = append(args, "-Command", stripPowerShellQuotes(strings.TrimSpace(restOfCommand[9:])))
	default:
		args = append(args, tokenizeArgs(restOfCommand)...)
	}

	e.logger.LogWithCategory(logger.DEBUG, "execution", "PowerShell direct execution: %s %v", shellExe, args)
	return exec.Command(shellExe, args...)
}

// tokenizeArgs splits a command string into arguments, respecting double-quoted
// and single-quoted substrings so that spaces inside quotes don't split tokens.
func tokenizeArgs(s string) []string {
	var args []string
	var current strings.Builder
	inDouble := false
	inSingle := false

	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch {
		case ch == '"' && !inSingle:
			inDouble = !inDouble
		case ch == '\'' && !inDouble:
			inSingle = !inSingle
		case (ch == ' ' || ch == '\t') && !inDouble && !inSingle:
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(ch)
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}

func stripPowerShellQuotes(s string) string {
	if strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'") {
		s = s[1 : len(s)-1]
		return strings.ReplaceAll(s, "''", "'")
	}
	if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		s = s[1 : len(s)-1]
		s = strings.ReplaceAll(s, "\"\"", "\"")
		return strings.ReplaceAll(s, "`\"", "\"")
	}
	return s
}

// buildBashCommand builds a Bash command for direct execution with context support.
func (e *PlaybookExecutor) buildBashCommand(_ context.Context, command string) *exec.Cmd {
	parts := tokenizeArgs(command)
	if len(parts) == 0 {
		return nil
	}

	args := []string{}
	if len(parts) > 1 {
		args = parts[1:]
	}

	e.logger.LogWithCategory(logger.DEBUG, "execution", "Bash direct execution: bash.exe %v", args)
	return exec.Command("bash.exe", args...)
}

// buildWSLCommand builds a WSL command for direct execution with context support.
func (e *PlaybookExecutor) buildWSLCommand(_ context.Context, command string) *exec.Cmd {
	parts := tokenizeArgs(command)
	if len(parts) == 0 {
		return nil
	}

	args := []string{}
	if len(parts) > 1 {
		args = parts[1:]
	}

	e.logger.LogWithCategory(logger.DEBUG, "execution", "WSL direct execution: wsl.exe %v", args)
	return exec.Command("wsl.exe", args...)
}
