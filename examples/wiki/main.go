package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jadedragon942/ddao/orm"
	"github.com/jadedragon942/ddao/storage/sqlite"
)

func main() {
	ctx := context.Background()

	schema := createWikiSchema()

	storage := sqlite.New()
	err := storage.Connect(ctx, "wiki.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer storage.ResetConnection(ctx)

	err = storage.CreateTables(ctx, schema)
	if err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	ormInstance := orm.New(schema).WithStorage(storage)

	authService := NewAuthService(ormInstance)
	wikiService := NewWikiService(ormInstance)
	handlers := NewWikiHandlers(authService, wikiService)

	http.HandleFunc("/", handlers.HomeHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)
	http.HandleFunc("/page", handlers.ViewPageHandler)
	http.HandleFunc("/edit", authService.RequireAuth(handlers.EditPageHandler))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	log.Println("Starting wiki server on :8080")
	log.Println("Visit http://localhost:8080 to access the wiki")
	log.Fatal(http.ListenAndServe(":8080", nil))
}