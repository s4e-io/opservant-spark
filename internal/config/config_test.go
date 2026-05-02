package config

import (
	"os"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Logging.Level != "info" {
		t.Errorf("expected level info, got %s", cfg.Logging.Level)
	}
	if cfg.Logging.LogToFile != false {
		t.Error("expected log_to_file false")
	}
	if cfg.Logging.LogDir != "./logs" {
		t.Errorf("expected log_dir ./logs, got %s", cfg.Logging.LogDir)
	}
	if cfg.Agent.Name != "" {
		t.Errorf("expected empty agent name, got %s", cfg.Agent.Name)
	}
	if cfg.Agent.UUID != "" {
		t.Errorf("expected empty agent uuid, got %s", cfg.Agent.UUID)
	}
}

func TestLoadValidYAML(t *testing.T) {
	f, err := os.CreateTemp("", "spark-config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	content := `
agent:
  name: test-agent
  uuid: 550e8400-e29b-41d4-a716-446655440000
logging:
  level: debug
  log_to_file: false
  log_dir: ./logs
`
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()

	cfg, err := Load(f.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Agent.Name != "test-agent" {
		t.Errorf("expected agent name test-agent, got %s", cfg.Agent.Name)
	}
	if cfg.Agent.UUID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("unexpected uuid: %s", cfg.Agent.UUID)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("expected level debug, got %s", cfg.Logging.Level)
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	f, err := os.CreateTemp("", "spark-config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	if _, err := f.WriteString("{unclosed: [bracket"); err != nil {
		t.Fatal(err)
	}
	f.Close()

	_, err = Load(f.Name())
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func TestValidate_OK(t *testing.T) {
	cfg := &Config{
		Agent:   AgentConfig{Name: "spark", UUID: "550e8400-e29b-41d4-a716-446655440000"},
		Logging: LoggingConfig{Level: "info", LogToFile: false, LogDir: "./logs"},
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_EmptyName(t *testing.T) {
	cfg := &Config{
		Agent:   AgentConfig{Name: "", UUID: "550e8400-e29b-41d4-a716-446655440000"},
		Logging: LoggingConfig{Level: "info"},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for empty agent name")
	}
}

func TestValidate_BadUUID(t *testing.T) {
	cfg := &Config{
		Agent:   AgentConfig{Name: "spark", UUID: "not-a-uuid"},
		Logging: LoggingConfig{Level: "info"},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid UUID")
	}
}

func TestValidate_EmptyUUID(t *testing.T) {
	cfg := &Config{
		Agent:   AgentConfig{Name: "spark", UUID: ""},
		Logging: LoggingConfig{Level: "info"},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for empty UUID")
	}
}

func TestValidate_BadLevel(t *testing.T) {
	cfg := &Config{
		Agent:   AgentConfig{Name: "spark", UUID: "550e8400-e29b-41d4-a716-446655440000"},
		Logging: LoggingConfig{Level: "verbose"},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid log level")
	}
}

func TestValidate_EmptyLevel(t *testing.T) {
	cfg := &Config{
		Agent:   AgentConfig{Name: "spark", UUID: "550e8400-e29b-41d4-a716-446655440000"},
		Logging: LoggingConfig{Level: ""},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for empty log level")
	}
}

func TestValidate_AllValidLevels(t *testing.T) {
	for _, level := range []string{"trace", "debug", "info", "warn", "error"} {
		cfg := &Config{
			Agent:   AgentConfig{Name: "spark", UUID: "550e8400-e29b-41d4-a716-446655440000"},
			Logging: LoggingConfig{Level: level},
		}
		if err := cfg.Validate(); err != nil {
			t.Errorf("level %q should be valid, got: %v", level, err)
		}
	}
}

func TestValidate_LogToFileNoDir(t *testing.T) {
	cfg := &Config{
		Agent:   AgentConfig{Name: "spark", UUID: "550e8400-e29b-41d4-a716-446655440000"},
		Logging: LoggingConfig{Level: "info", LogToFile: true, LogDir: ""},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for log_to_file=true with empty log_dir")
	}
}

func TestValidate_LogToFileWithDir(t *testing.T) {
	cfg := &Config{
		Agent:   AgentConfig{Name: "spark", UUID: "550e8400-e29b-41d4-a716-446655440000"},
		Logging: LoggingConfig{Level: "info", LogToFile: true, LogDir: "/tmp"},
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
