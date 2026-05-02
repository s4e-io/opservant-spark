package config

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

type Config struct {
	Agent   AgentConfig   `mapstructure:"agent"`
	Logging LoggingConfig `mapstructure:"logging"`
}

// AgentConfig contains agent identity settings.
type AgentConfig struct {
	Name string `mapstructure:"name"`
	UUID string `mapstructure:"uuid"`
}

// LoggingConfig contains all logging settings.
type LoggingConfig struct {
	Level     string `mapstructure:"level"`
	LogToFile bool   `mapstructure:"log_to_file"`
	LogDir    string `mapstructure:"log_dir"`
}

// Load reads the YAML config file at configFile and returns the parsed Config.
// If the file does not exist, built-in defaults are returned.
func Load(configFile string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(configFile)
	setViperDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		if isNotFound(err) {
			return Default(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}

// Default returns a Config populated with built-in defaults.
func Default() *Config {
	return &Config{
		Agent: AgentConfig{},
		Logging: LoggingConfig{
			Level:     "info",
			LogToFile: false,
			LogDir:    "./logs",
		},
	}
}

func setViperDefaults(v *viper.Viper) {
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.log_to_file", false)
	v.SetDefault("logging.log_dir", "./logs")
}

// isNotFound reports whether a viper error indicates the config file was not found.
func isNotFound(err error) bool {
	_, ok := err.(viper.ConfigFileNotFoundError)
	return ok
}

// validLogLevel returns an error if the given level is not one of the accepted values.
func validLogLevel(field, level string) error {
	if level == "" {
		return fmt.Errorf("%s must be set", field)
	}
	for _, v := range []string{"trace", "debug", "info", "warn", "error"} {
		if level == v {
			return nil
		}
	}
	return fmt.Errorf("%s must be one of: [trace debug info warn error], got %q", field, level)
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.Agent.Name == "" {
		return fmt.Errorf("agent.name cannot be empty")
	}
	if _, err := uuid.Parse(c.Agent.UUID); err != nil {
		return fmt.Errorf("agent.uuid is not a valid UUID: %w", err)
	}

	if err := validLogLevel("logging.level", c.Logging.Level); err != nil {
		return err
	}
	if c.Logging.LogToFile && c.Logging.LogDir == "" {
		return fmt.Errorf("logging.log_dir cannot be empty when log_to_file is enabled")
	}

	return nil
}
