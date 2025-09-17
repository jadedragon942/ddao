package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jadedragon942/ddao/orm"
	"github.com/jadedragon942/ddao/storage/sqlite"
)

func main() {
	var (
		port       = flag.Int("port", 389, "LDAP server port")
		dbPath     = flag.String("db", "ldap.db", "Database file path")
		bindDN     = flag.String("bind-dn", "cn=admin,dc=example,dc=com", "Bind DN for admin user")
		bindPW     = flag.String("bind-pw", "admin", "Bind password for admin user")
		baseDN     = flag.String("base-dn", "dc=example,dc=com", "Base DN for LDAP tree")
		verbose    = flag.Bool("v", false, "Verbose logging")
	)
	flag.Parse()

	ctx := context.Background()

	// Create schema
	schema := createLDAPSchema()

	// Initialize storage
	storage := sqlite.New()
	err := storage.Connect(ctx, *dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer storage.ResetConnection(ctx)

	// Create tables
	err = storage.CreateTables(ctx, schema)
	if err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	// Create ORM
	ormInstance := orm.New(schema).WithStorage(storage)

	// Initialize LDAP server
	server := NewLDAPServer(*port, *baseDN, *bindDN, *bindPW, ormInstance, *verbose)

	// Setup initial data
	err = setupInitialData(ctx, ormInstance, *baseDN, *bindDN, *bindPW)
	if err != nil {
		log.Fatalf("Failed to setup initial data: %v", err)
	}

	// Start server
	go func() {
		log.Printf("Starting LDAP server on port %d", *port)
		log.Printf("Base DN: %s", *baseDN)
		log.Printf("Admin DN: %s", *bindDN)
		if err := server.Start(); err != nil {
			log.Fatalf("Failed to start LDAP server: %v", err)
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Shutting down LDAP server...")
	server.Stop()
	fmt.Println("Server stopped")
}