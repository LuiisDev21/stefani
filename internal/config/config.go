// Package config holds user-visible application settings. Secrets never live
// here: they belong to the OS keychain (see internal/secret).
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Defaults are the values a fresh install starts with.
const (
	DefaultBaseURL   = "https://api.openai.com/v1"
	DefaultModel     = "gpt-4o-mini"
	DefaultAccent    = "en-US"
	DefaultTimeout   = 45
	DefaultRetries   = 2
)

// LLM describes the OpenAI-compatible endpoint used for phonetics, translation
// and vocabulary. Any provider that speaks the same protocol works.
type LLM struct {
	BaseURL        string  `json:"baseUrl"`
	Model          string  `json:"model"`
	Temperature    float64 `json:"temperature"`
	TimeoutSeconds int     `json:"timeoutSeconds"`
	MaxRetries     int     `json:"maxRetries"`
	JSONMode       bool    `json:"jsonMode"`
}

// Features gates providers that may be unavailable or experimental.
type Features struct {
	YouTubeMusic bool `json:"youtubeMusic"`
}

// Window stores the last desktop window geometry.
type Window struct {
	Width     int  `json:"width"`
	Height    int  `json:"height"`
	X         int  `json:"x"`
	Y         int  `json:"y"`
	Maximised bool `json:"maximised"`
	Valid     bool `json:"valid"`
}

// Config is the whole persisted settings document.
type Config struct {
	LLM      LLM      `json:"llm"`
	Features Features `json:"features"`
	Window   Window   `json:"window"`
	Accent   string   `json:"accent"`
}

// Default returns a usable configuration.
func Default() Config {
	return Config{
		LLM: LLM{
			BaseURL:        DefaultBaseURL,
			Model:          DefaultModel,
			Temperature:    0.2,
			TimeoutSeconds: DefaultTimeout,
			MaxRetries:     DefaultRetries,
			JSONMode:       true,
		},
		Features: Features{YouTubeMusic: false},
		Accent:   DefaultAccent,
	}
}

// Store reads and writes the config file.
type Store struct {
	path string
}

// NewStore points at the config file, creating parent directories lazily.
func NewStore(path string) *Store { return &Store{path: path} }

// Path returns the config file location.
func (s *Store) Path() string { return s.path }

// Load reads the file, filling defaults for missing fields. A missing file is
// not an error: first run starts from defaults.
func (s *Store) Load() (Config, error) {
	cfg := Default()
	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}
	if len(data) == 0 {
		return cfg, nil
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Default(), fmt.Errorf("config file is not valid JSON: %w", err)
	}
	cfg.normalize()
	return cfg, nil
}

// Save writes the config atomically.
func (s *Store) Save(cfg Config) error {
	cfg.normalize()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("replace config: %w", err)
	}
	return nil
}

func (c *Config) normalize() {
	d := Default()
	c.LLM.BaseURL = strings.TrimRight(strings.TrimSpace(c.LLM.BaseURL), "/")
	if c.LLM.BaseURL == "" {
		c.LLM.BaseURL = d.LLM.BaseURL
	}
	if strings.TrimSpace(c.LLM.Model) == "" {
		c.LLM.Model = d.LLM.Model
	}
	if c.LLM.TimeoutSeconds <= 0 || c.LLM.TimeoutSeconds > 300 {
		c.LLM.TimeoutSeconds = d.LLM.TimeoutSeconds
	}
	if c.LLM.MaxRetries < 0 || c.LLM.MaxRetries > 5 {
		c.LLM.MaxRetries = d.LLM.MaxRetries
	}
	if c.LLM.Temperature < 0 || c.LLM.Temperature > 2 {
		c.LLM.Temperature = d.LLM.Temperature
	}
	if c.Accent == "" {
		c.Accent = d.Accent
	}
}

// Paths resolves the per-user directories STEFANI uses. Branding (name and
// identifier) is centralised in internal/brand, so a rename does not touch the
// storage layout logic.
type Paths struct {
	ConfigDir string
	DataDir   string
	DBPath    string
	DemoDir   string
	Config    string
	Logs      string
}

// ResolvePaths computes paths under the OS config directory, creating the
// directories it returns.
func ResolvePaths(appFolder string) (Paths, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return Paths{}, fmt.Errorf("resolve user config directory: %w", err)
	}
	root := filepath.Join(base, appFolder)
	p := Paths{
		ConfigDir: root,
		DataDir:   filepath.Join(root, "data"),
		DBPath:    filepath.Join(root, "data", "stefani.db"),
		DemoDir:   filepath.Join(root, "data", "demo"),
		Config:    filepath.Join(root, "config.json"),
		Logs:      filepath.Join(root, "logs"),
	}
	for _, dir := range []string{p.ConfigDir, p.DataDir, p.DemoDir, p.Logs} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return Paths{}, fmt.Errorf("create %s: %w", dir, err)
		}
	}
	return p, nil
}
