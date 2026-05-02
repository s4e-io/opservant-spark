package agent

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/s4e-io/opservant-spark/internal/logger"
	"github.com/s4e-io/opservant-spark/internal/models"
)

// requestHumanApproval requests user approval for high-risk actions.
// Returns nil if approved, error if denied.
func (e *PlaybookExecutor) requestHumanApproval(action *models.Action) error {
	fmt.Print("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Print(" HUMAN APPROVAL REQUIRED \n")
	fmt.Print(strings.Repeat("=", 60) + "\n")
	fmt.Printf("Action: %s\n", action.Name)
	fmt.Printf("Description: %s\n", action.Description)
	fmt.Printf("Command: %s\n", action.Command)
	fmt.Printf("Risk Level: %s\n", action.RiskLevel)
	if action.RollbackCmd != "" {
		fmt.Printf("Revertable: [SUCCESS] YES (Rollback: %s)\n", action.RollbackCmd)
	} else {
		fmt.Printf("Revertable: [ERROR] NO\n")
	}
	fmt.Printf("SupportedOS: %s\n", strings.Join(action.SupportedOS, ", "))
	fmt.Printf("Timeout: %d seconds\n", action.Timeout)
	fmt.Print(strings.Repeat("=", 60) + "\n")

	if action.RiskLevel == "high" || action.RiskLevel == "critical" {
		fmt.Printf("[WARNING] This is a %s risk action!\n", strings.ToUpper(action.RiskLevel))
		if action.RequiresAdmin {
			fmt.Printf("[WARNING] WARNING: This action requires administrative privileges!\n")
		}
		fmt.Printf("\n")
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Do you want to proceed? [y/N]: ")

		line, err := reader.ReadString('\n')
		if err != nil {
			e.logger.LogWithCategory(logger.WARN, "execution", "Human approval input error: %v, denying: %s", err, action.Slug)
			return fmt.Errorf("approval input error: %w", err)
		}
		response := strings.ToLower(strings.TrimSpace(line))

		switch response {
		case "y", "yes":
			e.logger.LogWithCategory(logger.INFO, "execution", "Human approval granted: %s", action.Slug)
			return nil
		case "n", "no", "":
			e.logger.LogWithCategory(logger.WARN, "execution", "Human approval denied: %s", action.Slug)
			return fmt.Errorf("action %s denied by user", action.Slug)
		default:
			fmt.Printf("Invalid response. Please enter 'y' for yes or 'n' for no.\n")
		}
	}
}
