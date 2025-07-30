package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Harshi-itaSinha/target-engine/monitoring"
	"github.com/gorilla/mux"
)

func main() {
	
	router := setupRouter()
    
    metrics := monitoring.NewMetrics()
	go startMetricsServer("9090", metrics)
	

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

func setupRouter() *mux.Router {

	router := mux.NewRouter()
    apiRouter := router.PathPrefix("/v1").Subrouter()
	apiRouter.HandleFunc("/delivery", getCampaigns).Methods("GET")
	apiRouter.HandleFunc("/stats", getStats).Methods("GET")
	router.HandleFunc("/health", healthCheck).Methods("GET")

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


func getCampaigns(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Mock campaigns response"))
}

func getStats(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Mock stats response"))
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
