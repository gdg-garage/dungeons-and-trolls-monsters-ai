package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
	"github.com/gdg-garage/dungeons-and-trolls-monsters-ai/bot"
	"github.com/gdg-garage/dungeons-and-trolls-monsters-ai/swaggerutil"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func fallbackLog(msg string) {
	noPrefixLog := log.New(os.Stdout, "", 0)
	noPrefixLog.Printf("{'level':'fatal', 'message': '%s'}\n", msg)
}

func main() {
	loggerConfig := zap.NewProductionConfig()

	// Set key names and time format for Better Stack
	loggerConfig.EncoderConfig.MessageKey = "message"
	loggerConfig.EncoderConfig.TimeKey = "dt"
	loggerConfig.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder

	// Set log level to Debug
	loggerConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

	// Set JSON output encoder
	loggerConfig.Encoding = "json"

	// Create the logger
	logger, err := loggerConfig.Build()
	if err != nil {
		fallbackLog(fmt.Sprintf("FATAL: Can't initialize zap logger: %v", err))
		os.Exit(1)
	}
	defer logger.Sync()

	stop, found := os.LookupEnv("DNT_PAUSE_APP")
	if found && stop != "" {
		logger.Error("MONSTER AI IS PAUSED! Unset DNT_PAUSE_APP env variable to unpause.")
		os.Exit(0)
	}

	// Read command line arguments OR environment variables
	apiKey, found := os.LookupEnv("DNT_API_KEY")
	if !found {
		if len(os.Args) < 2 {
			logger.Fatal("USAGE: ./dungeons-and-trolls-monsters-ai API_KEY")
		} else {
			apiKey = os.Args[1]
		}
	}

	// Initialize the HTTP client and set the base URL for the API
	cfg := swagger.NewConfiguration()
	cfg.BasePath = "https://docker.tivvit.cz"

	// Set the X-API-key header value
	ctx := context.WithValue(context.Background(), swagger.ContextAPIKey, swagger.APIKey{Key: apiKey})

	// Create a new client instance
	client := swagger.NewAPIClient(cfg)

	if len(os.Args) > 2 && os.Args[2] == "respawn" {
		respawn(ctx, logger.Sugar(), client)
		return
	}

	botDispatcher := bot.NewBotDispatcher(client, ctx, logger.Sugar())
	backoff := 10 * time.Millisecond
	for {
		logger.Info("Fetching game state for NEW TICK ...")
		// Use the client to make API requests
		gameResp, httpResp, err := client.DungeonsAndTrollsApi.DungeonsAndTrollsGame(ctx, nil)
		if err != nil {
			swaggerutil.LogError(logger.Sugar(), err, httpResp, "Game", nil)
			logger.Info("Sleeping before retrying",
				zap.Duration("duration", backoff),
			)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}
		backoff = 10 * time.Millisecond
		logger.Info("================ Game state fetched for NEW TICK ================")
		err = botDispatcher.HandleTick(&gameResp)
		if err != nil {
			logger.Error("Error when running monster AI",
				zap.Error(err),
			)
			continue
		}
		// prettyprint.Command(loggerWTick, command)

		loggerResponse := logger
		emptyCommand := swagger.DungeonsandtrollsCommandsForMonsters{}
		// Wait until the end of the tick
		_, httpResp, err = client.DungeonsAndTrollsApi.DungeonsAndTrollsMonstersCommands(ctx, emptyCommand, nil)
		swaggerutil.LogResponse(loggerResponse.Sugar(), err, httpResp, "MonstersCommands (empty, blocking)", emptyCommand)
	}
}

func respawn(ctx context.Context, logger *zap.SugaredLogger, client *swagger.APIClient) {
	dummyPayload := ctx
	logger.Warn("Respawning ...")
	_, httpResp, err := client.DungeonsAndTrollsApi.DungeonsAndTrollsRespawn(ctx, dummyPayload, nil)
	swaggerutil.LogResponse(logger, err, httpResp, "Respawn", nil)
}
