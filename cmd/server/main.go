package main

import (
	"log"
	"net/http"
	"os"

	"github.com/johanesalxd/brave-search-app/internal/brave"
	"github.com/johanesalxd/brave-search-app/internal/handlers"
)

func main() {
	apiKey := os.Getenv("BRAVE_API_KEY")
	if apiKey == "" {
		log.Fatal("BRAVE_API_KEY environment variable is required")
	}

	braveClient := brave.NewClient(apiKey)
	server, err := handlers.NewServer(braveClient, "templates")
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", server.HandleIndex)
	mux.HandleFunc("/search", server.HandleSearch)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5001"
	}

	log.Printf("server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
