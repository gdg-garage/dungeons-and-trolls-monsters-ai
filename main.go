package main

import (
	"context"
	"fmt"
	"log"
	"os"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
	"github.com/gdg-garage/dungeons-and-trolls-monsters-ai/bot"
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

	if len(os.Args) > 2 && os.Args[2] == "respawn" {
		respawn(ctx, client)
		return
	}

	// Use the client to make API requests
	gameResp, httpResp, err := client.DungeonsAndTrollsApi.DungeonsAndTrollsGame(ctx, nil)
	if err != nil {
		log.Printf("HTTP Response: %+v\n", httpResp)
		log.Fatal(err)
	}
	// fmt.Println("Response:", resp)
	fmt.Println("Running bot ...")
	command := bot.Run(gameResp)
	fmt.Printf("Command: %+v\n", command)

	_, httpResp, err = client.DungeonsAndTrollsApi.DungeonsAndTrollsCommands(ctx, *command)
	if err != nil {
		log.Printf("HTTP Response: %+v\n", httpResp)
		log.Fatal(err)
	}
}

func respawn(ctx context.Context, client *swagger.APIClient) {
	dummyPayload := ctx
	log.Println("Respawning ...")
	_, httpResp, err := client.DungeonsAndTrollsApi.DungeonsAndTrollsRespawn(ctx, dummyPayload)
	if err != nil {
		log.Printf("HTTP Response: %+v\n", httpResp)
		log.Fatal(err)
	}
}
