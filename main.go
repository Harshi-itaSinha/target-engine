package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Harshi-itaSinha/target-engine/internal/config"
	"github.com/Harshi-itaSinha/target-engine/internal/database.go"
	"github.com/Harshi-itaSinha/target-engine/internal/handler"
	"github.com/Harshi-itaSinha/target-engine/internal/middleware"
	"github.com/Harshi-itaSinha/target-engine/internal/repository"
	"github.com/Harshi-itaSinha/target-engine/internal/service"
	"github.com/Harshi-itaSinha/target-engine/monitoring"
	"github.com/gorilla/mux"
)

func main() {

	cfg := config.LoadConfig()
	// repo := repository.NewMemoryRepository()
	// defer repo.Close()

	// 2. Initialize MongoDB client
	uri := config.GetEnv("MONGO_URI")
	//cfg.Database.ConnectionString
	dbClient, err := database.NewMongoClient(uri)
	if err != nil {
		log.Fatalf("Failed to initialize MongoDB client: %v", err)
	}

	// 3. Get the database
	database := dbClient.Database(cfg.Database.DatabaseName)

	// 4. Initialize repository with MongoDB database and client
	repo := repository.NewRepository(database, dbClient)
	defer func() {
		if err := repo.Close(); err != nil {
			log.Printf("Failed to close repository: %v", err)
		}
	}()
	defer repo.Close()

	targetingService := service.NewTargetingService(repo, cfg)

	deliveryHandler := handler.NewDeliveryHandler(targetingService)

	var metrics *monitoring.Metrics
	if cfg.Metrics.Enabled {
		metrics = monitoring.NewMetrics()
	}

	router := setupRouter(deliveryHandler, cfg, metrics)

	if cfg.Metrics.Enabled {
		go startMetricsServer(cfg.Metrics.Port, metrics)
	}

	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Println("Starting server on port 8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Forced shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

func setupRouter(deliveryHandler *handler.DeliveryHandler, cfg *config.Config, metrics *monitoring.Metrics) *mux.Router {

	router := mux.NewRouter()

	// Apply global middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.CORS)
	router.Use(middleware.Recovery)
	router.Use(middleware.Health)
	router.Use(middleware.Timeout(10 * time.Second))

	if cfg.Metrics.Enabled && metrics != nil {
		router.Use(metrics.MetricsMiddleware)
	}

	apiRouter := router.PathPrefix("/v1").Subrouter()
	apiRouter.HandleFunc("/delivery", deliveryHandler.GetCampaigns).Methods("GET")
	apiRouter.HandleFunc("/stats", deliveryHandler.GetStats).Methods("GET")
	apiRouter.HandleFunc("/target",deliveryHandler.CreateTargetingRule).Methods("POST")
	apiRouter.HandleFunc("/campaign",deliveryHandler.CreateCampaign).Methods("POST")
	router.HandleFunc("/health", deliveryHandler.Health).Methods("GET")

	return router
}

func startMetricsServer(port string, metrics *monitoring.Metrics) {
	metricsRouter := mux.NewRouter()
	metricsRouter.Handle("/metrics", metrics.Handler())

	metricsServer := &http.Server{
		Addr:    ":" + port,
		Handler: metricsRouter,
	}

	log.Printf("Starting metrics server on port %s", port)
	if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("Metrics server error: %v", err)
	}
}
