// Package config provides application-level configuration for rss2email.
//
// Configuration is loaded from a YAML file (default: ~/.rss2email/config.yaml)
// with environment variable fallbacks for backward compatibility.
//
// Priority order (highest wins):
//  1. Config file values
//  2. Environment variables
//  3. Defaults
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/skx/rss2email/state"
	"gopkg.in/yaml.v3"
)

// SMTPConfig holds SMTP connection settings.
type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Config holds the top-level application configuration.
type Config struct {
	// SMTP holds the SMTP delivery configuration.
	SMTP SMTPConfig `yaml:"smtp"`

	// From is the default sender address.
	From string `yaml:"from"`
}

// path is the resolved config file path, stored after Load.
var path string

// Path returns the path to the config file.
func Path() string {
	if path != "" {
		return path
	}
	return filepath.Join(state.Directory(), "config.yaml")
}

// Load reads configuration from the YAML file and environment variables.
//
// File values take precedence over environment variables. If the config
// file does not exist, only environment variables are used (no error).
func Load() (*Config, error) {
	cfg := &Config{}

	// Try to read the config file
	configPath := Path()
	data, err := os.ReadFile(configPath)
	if err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", configPath, err)
		}
	}

	// Fill in blanks from environment variables
	cfg.applyEnvDefaults()

	return cfg, nil
}

// LoadFrom reads configuration from the specified path.
func LoadFrom(filePath string) (*Config, error) {
	path = filePath
	defer func() { path = "" }()

	cfg := &Config{}

	data, err := os.ReadFile(filePath)
	if err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", filePath, err)
		}
	}

	cfg.applyEnvDefaults()
	return cfg, nil
}

// applyEnvDefaults fills any unset fields from environment variables.
func (c *Config) applyEnvDefaults() {
	if c.SMTP.Host == "" {
		c.SMTP.Host = os.Getenv("SMTP_HOST")
	}
	if c.SMTP.Port == 0 {
		if p := os.Getenv("SMTP_PORT"); p != "" {
			if n, err := strconv.Atoi(p); err == nil {
				c.SMTP.Port = n
			}
		}
		// Default port
		if c.SMTP.Port == 0 {
			c.SMTP.Port = 587
		}
	}
	if c.SMTP.Username == "" {
		c.SMTP.Username = os.Getenv("SMTP_USERNAME")
	}
	if c.SMTP.Password == "" {
		c.SMTP.Password = os.Getenv("SMTP_PASSWORD")
	}
	if c.From == "" {
		c.From = os.Getenv("FROM")
	}
}

// HasSMTP returns true if enough SMTP configuration is present to
// attempt direct SMTP delivery.
func (c *Config) HasSMTP() bool {
	return c.SMTP.Host != "" && c.SMTP.Username != "" && c.SMTP.Password != ""
}

// Validate checks the configuration for obvious problems.
func (c *Config) Validate() []string {
	var issues []string

	if c.SMTP.Host == "" {
		issues = append(issues, "smtp.host is not configured (set in config.yaml or SMTP_HOST env)")
	}
	if c.SMTP.Username == "" {
		issues = append(issues, "smtp.username is not configured (set in config.yaml or SMTP_USERNAME env)")
	}
	if c.SMTP.Password == "" {
		issues = append(issues, "smtp.password is not configured (set in config.yaml or SMTP_PASSWORD env)")
	}
	if c.SMTP.Port < 1 || c.SMTP.Port > 65535 {
		issues = append(issues, fmt.Sprintf("smtp.port %d is invalid (must be 1-65535)", c.SMTP.Port))
	}

	return issues
}
