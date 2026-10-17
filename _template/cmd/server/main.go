package main

import (
	"log"
	"net/http"
	"os"

	"goph/internal/server"
)

func main() {
	app, err := server.NewApp()
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	addr := os.Getenv("HOST")
	if addr == "" {
		addr = ":8080"
	}

	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, app.Router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
