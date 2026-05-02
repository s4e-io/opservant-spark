package agent

import (
	"fmt"

	"github.com/s4e-io/opservant-spark/internal/models"
)

// checkActionDependencies reports whether all declared action dependencies have succeeded.
func (e *PlaybookExecutor) checkActionDependencies(action *models.Action, actionResults map[string]bool) (bool, string) {
	if len(action.DependsOn) == 0 {
		return true, ""
	}

	for _, depSlug := range action.DependsOn {
		result, executed := actionResults[depSlug]
		if !executed {
			return false, fmt.Sprintf("dependency '%s' has not been executed yet", depSlug)
		}
		if !result {
			return false, fmt.Sprintf("dependency '%s' failed", depSlug)
		}
	}

	return true, ""
}

// checkTaskDependencies reports whether all declared task dependencies have succeeded.
func (e *PlaybookExecutor) checkTaskDependencies(task *models.Task, taskResults map[string]bool) (bool, string) {
	if len(task.DependsOn) == 0 {
		return true, ""
	}

	for _, depSlug := range task.DependsOn {
		result, executed := taskResults[depSlug]
		if !executed {
			return false, fmt.Sprintf("dependency '%s' has not been executed yet", depSlug)
		}
		if !result {
			return false, fmt.Sprintf("dependency '%s' failed", depSlug)
		}
	}

	return true, ""
}
