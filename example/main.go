package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/krasilovalex/vessel"
)

func main() {
	ctx := context.Background()

	cwd, _ := os.Getwd()
	pgData := filepath.Join(cwd, "pgdata")

	fmt.Println("Initialize PostgreSQL container")

	pg := vessel.NewContainer("postgres:15-alpine").
		WithName("vessel-pg-test").
		WithPort("5435", "5432").
		WithEnv("POSTGRES_PASWWORD", "supersecret").
		WithVolume(pgData, "/var/lib/postgresql/data")

	if err := pg.Up(ctx); err != nil {
		log.Fatalf("Failed to start Postgres: %v", err)
	}

	fmt.Println("PostgreSQL is running on localhost:5435")
	fmt.Println("Press Ctrl+C to gracefully stop and remove the container...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	fmt.Println("Received shutdown signal. Cleaning up...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := pg.Stop(shutdownCtx); err != nil {
		log.Printf("Warning: error stopping: %v\n", err)
	}

	if err := pg.Remove(shutdownCtx); err != nil {
		log.Printf("Warning: error removing: %v\n", err)
	}

	fmt.Println("Graceful shutdown complete. Bye!")

}
