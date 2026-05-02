package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("🚀 Starting the full app...")

	// Пытаемся прочитать файл, который лежал рядом в папке
	data, err := os.ReadFile("config.json")
	if err != nil {
		fmt.Printf("❌ Failed to read config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Config loaded successfully: %s\n", string(data))

	// Бесконечный цикл, чтобы контейнер не упал сразу,
	// и мы могли увидеть его в docker ps (если захотим)
	select {}
}
