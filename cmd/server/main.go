package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/yaninyzwitty/movie-project-grpc/internal/database"
	"github.com/yaninyzwitty/movie-project-grpc/internal/database/pkg"
)

var (
	s string
)

func main() {
	var cfg pkg.Config
	file, err := os.Open("config.yaml")
	if err != nil {
		slog.Error("failed to open config.yaml")
		os.Exit(1)
	}
	defer file.Close()
	if err := cfg.LoadFile(file); err != nil {
		slog.Error("failed to load config file", "error", err)
		os.Exit(1)
	}

	if password := os.Getenv("ASTRA_DB_PASSWORD"); password != "" {
		s = password

	}
	_, cancel := context.WithTimeout(context.Background(), cfg.Server.Timeout)
	defer cancel()
	astraConn := database.NewAstraDb()

	config := &database.AstraConfig{
		Path:     cfg.Database.Path,
		Username: cfg.Database.Username,
		Password: s,
		Timeout:  cfg.Server.Timeout,
	}
	session, err := astraConn.CreateDBConn(config)
	if err != nil {
		slog.Error("failed to connect to Astra DB", "error", err)
		os.Exit(1)
	}

	defer session.Close()

}
