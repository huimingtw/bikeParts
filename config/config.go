package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type Settings struct {
	Port                      string `json:"port"`
	EmailUser                 string `json:"email_user"`
	EmailPass                 string `json:"email_pass"`
	EmailTo                   string `json:"email_to"`
	SMTPPort                  string `json:"smtp_port"`
	EmailNotificationsEnabled bool   `json:"email_notifications_enabled"`
}

type AppConfig struct {
	mu       sync.RWMutex
	settings Settings
	path     string // empty = in-memory only (e.g. tests)
}

// Load reads config from the user config directory, falling back to env vars for defaults.
func Load() (*AppConfig, error) {
	path, err := configFilePath()
	if err != nil {
		return nil, err
	}

	cfg := &AppConfig{
		path: path,
		settings: Settings{
			Port:                      envOr("PORT", "8080"),
			SMTPPort:                  envOr("SMTP_PORT", "587"),
			EmailNotificationsEnabled: true,
		},
	}

	// Env vars as baseline
	cfg.settings.EmailUser = os.Getenv("EMAIL_USER")
	cfg.settings.EmailPass = os.Getenv("EMAIL_PASS")
	cfg.settings.EmailTo = os.Getenv("EMAIL_TO")

	// Config file overrides env vars if it exists
	data, err := os.ReadFile(path)
	if err == nil {
		// Pre-populate defaults so missing bool fields stay true rather than zeroing.
		s := Settings{EmailNotificationsEnabled: true}
		if jsonErr := json.Unmarshal(data, &s); jsonErr == nil {
			cfg.settings = s
		}
	}

	return cfg, nil
}

// NewInMemory creates a config with no file persistence (used in tests).
func NewInMemory() *AppConfig {
	return &AppConfig{
		settings: Settings{Port: "8080", SMTPPort: "587", EmailNotificationsEnabled: true},
	}
}

func (c *AppConfig) Get() Settings {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.settings
}

// Set updates the in-memory config and persists to disk (unless in-memory only).
func (c *AppConfig) Set(s Settings) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.settings = s

	if c.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(c.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0o600)
}

// IsConfigured returns true when email settings have been saved.
func (c *AppConfig) IsConfigured() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.settings.EmailUser != ""
}

func configFilePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "bikeparts", "config.json"), nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
