package initializer

import (
	"context"
	"fmt"

	"github.com/s4e-io/opservant-spark/internal/logger"
)

type loggerInitializer struct {
	logger *logger.Log
}

func (l *loggerInitializer) Name() string {
	return "Logger"
}

func (l *loggerInitializer) Initialize(ctx context.Context, state *InitState) error {
	cfg := state.GetConfig()
	if cfg == nil {
		return fmt.Errorf("config not initialized")
	}

	log := logger.New(state.LogLevel, cfg.Logging)
	if log == nil {
		return fmt.Errorf("failed to create logger")
	}

	l.logger = log
	state.SetLogger(log)
	return nil
}

func (l *loggerInitializer) Cleanup() error {
	if l.logger != nil {
		l.logger.Close()
	}
	return nil
}
