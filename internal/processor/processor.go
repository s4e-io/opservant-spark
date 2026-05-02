// Package processor ingests, parses, validates, and transforms playbook files.
package processor

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/s4e-io/opservant-spark/internal/config"
	"github.com/s4e-io/opservant-spark/internal/logger"
	"github.com/s4e-io/opservant-spark/internal/models"
)

// Processor coordinates the playbook processing pipeline.
type Processor struct {
	config    *config.Config
	logger    *logger.Log
	ingester  *Ingester
	parser    *Parser
	validator *Validator
	enricher  *Enricher
}

// NewProcessor builds a processor with the standard pipeline stages.
func NewProcessor(cfg *config.Config, log *logger.Log) *Processor {
	return &Processor{
		config:    cfg,
		logger:    log,
		ingester:  NewIngester(cfg, log),
		parser:    NewParser(cfg, log),
		validator: NewValidator(cfg, log),
		enricher:  NewEnricher(cfg, log),
	}
}

// Load reads and processes a playbook file.
func (p *Processor) Load(inputPath string) (*models.Playbook, error) {
	p.logger.Info("Starting file processing: %s", inputPath)

	startTime := time.Now()

	rawData, err := p.ingester.IngestFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("ingestion failed: %w", err)
	}

	parsedData, err := p.parser.Parse(rawData, inputPath)
	if err != nil {
		return nil, fmt.Errorf("parsing failed: %w", err)
	}

	if err := p.validator.Validate(parsedData); err != nil {
		return nil, err
	}

	enrichedData, err := p.enricher.Enrich(parsedData)
	if err != nil {
		return nil, fmt.Errorf("enrichment failed: %w", err)
	}

	duration := time.Since(startTime)
	p.logProcessSummary(enrichedData, duration)

	return enrichedData, nil
}

func (p *Processor) logProcessSummary(playbook *models.Playbook, duration time.Duration) {
	taskCount := len(playbook.Tasks)
	actionCount := 0

	for _, task := range playbook.Tasks {
		actionCount += len(task.Actions)
	}

	p.logger.Info("Playbook processed: %s, tasks: %d, actions: %d, duration: %v",
		playbook.Slug, taskCount, actionCount, duration)
}

func fileFormat(path string) string {
	return strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
}
