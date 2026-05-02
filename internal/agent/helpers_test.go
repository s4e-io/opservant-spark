package agent

import (
	"github.com/s4e-io/opservant-spark/internal/config"
	"github.com/s4e-io/opservant-spark/internal/logger"
)

func newTestExecutor() *PlaybookExecutor {
	log := logger.New("error", config.LoggingConfig{})
	return NewPlaybookExecutor(log, &config.Config{})
}
