package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/jadedragon942/ddao/orm"
)

func main() {
	cli := humacli.New(func(hooks humacli.Hooks, config *Config) {
		// Create storage factory and get storage engine
		factory := &StorageFactory{}
		storage, err := factory.CreateStorage(StorageEngine(config.StorageEngine), config.ConnectionURL)
		if err != nil {
			log.Fatalf("Failed to create storage: %v", err)
		}

		// Connect to storage
		ctx := context.Background()
		err = factory.ConnectStorage(ctx, storage, config.ConnectionURL)
		if err != nil {
			log.Fatalf("Failed to connect to storage: %v", err)
		}
		defer storage.ResetConnection(ctx)

		// Create schema and initialize tables
		schema := createExampleSchema()
		err = storage.CreateTables(ctx, schema)
		if err != nil {
			log.Fatalf("Failed to create tables: %v", err)
		}

		// Create ORM
		ormInstance := orm.New(schema).WithStorage(storage)

		// Create service
		service := NewRestdService(ormInstance)

		// Create router with middleware
		router := chi.NewMux()
		router.Use(middleware.RequestID)
		router.Use(middleware.RealIP)
		router.Use(middleware.Logger)
		router.Use(middleware.Recoverer)

		// Create Huma API
		api := humachi.New(router, huma.DefaultConfig("DDAO REST API", "1.0.0"))
		api.OpenAPI().Info.Description = "A RESTful microservice using the Huma framework with pluggable DDAO storage engines"

		// Register routes
		registerRoutes(api, service)

		// Add a health check endpoint
		huma.Get(api, "/health", func(ctx context.Context, input *struct{}) (*struct {
			Body struct {
				Status  string `json:"status" example:"ok" doc:"Health status"`
				Storage string `json:"storage" example:"sqlite" doc:"Current storage engine"`
			}
		}, error) {
			resp := &struct {
				Body struct {
					Status  string `json:"status" example:"ok" doc:"Health status"`
					Storage string `json:"storage" example:"sqlite" doc:"Current storage engine"`
				}
			}{}
			resp.Body.Status = "ok"
			resp.Body.Storage = config.StorageEngine
			return resp, nil
		})

		// Start server
		server := &http.Server{
			Addr:    fmt.Sprintf(":%d", config.Port),
			Handler: router,
		}

		hooks.OnStart(func() {
			log.Printf("Starting DDAO REST API server on port %d with %s storage", config.Port, config.StorageEngine)
			log.Printf("API documentation available at: http://localhost:%d/docs", config.Port)
			log.Printf("OpenAPI spec available at: http://localhost:%d/openapi.json", config.Port)
		})

		hooks.OnStop(func() {
			log.Println("Stopping server...")
			storage.ResetConnection(context.Background())
		})

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	})

	cli.Run()
}

func registerRoutes(api huma.API, service *RestdService) {
	// User routes
	huma.Post(api, "/users", service.CreateUser)
	huma.Get(api, "/users/{userId}", service.GetUser)
	huma.Get(api, "/users", service.GetUserByEmail)
	huma.Put(api, "/users/{userId}", service.UpdateUser)
	huma.Delete(api, "/users/{userId}", service.DeleteUser)

	// Post routes
	huma.Post(api, "/posts", service.CreatePost)
	huma.Get(api, "/posts/{postId}", service.GetPost)
	huma.Put(api, "/posts/{postId}", service.UpdatePost)
	huma.Delete(api, "/posts/{postId}", service.DeletePost)
}