package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := net.DialTimeout("tcp", "api-cache:6379", 2*time.Second)
		if err != nil {
			fmt.Fprintf(w, "❌ Backend is running, but Redis is UNREACHABLE: %v\n", err)
			return
		}
		defer conn.Close()

		fmt.Fprintf(w, "✅ Backend is running!\n🚀 Successfully connected to Redis at api-cache:6379 inside Vessel Network!\n")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server is starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
