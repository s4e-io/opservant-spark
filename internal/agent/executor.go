package agent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/s4e-io/opservant-spark/internal/config"
	"github.com/s4e-io/opservant-spark/internal/logger"
	"github.com/s4e-io/opservant-spark/internal/models"
	"github.com/s4e-io/opservant-spark/internal/system"
)

// PlaybookExecutor handles playbook execution
type PlaybookExecutor struct {
	logger        *logger.Log
	config        *config.Config
	actionOutputs []models.ActionOutput
}

// NewPlaybookExecutor creates a new playbook executor
func NewPlaybookExecutor(logger *logger.Log, config *config.Config) *PlaybookExecutor {
	return &PlaybookExecutor{
		logger:        logger,
		config:        config,
		actionOutputs: make([]models.ActionOutput, 0),
	}
}

// Execute runs all tasks in the playbook sequentially, respecting dependencies.
func (e *PlaybookExecutor) Execute(ctx context.Context, playbook *models.Playbook) error {
	startTime := time.Now()

	e.actionOutputs = make([]models.ActionOutput, 0)

	playbookTimeout := time.Duration(playbook.Timeout) * time.Second
	if playbook.Timeout <= 0 {
		playbookTimeout = 1 * time.Hour
	}

	e.logger.LogPlaybookStart(playbook.Slug, playbook.RiskLevel, playbookTimeout)
	ctx, cancel := context.WithTimeout(ctx, playbookTimeout)
	defer cancel()

	success := true

	taskResults := make(map[string]bool)

	for _, task := range playbook.Tasks {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.Canceled) {
				e.logger.LogWithCategory(logger.INFO, "execution", "Playbook cancelled: %s", playbook.Slug)
				return fmt.Errorf("playbook cancelled")
			}
			e.logger.LogWithCategory(logger.ERROR, "execution", "Playbook timeout exceeded: %s", playbook.Slug)
			return fmt.Errorf("playbook timeout exceeded after %v", time.Since(startTime))
		default:
		}

		// Check if task dependencies are satisfied
		shouldExecute, reason := e.checkTaskDependencies(&task, taskResults)
		if !shouldExecute {
			e.logger.LogWithCategory(logger.WARN, "execution", "Task skipped: %s, reason: %s", task.Slug, reason)
			taskResults[task.Slug] = false // Mark as failed (skipped)
			success = false
			continue
		}

		executedSlugs, err := e.executeTask(ctx, &task, task.Variables)
		if err != nil {
			taskResults[task.Slug] = false
			success = false

			if playbook.AutoRevertOnFailure {
				e.logger.LogWithCategory(logger.INFO, "execution", "Auto-revert enabled, starting rollback")
				e.executeRollback(ctx, &task, task.Variables, executedSlugs)
			}
		} else {
			taskResults[task.Slug] = true
		}
	}

	if errors.Is(ctx.Err(), context.Canceled) {
		e.logger.LogWithCategory(logger.INFO, "execution", "Playbook cancelled: %s", playbook.Slug)
		return fmt.Errorf("playbook cancelled")
	}

	duration := time.Since(startTime)
	e.logger.LogPlaybookComplete(playbook.Slug, success, duration)

	if !success {
		return fmt.Errorf("playbook %q had failed tasks", playbook.Slug)
	}

	return nil
}

