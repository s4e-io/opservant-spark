package processor

import (
	"time"

	"github.com/s4e-io/opservant-spark/internal/config"
	"github.com/s4e-io/opservant-spark/internal/logger"
	"github.com/s4e-io/opservant-spark/internal/models"
)

type Enricher struct {
	logger *logger.Log
}

func NewEnricher(_ *config.Config, log *logger.Log) *Enricher {
	return &Enricher{
		logger: log,
	}
}

func (e *Enricher) Enrich(playbook *models.Playbook) (*models.Playbook, error) {
	e.logger.Trace("Starting enrichment: %s", playbook.Slug)

	startTime := time.Now()

	e.resolveAllVariables(playbook)
	e.cleanupEmptyFields(playbook)

	duration := time.Since(startTime)
	e.logger.Trace("Enrichment completed, duration: %v", duration)

	return playbook, nil
}

func (e *Enricher) resolveAllVariables(playbook *models.Playbook) {
	for i := range playbook.Tasks {
		task := &playbook.Tasks[i]

		if task.Variables == nil {
			task.Variables = make(map[string]interface{})
		}

		for k, v := range playbook.Variables {
			if _, exists := task.Variables[k]; !exists {
				task.Variables[k] = v
			}
		}
	}
}

func (e *Enricher) cleanupEmptyFields(playbook *models.Playbook) {
	if len(playbook.TargetTags) == 0 {
		playbook.TargetTags = nil
	}

	if len(playbook.Notes) == 0 {
		playbook.Notes = nil
	}

	for i := range playbook.Tasks {
		task := &playbook.Tasks[i]

		if len(task.DependsOn) == 0 {
			task.DependsOn = nil
		}

		if len(task.SupportedOS) == 0 {
			task.SupportedOS = nil
		}
	}
}
