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
	Connected   bool
}

type AdminServer struct {
	storage    storage.Storage
	config     *Config
	webServer  *WebServer
	connected  bool
}

func main() {
	port := flag.Int("port", 8080, "Port to run the web server on")
	flag.Parse()

	config := &Config{
		Port:      *port,
		Connected: false,
	}

	server := &AdminServer{
		config:    config,
		connected: false,
	}

	server.webServer = NewWebServer(server, config.Port)

	log.Printf("Starting Database Administration Tool")
	log.Printf("Web Interface: http://localhost:%d", config.Port)

	if err := server.webServer.Start(); err != nil {
		log.Fatalf("Failed to start web server: %v", err)
	}
}

func (s *AdminServer) Connect(storageType, connString string) error {
	stor, err := createStorage(storageType)
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}

	ctx := context.Background()
	if err := stor.Connect(ctx, connString); err != nil {
		return fmt.Errorf("failed to connect to storage: %w", err)
	}

	// Create a basic schema for demonstration
	testSchema := createTestSchema()
	if err := stor.CreateTables(ctx, testSchema); err != nil {
		log.Printf("Warning: Failed to create test tables: %v", err)
	}

	s.storage = stor
	s.config.StorageType = storageType
	s.config.ConnString = connString
	s.config.Connected = true
	s.connected = true

	log.Printf("Connected to %s storage: %s", storageType, connString)
	return nil
}

func (s *AdminServer) Disconnect() {
	if s.storage != nil {
		s.storage.ResetConnection(context.Background())
		s.storage = nil
	}
	s.connected = false
	s.config.Connected = false
	s.config.StorageType = ""
	s.config.ConnString = ""
	log.Printf("Disconnected from storage")
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