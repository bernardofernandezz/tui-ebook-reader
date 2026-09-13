// Package store guarda as preferências (config.json) e o estado de leitura
// (state.json) do tbook, em os.UserConfigDir()/tbook.
package store

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	DefaultTheme = "dark"
	DefaultWidth = 76

	configFile = "config.json"
	stateFile  = "state.json"
)

// Config são as preferências do usuário.
type Config struct {
	Theme      string `json:"theme"`
	Width      int    `json:"width"`
	LibraryDir string `json:"library_dir"`
}

// LoadConfig lê as preferências e completa o que estiver faltando com os padrões.
func LoadConfig() *Config {
	c := &Config{}
	_ = readJSON(path(configFile), c)

	if c.Theme == "" {
		c.Theme = DefaultTheme
	}
	if c.Width == 0 {
		c.Width = DefaultWidth
	}
	if c.LibraryDir == "" {
		c.LibraryDir = defaultLibraryDir()
	}
	return c
}

// Save grava as preferências no disco.
func (c *Config) Save() error {
	return writeJSON(path(configFile), c)
}

func defaultLibraryDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, "Books")
}

func path(name string) string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "tbook", name)
}

func readJSON(file string, v any) error {
	if file == "" {
		return os.ErrNotExist
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func writeJSON(file string, v any) error {
	if file == "" {
		return os.ErrInvalid
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return err
	}
	return os.WriteFile(file, data, 0o644)
}
