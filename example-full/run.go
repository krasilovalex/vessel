package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/krasilovalex/vessel"
)

func main() {
	ctx := context.Background()

	fmt.Println("📦 Packing the whole directory into in-memory tar...")

	builder := vessel.NewBuilder("vessel-dirss-test:latest").
		From("golang:1.25-alpine").
		Workdir("/app").
		CopyDir("example-full/app", ".").
		Run("go build -o server main.go").
		Cmd("./server")

	if err := builder.Build(ctx); err != nil {
		log.Fatalf("❌ Build failed: %v", err)
	}

	fmt.Println("\n🚀 Running the app...")

	app := vessel.NewContainer("vessel-dirss-test:latest").
		WithName("my-dirss-app")

	if err := app.Up(ctx); err != nil {
		log.Fatalf("❌ Run failed: %v", err)
	}

	fmt.Println("Press Ctrl+C to stop and clean up...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n🛑 Cleaning up...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = app.Stop(shutdownCtx)
	_ = app.Remove(shutdownCtx)
	fmt.Println("✨ Done!")
}
