package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"arrivo-transit-api/internal/cache"
	"arrivo-transit-api/internal/config"
	"arrivo-transit-api/internal/database"
	"arrivo-transit-api/internal/handlers"
	"arrivo-transit-api/internal/ovapi"
	"arrivo-transit-api/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %s", err)
	}

	db, err := database.NewDB(context.Background(), cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("Error connecting to database: %s", err)
	}

	redisClient, err := cache.NewRedis(cfg.RedisDSN)
	if err != nil {
		log.Fatalf("Error connecting to redis: %s", err)
	}

	// Initialize OVapi client and transit service
	ovapiClient := ovapi.NewClient()
	lruCache := cache.NewLRUCache(1000) // Cache up to 1000 items
	transitService := services.NewTransitService(ovapiClient, lruCache, redisClient, db)
	transitHandler := handlers.NewTransitHandler(transitService)

	// Initialize Swagger handler
	apiSpecPath := filepath.Join("docs", "api.yaml")
	swaggerHandler := handlers.NewSwaggerHandler(apiSpecPath)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// API Documentation endpoints
	r.Get("/swagger/*", swaggerHandler.ServeSwaggerUI())
	r.Get("/docs/*", swaggerHandler.ServeStaticDocs("docs"))

	// Transit endpoints
	r.Route("/api/v1", func(r chi.Router) {
			// Swagger spec endpoint
			r.Get("/swagger/doc.json", swaggerHandler.ServeOpenAPISpec())
			r.Get("/swagger/doc.yaml", swaggerHandler.ServeOpenAPISpec())
			
			// Transit API endpoints
			r.Get("/stops/search", transitHandler.SearchStops)
			r.Get("/stops/nearby", transitHandler.GetNearbyStops)
			r.Get("/stops/{stopID}/departures", transitHandler.GetDepartures)
			r.Get("/routes/search", transitHandler.SearchRoutes)
			r.Get("/routes/{routeID}/vehicles", transitHandler.GetVehiclesByRoute)
			r.Get("/vehicles/active", transitHandler.GetAllActiveVehicles)
		})

	fmt.Printf("Starting server on :%s\n", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}