package agent

import (
	"context"

	"github.com/s4e-io/opservant-spark/internal/config"
	"github.com/s4e-io/opservant-spark/internal/logger"
	"github.com/s4e-io/opservant-spark/internal/processor"
)

// Agent coordinates playbook processing and execution.
type Agent struct {
	config    *config.Config
	logger    *logger.Log
	executor  *PlaybookExecutor
	processor *processor.Processor
}

func New(cfg *config.Config, log *logger.Log) *Agent {
	executor := NewPlaybookExecutor(log, cfg)
	proc := processor.NewProcessor(cfg, log)

	return &Agent{
		config:    cfg,
		logger:    log,
		executor:  executor,
		processor: proc,
	}
}

// ExecutePlaybook processes the file at playbookPath through the full pipeline and runs it.
func (a *Agent) ExecutePlaybook(ctx context.Context, playbookPath string) error {
	processedPlaybook, err := a.processor.Load(playbookPath)
	if err != nil {
		return err
	}
	return a.executor.Execute(ctx, processedPlaybook)
}
