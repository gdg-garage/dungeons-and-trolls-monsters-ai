package main

import (
	"context"
	"fmt"
	"log"
	"os"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
	// swagger "github.com/gdg-garage/dungeons-and-trolls-monsters-ai/swagger"
)

func main() {
	// Read command line arguments
	if len(os.Args) < 2 {
		log.Fatal("USAGE: ./dungeons-and-trolls-monsters-ai API_KEY")
	}
	apiKey := os.Args[1]

	// Initialize the HTTP client and set the base URL for the API
	cfg := swagger.NewConfiguration()
	cfg.BasePath = "https://docker.tivvit.cz"

	// Set the X-API-key header value
	ctx := context.WithValue(context.Background(), swagger.ContextAPIKey, swagger.APIKey{Key: apiKey})

	// Create a new client instance
	client := swagger.NewAPIClient(cfg)

	// Use the client to make API requests
	resp, _, err := client.DungeonsAndTrollsApi.DungeonsAndTrollsGame(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Response:", resp)
}
