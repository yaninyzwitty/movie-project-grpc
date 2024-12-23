package database

import (
	"fmt"
	"log/slog"
	"time"

	gocqlastra "github.com/datastax/gocql-astra"
	"github.com/gocql/gocql"
)

type AstraConfig struct {
	Path     string
	Username string
	Password string
	Timeout  time.Duration
}

type AstraDb interface {
	CreateDBConn(config *AstraConfig) (*gocql.Session, error)
}

type astradb struct{}

func NewAstraDb() AstraDb {
	return &astradb{}
}

func (a *astradb) CreateDBConn(config *AstraConfig) (*gocql.Session, error) {
	// initialize cluster
	cluster, err := gocqlastra.NewClusterFromBundle(config.Path, config.Username, config.Password, config.Timeout)
	if err != nil {
		slog.Error("Failed to load bundle", "error", err)
		return nil, fmt.Errorf("failed to load bundle: %v", err)
	}

	session, err := gocql.NewSession(*cluster)
	if err != nil {
		slog.Error("Failed to create session", "error", err)
		return nil, fmt.Errorf("failed to create session: %v", err)
	}
	slog.Info("Connected to Astra DB")

	return session, nil
}
