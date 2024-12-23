package pkg

import (
	"io"
	"log/slog"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   Server `yaml:"server"`
	Database DB     `yaml:"database"`
}

type Server struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type DB struct {
	Path     string `yaml:"path"`
	Username string `yaml:"username"`
}

func (c *Config) LoadFile(file io.Reader) error {
	data, err := io.ReadAll(file)
	if err != nil {
		slog.Error("Failed to read file", "error", err)
		return err
	}

	err = yaml.Unmarshal(data, c)
	if err != nil {
		slog.Error("Failed to unmarshal data", "error", err)
		return err

	}

	return nil
}
