package processor

import (
	"fmt"
	"strings"

	"github.com/s4e-io/opservant-spark/internal/config"
	"github.com/s4e-io/opservant-spark/internal/logger"
	"github.com/s4e-io/opservant-spark/internal/models"
)

// Validator enforces structural and policy checks for playbooks.
type Validator struct {
	logger *logger.Log
}

// ValidationError describes a validation issue and severity.
type ValidationError struct {
	Field   string
	Message string
	Level   string // error, warning, info
}

func (ve ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", ve.Field, ve.Message)
}

// ValidationResult holds errors produced by validation.
type ValidationResult struct {
	Valid  bool
	Errors []ValidationError
}

func NewValidator(_ *config.Config, log *logger.Log) *Validator {
	return &Validator{
		logger: log,
	}
}

// Validate checks a playbook for structural errors and policy warnings.
func (v *Validator) Validate(playbook *models.Playbook) error {
	v.logger.Trace("Starting validation: %s", playbook.Slug)

	result := &ValidationResult{
		Valid:  true,
		Errors: make([]ValidationError, 0),
	}

	v.validateStructure(playbook, result)

	v.validateSafetyRules(playbook, result)

	v.logValidationResults(result)

	v.logger.Trace("Validation completed: %s", playbook.Slug)

	if !result.Valid {
		errorMessages := make([]string, 0)
		for _, err := range result.Errors {
			if err.Level == "error" {
				errorMessages = append(errorMessages, err.Message)
			}
		}
		if len(errorMessages) > 0 {
			return fmt.Errorf("validation failed: %s", strings.Join(errorMessages, "; "))
		}
	}

	return nil
}

func (v *Validator) validateStructure(playbook *models.Playbook, result *ValidationResult) {
	if playbook.Slug == "" {
		v.addError(result, "playbook.slug", "Playbook slug is required", "error")
	}

	if playbook.Name == "" {
		v.addError(result, "playbook.name", "Playbook name is required", "error")
	}

	if playbook.RiskLevel == "" {
		v.addError(result, "playbook.risk_level", "Risk level is required", "warning")
	} else {
		validRiskLevels := []string{"low", "medium", "high", "critical"}
		if !v.contains(validRiskLevels, playbook.RiskLevel) {
			v.addError(result, "playbook.risk_level",
				fmt.Sprintf("Invalid risk level '%s'. Valid: %v", playbook.RiskLevel, validRiskLevels), "warning")
		}
	}

	if playbook.Timeout <= 0 {
		v.addError(result, "playbook.timeout_seconds", "Playbook timeout must be greater than 0", "warning")
	}

	if len(playbook.Tasks) == 0 {
		v.addError(result, "playbook.tasks", "Playbook must have at least one task", "error")
		return
	}

	for i, task := range playbook.Tasks {
		v.validateTask(&task, i, result)
	}
}

func (v *Validator) validateTask(task *models.Task, index int, result *ValidationResult) {
	// Use slug if available, otherwise fall back to index
	taskPath := fmt.Sprintf("task[%s]", task.Slug)
	if task.Slug == "" {
		taskPath = fmt.Sprintf("task[%d]", index)
	}

	if task.Slug == "" {
		v.addError(result, taskPath+".slug", "Task slug is required", "error")
	}

	if task.Name == "" {
		v.addError(result, taskPath+".name", "Task name is required", "error")
	}

	if task.Timeout <= 0 {
		v.addError(result, taskPath+".timeout_seconds", "Task timeout should be greater than 0", "warning")
	}

	if len(task.Actions) == 0 {
		v.addError(result, taskPath+".actions", "Task must have at least one action", "error")
		return
	}

	for j, action := range task.Actions {
		v.validateAction(&action, task.Slug, j, result)
	}
}

func (v *Validator) validateAction(action *models.Action, taskSlug string, actionIndex int, result *ValidationResult) {
	// Use slugs if available, otherwise fall back to indices
	taskPath := fmt.Sprintf("task[%s]", taskSlug)
	if taskSlug == "" {
		taskPath = "task[unknown]"
	}

	actionPath := fmt.Sprintf("%s.action[%s]", taskPath, action.Slug)
	if action.Slug == "" {
		actionPath = fmt.Sprintf("%s.action[%d]", taskPath, actionIndex)
	}

	if action.Slug == "" {
		v.addError(result, actionPath+".slug", "Action slug is required", "error")
	}

	if action.Name == "" {
		v.addError(result, actionPath+".name", "Action name is required", "error")
	}

	if action.Command == "" {
		v.addError(result, actionPath+".command", "Action command is required", "error")
	}

	if len(action.SupportedOS) == 0 {
		v.addError(result, actionPath+".supported_os", "Action supported_os is required", "warning")
	} else {
		validPlatforms := []string{"windows", "linux", "macos", "darwin", "cross-platform"}
		for _, platform := range action.SupportedOS {
			if !v.contains(validPlatforms, platform) {
				v.addError(result, actionPath+".supported_os",
					fmt.Sprintf("Invalid supported_os '%s'. Valid: %v", platform, validPlatforms), "warning")
			}
		}
	}

	if action.Timeout <= 0 {
		v.addError(result, actionPath+".timeout_seconds", "Action timeout should be greater than 0", "warning")
	}
}

func (v *Validator) validateSafetyRules(playbook *models.Playbook, result *ValidationResult) {
	if playbook.RiskLevel == "critical" && !playbook.HumanInTheLoop {
		v.addError(result, "playbook.human_in_the_loop",
			"Critical risk level should require human approval", "warning")
	}

	if playbook.RiskLevel == "high" || playbook.RiskLevel == "critical" {
		if !playbookHasRollback(playbook) {
			v.addError(result, "playbook.rollback_command",
				"High/Critical risk playbooks should have at least one rollback_command defined", "warning")
		}
	}

	if playbook.AutoRevertOnFailure && !playbookHasRollback(playbook) {
		v.addError(result, "playbook.auto_revert_on_failure",
			"auto_revert_on_failure is true but no rollback_command defined — rollback will never execute", "warning")
	}
}

func playbookHasRollback(playbook *models.Playbook) bool {
	for _, task := range playbook.Tasks {
		for _, action := range task.Actions {
			if action.RollbackCmd != "" {
				return true
			}
		}
	}
	return false
}

func (v *Validator) addError(result *ValidationResult, field, message, level string) {
	ve := ValidationError{
		Field:   field,
		Message: message,
		Level:   level,
	}

	result.Errors = append(result.Errors, ve)

	if level == "error" {
		result.Valid = false
	}
}

func (v *Validator) contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

func (v *Validator) logValidationResults(result *ValidationResult) {
	errorCount := 0
	warningCount := 0
	infoCount := 0

	for _, err := range result.Errors {
		switch err.Level {
		case "error":
			errorCount++
			v.logger.Error("%s", err.Error())
		case "warning":
			warningCount++
			v.logger.Warn("%s", err.Error())
		case "info":
			infoCount++
			v.logger.Info("%s", err.Error())
		}
	}

}
