package main

import (
	"context"
	"fmt"

	"github.com/jadedragon942/ddao/storage"
	"github.com/jadedragon942/ddao/storage/cockroach"
	"github.com/jadedragon942/ddao/storage/postgres"
	"github.com/jadedragon942/ddao/storage/sqlite"
	"github.com/jadedragon942/ddao/storage/tidb"
	"github.com/jadedragon942/ddao/storage/yugabyte"
)

type StorageEngine string

const (
	SQLite     StorageEngine = "sqlite"
	PostgreSQL StorageEngine = "postgres"
	CockroachDB StorageEngine = "cockroach"
	YugabyteDB StorageEngine = "yugabyte"
	TiDB       StorageEngine = "tidb"
)

type Config struct {
	Port           int    `help:"Port to listen on" short:"p" default:"8080"`
	StorageEngine  string `help:"Storage engine to use (sqlite, postgres, cockroach, yugabyte, tidb)" short:"s" default:"sqlite"`
	ConnectionURL  string `help:"Database connection URL" short:"c" default:"restd.db"`
	DatabaseName   string `help:"Database name" short:"d" default:"restd"`
}

type StorageFactory struct{}

func (sf *StorageFactory) CreateStorage(engine StorageEngine, connectionURL string) (storage.Storage, error) {
	switch engine {
	case SQLite:
		return sqlite.New(), nil
	case PostgreSQL:
		return postgres.New(), nil
	case CockroachDB:
		return cockroach.New(), nil
	case YugabyteDB:
		return yugabyte.New(), nil
	case TiDB:
		return tidb.New(), nil
	default:
		return nil, fmt.Errorf("unsupported storage engine: %s", engine)
	}
}

func (sf *StorageFactory) ConnectStorage(ctx context.Context, storage storage.Storage, connectionURL string) error {
	return storage.Connect(ctx, connectionURL)
}