// executeTask runs all actions in the task, respecting dependencies and timeouts.
func (e *PlaybookExecutor) executeTask(parentCtx context.Context, task *models.Task, variables map[string]interface{}) (map[string]bool, error) {
	startTime := time.Now()
	e.logger.LogTaskStart(task.Slug, task.Name)

	taskTimeout := time.Duration(task.Timeout) * time.Second
	if task.Timeout <= 0 {
		taskTimeout = 10 * time.Minute
	}
	e.logger.LogWithCategory(logger.TRACE, "execution", "Task timeout: %v", taskTimeout)
	ctx, cancel := context.WithTimeout(parentCtx, taskTimeout)
	defer cancel()

	actionResults := make(map[string]bool)
	executedSlugs := make(map[string]bool)
	taskSuccess := true
	anyActionFailed := false
	totalActions := len(task.Actions)
	skippedActions := 0

	for _, action := range task.Actions {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.Canceled) {
				return executedSlugs, fmt.Errorf("task cancelled")
			}
			e.logger.LogWithCategory(logger.ERROR, "execution", "Task timeout exceeded: %s", task.Slug)
			return executedSlugs, fmt.Errorf("task timeout exceeded after %v", time.Since(startTime))
		default:
		}

		shouldExecute, reason := e.checkActionDependencies(&action, actionResults)
		if !shouldExecute {
			e.logger.LogActionSkipped(action.Slug, reason)
			actionResults[action.Slug] = false
			skippedActions++
			continue
		}

		if err := e.executeAction(ctx, &action, task, variables); err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				return executedSlugs, fmt.Errorf("task cancelled")
			}
			// Check if this is a platform skip error
			var skipErr ActionSkippedError
			if errors.As(err, &skipErr) {
				actionResults[action.Slug] = false // Mark as failed (skipped)
				skippedActions++

				// Create action output for skipped action so it appears in playbook_logs
				actionOutput := models.ActionOutput{
					TaskSlug:   task.Slug,
					ActionSlug: action.Slug,
					Command:    action.Command,
					Output:     fmt.Sprintf("Action skipped: %s", skipErr.Error()),
					Success:    false,
					Duration:   0,
					Timestamp:  time.Now().Format(time.RFC3339),
				}
				e.actionOutputs = append(e.actionOutputs, actionOutput)
			} else {
				actionResults[action.Slug] = false
				executedSlugs[action.Slug] = true
				anyActionFailed = true

				// Create action output for failed action so it appears in playbook_logs
				actionOutput := models.ActionOutput{
					TaskSlug:   task.Slug,
					ActionSlug: action.Slug,
					Command:    action.Command,
					Output:     fmt.Sprintf("Action failed: %v", err),
					Success:    false,
					Duration:   0,
					Timestamp:  time.Now().Format(time.RFC3339),
				}
				e.actionOutputs = append(e.actionOutputs, actionOutput)
			}

		} else {
			actionResults[action.Slug] = true
			executedSlugs[action.Slug] = true
		}
	}

	// Task fails if any action failed OR if all actions were skipped
	if anyActionFailed {
		taskSuccess = false
	} else if skippedActions == totalActions && totalActions > 0 {
		taskSuccess = false
	}

	if !taskSuccess && skippedActions == totalActions && totalActions > 0 {
		e.logger.LogWithCategory(logger.ERROR, "execution", "Task failed: %s, all actions skipped", task.Slug)
	} else {
		e.logger.LogTaskComplete(task.Slug, taskSuccess)
	}

	if !taskSuccess {
		return executedSlugs, fmt.Errorf("task %s failed", task.Slug)
	}

	return executedSlugs, nil
}

// ActionSkippedError represents an action that was skipped due to platform mismatch
type ActionSkippedError struct {
	ActionSlug string
	Reason     string
}

func (e ActionSkippedError) Error() string {
	return fmt.Sprintf("action %s skipped: %s", e.ActionSlug, e.Reason)
}

