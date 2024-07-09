package main

import (
	"log"
	"os"

	// Blank-import the function package so the init() runs
	_ "github_pr_notify"

	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
)

func main() {
	// Use PORT environment variable, or default to 8080.
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	// Log the port the function will be running on
	log.Printf("Listening on port: %s\n", port)

	if err := funcframework.Start(port); err != nil {
		log.Fatalf("funcframework.Start: %v\n", err)
	}
}
