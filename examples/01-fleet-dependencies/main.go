package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/krasilovalex/vessel"
	"github.com/pterm/pterm"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pterm.DefaultHeader.Println("Vessel Orchestrator Example")
	pterm.Info.Println("Building and launching infrastructure...")

	// 1. Собираем бэкенд прямо из папки ./app
	builder := vessel.NewBuilder("my-backend:latest").
		From("golang:1.25-alpine AS builder").
		Workdir("/app").
		CopyDir("./app", ".").
		Run("go build -o server main.go").
		From("alpine:latest").
		Workdir("/root/").
		CopyFrom("builder", "/app/server", ".").
		Cmd("./server")

	if err := builder.Build(ctx); err != nil {
		pterm.Fatal.Printf("Failed to build backend image: %v\n", err)
	}

	// 2. Описываем Redis (с сохранением данных)
	redis := vessel.NewContainer("redis:7-alpine").
		WithName("api-cache").
		WithPort("6379", "6379").
		WithVolume("vessel-example-redis", "/data")

	// 3. Описываем Бэкенд (зависит от Redis)
	backend := vessel.NewContainer("my-backend:latest").
		WithName("api-server").
		WithPort("8080", "8080").
		DependsOn(redis) // Магия графа запуска здесь!

	// 4. Формируем Флот и запускаем
	fleet := vessel.NewFleet(redis, backend)
	if err := fleet.Up(ctx); err != nil {
		pterm.Fatal.Printf("Fleet crash: %v\n", err)
	}

	// 5. Выводим инструкции
	pterm.DefaultBox.WithTitle("Environment Ready").Println(
		"🟢 API Server: http://localhost:8080\n" +
			"🔴 Redis Cache: localhost:6379\n\n" +
			"Press Ctrl+C to gracefully shutdown.",
	)

	// 6. Ждем Ctrl+C для аккуратного удаления
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	pterm.Info.Println("Shutting down infrastructure...")
	_ = fleet.Down(context.Background())
}