// executeAction runs a single action, enforcing platform, privilege, and approval checks.
func (e *PlaybookExecutor) executeAction(parentCtx context.Context, action *models.Action, task *models.Task, variables map[string]interface{}) error {
	startTime := time.Now()

	select {
	case <-parentCtx.Done():
		return fmt.Errorf("action cancelled: %v", parentCtx.Err())
	default:
	}

	if !e.isPlatformSupported(action.SupportedOS) {
		reason := fmt.Sprintf("Platform %s not supported on %s", action.SupportedOS, runtime.GOOS)
		e.logger.LogActionSkipped(action.Slug, reason)
		return ActionSkippedError{
			ActionSlug: action.Slug,
			Reason:     reason,
		}
	}

	// Human-in-the-loop approval check
	if action.ApprovalRequired {
		if err := e.requestHumanApproval(action); err != nil {
			return err
		}
	}

	if action.RequiresAdmin {
		if !e.isAdmin() {
			e.logger.LogWithCategory(logger.ERROR, "execution", "Action unauthorized: %s, requires root", action.Slug)
			return fmt.Errorf("unauthorized: root privileges required for action %s", action.Slug)
		}
		e.logger.LogWithCategory(logger.DEBUG, "execution", "Root check passed: %s", action.Slug)
	}

	command, err := e.resolveVariables(action.Command, variables)
	if err != nil {
		e.logger.LogWithCategory(logger.ERROR, "security", "Command injection blocked for action %s: %v", action.Slug, err)
		return fmt.Errorf("variable resolution failed for action %s: %w", action.Slug, err)
	}

	e.logger.LogActionStart(action.Slug, action.Name, command)

	output, err := e.runCommand(parentCtx, command, action.WorkingDir, action.Environment, action.Timeout)
	duration := time.Since(startTime)

	if errors.Is(parentCtx.Err(), context.Canceled) {
		return fmt.Errorf("action cancelled")
	}

	success := err == nil

	e.logger.LogActionComplete(action.Slug, success, output, duration)

	// Store action output for execution summary
	actionOutput := models.ActionOutput{
		TaskSlug:   task.Slug,
		ActionSlug: action.Slug,
		Command:    command,
		Output:     output,
		Success:    success,
		Duration:   duration.Milliseconds(),
		Timestamp:  time.Now().Format(time.RFC3339),
	}
	e.actionOutputs = append(e.actionOutputs, actionOutput)

	if !success {
		return fmt.Errorf("action %s failed: %v", action.Slug, err)
	}

	return nil
}

func (e *PlaybookExecutor) GetActionOutputs() []models.ActionOutput {
	return e.actionOutputs
}

// executeRollback attempts to rollback a failed task
func (e *PlaybookExecutor) executeRollback(parentCtx context.Context, task *models.Task, variables map[string]interface{}, executedSlugs map[string]bool) {
	e.logger.LogWithCategory(logger.INFO, "execution", "Starting rollback: %s", task.Slug)

	ctx, cancel := context.WithTimeout(parentCtx, 10*time.Minute)
	defer cancel()

	for i := len(task.Actions) - 1; i >= 0; i-- {
		action := task.Actions[i]

		if action.RollbackCmd != "" && executedSlugs[action.Slug] {
			e.logger.LogWithCategory(logger.INFO, "execution", "Action rollback started: %s", action.Slug)

			rollbackCmd, err := e.resolveVariables(action.RollbackCmd, variables)
			if err != nil {
				e.logger.LogWithCategory(logger.ERROR, "execution", "Rollback variable resolution failed for %s: %v", action.Slug, err)
				continue
			}

			rollbackTimeout := action.RollbackTimeout
			if rollbackTimeout <= 0 {
				rollbackTimeout = 60
			}

			output, err := e.runCommand(ctx, rollbackCmd, action.WorkingDir, action.Environment, rollbackTimeout)
			if err != nil {
				e.logger.LogWithCategory(logger.ERROR, "execution", "Action rollback failed: %s, err: %v", action.Slug, err)
			} else {
				e.logger.LogWithCategory(logger.INFO, "execution", "Action rollback completed: %s", action.Slug)
				if output != "" {
					e.logger.LogWithCategory(logger.DEBUG, "execution", "Rollback output: %s", strings.TrimSpace(output))
				}
			}
		}
	}
}

