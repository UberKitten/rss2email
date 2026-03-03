package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	content := `
smtp:
  host: mail.example.com
  port: 465
  username: user@example.com
  password: secret123
from: sender@example.com
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadFrom(cfgPath)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	if cfg.SMTP.Host != "mail.example.com" {
		t.Errorf("expected host mail.example.com, got %s", cfg.SMTP.Host)
	}
	if cfg.SMTP.Port != 465 {
		t.Errorf("expected port 465, got %d", cfg.SMTP.Port)
	}
	if cfg.SMTP.Username != "user@example.com" {
		t.Errorf("expected username user@example.com, got %s", cfg.SMTP.Username)
	}
	if cfg.SMTP.Password != "secret123" {
		t.Errorf("expected password secret123, got %s", cfg.SMTP.Password)
	}
	if cfg.From != "sender@example.com" {
		t.Errorf("expected from sender@example.com, got %s", cfg.From)
	}
}

func TestEnvFallback(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	// No file — env vars should be used
	t.Setenv("SMTP_HOST", "env-host.example.com")
	t.Setenv("SMTP_PORT", "2525")
	t.Setenv("SMTP_USERNAME", "envuser")
	t.Setenv("SMTP_PASSWORD", "envpass")
	t.Setenv("FROM", "env@example.com")

	cfg, err := LoadFrom(cfgPath)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	if cfg.SMTP.Host != "env-host.example.com" {
		t.Errorf("expected env host, got %s", cfg.SMTP.Host)
	}
	if cfg.SMTP.Port != 2525 {
		t.Errorf("expected port 2525, got %d", cfg.SMTP.Port)
	}
	if cfg.SMTP.Username != "envuser" {
		t.Errorf("expected envuser, got %s", cfg.SMTP.Username)
	}
	if cfg.From != "env@example.com" {
		t.Errorf("expected env@example.com, got %s", cfg.From)
	}
}

func TestFileOverridesEnv(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	content := `
smtp:
  host: file-host.example.com
  username: fileuser
  password: filepass
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	t.Setenv("SMTP_HOST", "env-host.example.com")
	t.Setenv("SMTP_USERNAME", "envuser")
	t.Setenv("SMTP_PASSWORD", "envpass")

	cfg, err := LoadFrom(cfgPath)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	// File values should win
	if cfg.SMTP.Host != "file-host.example.com" {
		t.Errorf("expected file host to override env, got %s", cfg.SMTP.Host)
	}
	if cfg.SMTP.Username != "fileuser" {
		t.Errorf("expected file username to override env, got %s", cfg.SMTP.Username)
	}
}

func TestHasSMTP(t *testing.T) {
	cfg := &Config{}
	if cfg.HasSMTP() {
		t.Error("empty config should not have SMTP")
	}

	cfg.SMTP.Host = "mail.example.com"
	cfg.SMTP.Username = "user"
	cfg.SMTP.Password = "pass"
	if !cfg.HasSMTP() {
		t.Error("config with host/user/pass should have SMTP")
	}
}

func TestValidate(t *testing.T) {
	cfg := &Config{}
	issues := cfg.Validate()
	if len(issues) < 3 {
		t.Errorf("empty config should have at least 3 issues, got %d", len(issues))
	}

	cfg.SMTP.Host = "mail.example.com"
	cfg.SMTP.Username = "user"
	cfg.SMTP.Password = "pass"
	cfg.SMTP.Port = 587
	issues = cfg.Validate()
	if len(issues) != 0 {
		t.Errorf("valid config should have 0 issues, got %d: %v", len(issues), issues)
	}
}

func TestDefaultPort(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	// No file, no env — should default to 587
	t.Setenv("SMTP_PORT", "")

	cfg, err := LoadFrom(cfgPath)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}
	if cfg.SMTP.Port != 587 {
		t.Errorf("expected default port 587, got %d", cfg.SMTP.Port)
	}
}

func TestInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	if err := os.WriteFile(cfgPath, []byte("{{invalid yaml"), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := LoadFrom(cfgPath)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}
