package processor

import (
	"github.com/s4e-io/opservant-spark/internal/config"
	"github.com/s4e-io/opservant-spark/internal/logger"
	"github.com/s4e-io/opservant-spark/internal/models"
)

func newTestLogger() *logger.Log {
	return logger.New("error", config.LoggingConfig{})
}

func newTestValidator() *Validator {
	return NewValidator(nil, newTestLogger())
}

func newTestParser() *Parser {
	return NewParser(nil, newTestLogger())
}

func newTestEnricher() *Enricher {
	return NewEnricher(nil, newTestLogger())
}

func validPlaybook() *models.Playbook {
	return &models.Playbook{
		Slug:      "test-playbook",
		Name:      "Test Playbook",
		RiskLevel: "low",
		Timeout:   60,
		Tasks: []models.Task{
			{
				Slug:    "test-task",
				Name:    "Test Task",
				Timeout: 30,
				Actions: []models.Action{
					{
						Slug:        "test-action",
						Name:        "Test Action",
						Command:     "echo hello",
						SupportedOS: []string{"linux"},
						Timeout:     10,
					},
				},
			},
		},
	}
}
