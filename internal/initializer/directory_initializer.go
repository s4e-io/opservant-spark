package initializer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type directoryInitializer struct{}

func (d *directoryInitializer) Name() string {
	return "Directories"
}

func (d *directoryInitializer) Initialize(_ context.Context, state *InitState) error {
	cfg := state.GetConfig()
	if cfg == nil {
		return fmt.Errorf("config not initialized")
	}

	log := state.GetLogger()
	logDir := cfg.Logging.LogDir

	if logDir == "" || logDir == "." || !filepath.IsAbs(logDir) {
		return nil
	}

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", logDir, err)
	}

	if log != nil {
		log.Trace("Created directory: %s", logDir)
	}

	return nil
}

func (d *directoryInitializer) Cleanup() error {
	return nil
}
