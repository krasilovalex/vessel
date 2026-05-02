package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/krasilovalex/vessel"
	"github.com/pterm/pterm"
)

func main() {
	ctx := context.Background()

	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgCyan)).WithTextStyle(pterm.NewStyle(pterm.FgBlack)).Println("VESSEL ORCHESTRATOR")

	pterm.Info.Println("Stage 1: Building local backend image...")

	// 1. Собираем наш Go-код в образ "my-backend:latest"
	builder := vessel.NewBuilder("my-backend:latest").
		From("golang:1.25-alpine AS builder").
		Workdir("/app").
		CopyDir("example-full/app", ".").
		Run("go build -o server main.go").
		From("alpine:latest").
		Workdir("/root/").
		CopyFrom("builder", "/app/server", ".").
		Cmd("./server")

	if err := builder.Build(ctx); err != nil {
		pterm.Fatal.Printfln("Build failed: %v", err)
	}

	pterm.Info.Println("Stage 2: Preparing the Fleet...")

	// 2. Описываем наш бэкенд (используем только что собранный образ)
	backend := vessel.NewContainer("my-backend:latest").
		WithName("api-servers").
		WithPort("8080", "8080")

	// 3. Описываем кэш/БД
	redis := vessel.NewContainer("redis:7-alpine").
		WithName("api-caches").
		WithPort("6379", "6379")

	// 4. Объединяем их во Флот
	fleet := vessel.NewFleet(backend, redis)

	pterm.Info.Println("Stage 3: Launching environment concurrently...")

	// 5. Запускаем всё параллельно!
	if err := fleet.Up(ctx); err != nil {
		pterm.Fatal.Printfln("Environment startup failed: %v", err)
	}

	pterm.DefaultBox.WithTitle("Environment Ready").Println(
		"🟢 API Server: http://localhost:8080\n" +
			"🔴 Redis Cache: localhost:6379\n\n" +
			"Press Ctrl+C to gracefully shutdown.",
	)

	// Ожидаем сигнал
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	pterm.Println() // Пустая строка после ^C
	pterm.Warning.Println("Received shutdown signal. Dismantling the fleet...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 6. Параллельно всё тушим и удаляем
	if err := fleet.Down(shutdownCtx); err != nil {
		pterm.Error.Printfln("Error during teardown: %v", err)
	}

	pterm.Success.Println("All systems offline. Great job, Captain! 🫡")
}