// runCommand executes a system command with timeout and cleanup
func (e *PlaybookExecutor) runCommand(parentCtx context.Context, command, workingDir string, env map[string]string, actionTimeout int) (string, error) {
	if command == "" {
		return "", fmt.Errorf("empty command")
	}

	actionTimeoutDuration := time.Duration(actionTimeout) * time.Second
	if actionTimeout <= 0 {
		actionTimeoutDuration = 5 * time.Minute
	}

	// Use the shorter of action timeout or remaining parent deadline.
	var ctx context.Context
	var cancel context.CancelFunc

	if deadline, hasDeadline := parentCtx.Deadline(); hasDeadline {
		parentRemaining := time.Until(deadline)
		if parentRemaining <= 0 {
			return "", fmt.Errorf("parent context already expired")
		}

		effectiveTimeout := actionTimeoutDuration
		if parentRemaining < actionTimeoutDuration {
			effectiveTimeout = parentRemaining
			e.logger.LogWithCategory(logger.TRACE, "execution", "Using parent timeout: %v, action timeout: %v", effectiveTimeout, actionTimeoutDuration)
		} else {
			e.logger.LogWithCategory(logger.TRACE, "execution", "Using action timeout: %v, parent remaining: %v", effectiveTimeout, parentRemaining)
		}

		ctx, cancel = context.WithTimeout(parentCtx, effectiveTimeout)
	} else {
		e.logger.LogWithCategory(logger.TRACE, "execution", "Action timeout: %v", actionTimeoutDuration)
		ctx, cancel = context.WithTimeout(parentCtx, actionTimeoutDuration)
	}
	defer cancel()

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		if shellCmd, detected := e.detectAndExtractShell(ctx, command); detected {
			cmd = shellCmd
		} else {
			cmd = newCmdExec(ctx, command)
			e.logger.LogWithCategory(logger.DEBUG, "execution", "Using cmd for command execution")
		}
	default:
		cmd = exec.Command("sh", "-c", command)
	}

	if workingDir != "" {
		cmd.Dir = workingDir
	}

	if len(env) > 0 {
		cmd.Env = os.Environ()
		for key, value := range env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	output, err := runWithTimeout(cmd, ctx)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			if parentCtx.Err() != nil {
				e.logger.LogWithCategory(logger.ERROR, "execution", "Parent context timeout exceeded")
				return output, fmt.Errorf("parent context timeout exceeded: %w", ctx.Err())
			}
			e.logger.LogWithCategory(logger.ERROR, "execution", "Action timeout exceeded")
		} else {
			e.logger.LogWithCategory(logger.DEBUG, "execution", "Command failed: %v", err)
			e.logger.LogWithCategory(logger.DEBUG, "execution", "Command output: %s", strings.TrimSpace(output))
		}
		return output, err
	}

	return output, nil
}

// isPlatformSupported checks if action is supported on current platform
func (e *PlaybookExecutor) isPlatformSupported(platforms []string) bool {
	if len(platforms) == 0 {
		return true
	}

	currentPlatform := runtime.GOOS

	for _, platform := range platforms {
		if platform == "" || platform == "cross-platform" {
			return true
		}

		// Handle platform aliases
		switch currentPlatform {
		case "darwin":
			if platform == "macos" || platform == "darwin" {
				return true
			}
		case "linux":
			if platform == "linux" {
				return true
			}
		case "windows":
			if platform == "windows" {
				return true
			}
		default:
			if platform == currentPlatform {
				return true
			}
		}
	}

	return false
}

func (e *PlaybookExecutor) isAdmin() bool {
	return system.IsAdmin()
}

// dangerousPatterns are shell metacharacter sequences that must not appear in
// variable values substituted into commands. Blocks the main command-injection
// vectors on Unix (sh -c) and Windows (cmd /S /C).
var dangerousPatterns = []string{";", "&&", "||", "`", "$(", "\n", "\r"}

// dangerousPatternsWindows are additional cmd.exe metacharacters blocked on Windows.
var dangerousPatternsWindows = []string{"&", "|", "<", ">", "^"}

// validateVariableValue rejects values containing shell injection sequences.
// On Windows, also rejects bare cmd.exe metacharacters.
func validateVariableValue(key, value string) error {
	patterns := dangerousPatterns
	if runtime.GOOS == "windows" {
		patterns = append(patterns, dangerousPatternsWindows...)
	}
	for _, pattern := range patterns {
		if strings.Contains(value, pattern) {
			return fmt.Errorf("variable %q contains unsafe sequence %q", key, pattern)
		}
	}
	return nil
}

// resolveVariables replaces ${var} and {{var}} placeholders with validated values.
// Returns an error if any variable value contains shell injection sequences.
func (e *PlaybookExecutor) resolveVariables(command string, variables map[string]interface{}) (string, error) {
	result := command

	for key, value := range variables {
		placeholder1 := fmt.Sprintf("${%s}", key)
		placeholder2 := fmt.Sprintf("{{%s}}", key)
		replacement := fmt.Sprintf("%v", value)

		if err := validateVariableValue(key, replacement); err != nil {
			return "", err
		}

		result = strings.ReplaceAll(result, placeholder1, replacement)
		result = strings.ReplaceAll(result, placeholder2, replacement)
	}

	return result, nil
}
