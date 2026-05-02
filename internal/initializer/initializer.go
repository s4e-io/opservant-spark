package initializer

import (
	"context"
	"fmt"
	"sync"

	"github.com/s4e-io/opservant-spark/internal/config"
	"github.com/s4e-io/opservant-spark/internal/logger"
)

// Initializer defines a component that can be initialized.
type Initializer interface {
	Name() string
	Initialize(ctx context.Context, state *InitState) error
	Cleanup() error
}

// InitState holds shared state between initializers.
type InitState struct {
	Config   *config.Config
	Logger   *logger.Log
	LogLevel string
	mu       sync.RWMutex
}

// GetConfig returns the config (thread-safe).
func (s *InitState) GetConfig() *config.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Config
}

// SetConfig sets the config (thread-safe).
func (s *InitState) SetConfig(cfg *config.Config) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Config = cfg
}

// GetLogger returns the logger (thread-safe).
func (s *InitState) GetLogger() *logger.Log {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Logger
}

// SetLogger sets the logger (thread-safe).
func (s *InitState) SetLogger(log *logger.Log) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Logger = log
}

// Orchestrator manages initialization sequence.
type Orchestrator struct {
	initializers []Initializer
	state        *InitState
	configPath   string
	config       *config.Config
	logLevel     string
}

// OrchestratorOption is a functional option for orchestrator configuration.
type OrchestratorOption func(*Orchestrator)

// WithConfigPath sets the config file path.
func WithConfigPath(path string) OrchestratorOption {
	return func(o *Orchestrator) {
		o.configPath = path
	}
}

// WithConfig provides a preloaded config instead of re-reading from disk.
func WithConfig(cfg *config.Config) OrchestratorOption {
	return func(o *Orchestrator) {
		o.config = cfg
	}
}

// WithLogLevel sets the log level for this run.
func WithLogLevel(level string) OrchestratorOption {
	return func(o *Orchestrator) {
		o.logLevel = level
	}
}

// NewOrchestrator creates a new initializer orchestrator.
func NewOrchestrator(opts ...OrchestratorOption) *Orchestrator {
	o := &Orchestrator{
		state:      &InitState{},
		configPath: "config.yaml",
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

// AddInitializer adds an initializer to the sequence.
func (o *Orchestrator) AddInitializer(init Initializer) {
	o.initializers = append(o.initializers, init)
}

// Initialize runs all initializers in sequence.
func (o *Orchestrator) Initialize(ctx context.Context) error {
	if len(o.initializers) == 0 {
		o.setupDefaultInitializers()
	}

	o.state.LogLevel = o.logLevel

	for _, init := range o.initializers {
		log := o.state.GetLogger()
		if log != nil {
			log.Info("Initializing: %s", init.Name())
		}

		if err := init.Initialize(ctx, o.state); err != nil {
			if log != nil {
				log.Error("Initializer failed: %s, err: %v", init.Name(), err)
			}
			o.Cleanup()
			return fmt.Errorf("initialization failed at %s: %w", init.Name(), err)
		}

		if log != nil {
			log.Info("Successfully initialized: %s", init.Name())
		}
	}

	return nil
}

// setupDefaultInitializers sets up the default initialization sequence.
func (o *Orchestrator) setupDefaultInitializers() {
	o.AddInitializer(&configInitializer{configPath: o.configPath, config: o.config})
	o.AddInitializer(&loggerInitializer{})
	o.AddInitializer(&directoryInitializer{})
}

// GetState returns the current initialization state.
func (o *Orchestrator) GetState() *InitState {
	return o.state
}

// Cleanup cleans up all initialized components.
func (o *Orchestrator) Cleanup() {
	for i := len(o.initializers) - 1; i >= 0; i-- {
		init := o.initializers[i]
		if err := init.Cleanup(); err != nil {
			log := o.state.GetLogger()
			if log != nil {
				log.Warn("Failed to cleanup %s: %v", init.Name(), err)
			}
		}
	}
}
