package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	config := loadConfig()
	store, err := NewStore(config.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to initialize store: %v", err)
	}
	defer store.db.Close()

	router := NewRouter(store, config)
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", config.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Sphere backend starting on port %s", config.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
