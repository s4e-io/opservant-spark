package initializer

import (
	"context"
	"fmt"

	"github.com/s4e-io/opservant-spark/internal/config"
)

type configInitializer struct {
	configPath string
	config     *config.Config
}

func (c *configInitializer) Name() string {
	return "Configuration"
}

func (c *configInitializer) Initialize(ctx context.Context, state *InitState) error {
	if c.config != nil {
		state.SetConfig(c.config)
		return nil
	}

	cfg, err := config.Load(c.configPath)
	if err != nil {
		return fmt.Errorf("failed to load config from %s: %w", c.configPath, err)
	}

	state.SetConfig(cfg)
	return nil
}

func (c *configInitializer) Cleanup() error {
	return nil
}
