package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/jadedragon942/ddao/schema"
	"github.com/jadedragon942/ddao/storage"
	"github.com/jadedragon942/ddao/storage/cockroach"
	"github.com/jadedragon942/ddao/storage/oracle"
	"github.com/jadedragon942/ddao/storage/postgres"
	"github.com/jadedragon942/ddao/storage/s3"
	"github.com/jadedragon942/ddao/storage/scylla"
	"github.com/jadedragon942/ddao/storage/sqlite"
	"github.com/jadedragon942/ddao/storage/sqlserver"
	"github.com/jadedragon942/ddao/storage/tidb"
	"github.com/jadedragon942/ddao/storage/yugabyte"
)

type Config struct {
	StorageType string
	ConnString  string
	Port        int
}

type AdminServer struct {
	storage    storage.Storage
	config     *Config
	webServer  *WebServer
}

func main() {
	storageType := flag.String("storage", "sqlite", "Storage type (sqlite, postgres, sqlserver, oracle, cockroach, yugabyte, tidb, scylla, s3)")
	connString := flag.String("conn", "file:admin.db", "Connection string for the storage")
	port := flag.Int("port", 8080, "Port to run the web server on")
	flag.Parse()

	config := &Config{
		StorageType: *storageType,
		ConnString:  *connString,
		Port:        *port,
	}

	server, err := NewAdminServer(config)
	if err != nil {
		log.Fatalf("Failed to create admin server: %v", err)
	}

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start admin server: %v", err)
	}
}

func NewAdminServer(config *Config) (*AdminServer, error) {
	stor, err := createStorage(config.StorageType)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	ctx := context.Background()
	if err := stor.Connect(ctx, config.ConnString); err != nil {
		return nil, fmt.Errorf("failed to connect to storage: %w", err)
	}

	// Create a basic schema for demonstration
	testSchema := createTestSchema()
	if err := stor.CreateTables(ctx, testSchema); err != nil {
		log.Printf("Warning: Failed to create test tables: %v", err)
	}

	server := &AdminServer{
		storage: stor,
		config:  config,
	}

	server.webServer = NewWebServer(server, config.Port)

	return server, nil
}

func (s *AdminServer) Start() error {
	log.Printf("Starting Database Administration Tool")
	log.Printf("Storage Type: %s", s.config.StorageType)
	log.Printf("Connection: %s", s.config.ConnString)
	log.Printf("Web Interface: http://localhost:%d", s.config.Port)

	return s.webServer.Start()
}

func createStorage(storageType string) (storage.Storage, error) {
	switch storageType {
	case "sqlite":
		return sqlite.New(), nil
	case "postgres":
		return postgres.New(), nil
	case "sqlserver":
		return sqlserver.New(), nil
	case "oracle":
		return oracle.New(), nil
	case "cockroach":
		return cockroach.New(), nil
	case "yugabyte":
		return yugabyte.New(), nil
	case "tidb":
		return tidb.New(), nil
	case "scylla":
		return scylla.New(), nil
	case "s3":
		return s3.New(), nil
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", storageType)
	}
}

func createTestSchema() *schema.Schema {
	sch := &schema.Schema{
		Tables: make(map[string]*schema.TableSchema),
	}

	// Create a sample users table
	usersTable := &schema.TableSchema{
		TableName:  "users",
		Fields:     make(map[string]schema.ColumnData),
		FieldOrder: []string{"id", "name", "email", "created_at"},
	}

	usersTable.Fields["id"] = schema.ColumnData{
		Name:     "id",
		DataType: "TEXT",
		Nullable: false,
	}
	usersTable.Fields["name"] = schema.ColumnData{
		Name:     "name",
		DataType: "TEXT",
		Nullable: false,
	}
	usersTable.Fields["email"] = schema.ColumnData{
		Name:     "email",
		DataType: "TEXT",
		Nullable: true,
	}
	usersTable.Fields["created_at"] = schema.ColumnData{
		Name:     "created_at",
		DataType: "DATETIME",
		Nullable: true,
	}

	sch.Tables["users"] = usersTable

	// Create a sample products table
	productsTable := &schema.TableSchema{
		TableName:  "products",
		Fields:     make(map[string]schema.ColumnData),
		FieldOrder: []string{"id", "name", "price", "description"},
	}

	productsTable.Fields["id"] = schema.ColumnData{
		Name:     "id",
		DataType: "TEXT",
		Nullable: false,
	}
	productsTable.Fields["name"] = schema.ColumnData{
		Name:     "name",
		DataType: "TEXT",
		Nullable: false,
	}
	productsTable.Fields["price"] = schema.ColumnData{
		Name:     "price",
		DataType: "REAL",
		Nullable: false,
	}
	productsTable.Fields["description"] = schema.ColumnData{
		Name:     "description",
		DataType: "TEXT",
		Nullable: true,
	}

	sch.Tables["products"] = productsTable

	return sch
}