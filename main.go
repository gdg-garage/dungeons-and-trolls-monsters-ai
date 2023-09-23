package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	swagger "github.com/gdg-garage/dungeons-and-trolls-go-client"
	botPkg "github.com/gdg-garage/dungeons-and-trolls-monsters-ai/bot"
	"github.com/gdg-garage/dungeons-and-trolls-monsters-ai/prettyprint"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Read command line arguments OR environment variables
	apiKey, found := os.LookupEnv("DNT_API_KEY")
	if !found {
		if len(os.Args) < 2 {
			log.Fatal("USAGE: ./dungeons-and-trolls-monsters-ai API_KEY")
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

	loggerConfig := zap.NewProductionConfig()

	// Set key names and time format for Better Stack
	loggerConfig.EncoderConfig.MessageKey = "message"
	loggerConfig.EncoderConfig.TimeKey = "dt"
	loggerConfig.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder

	// Set log level to Debug
	loggerConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)

	// Set JSON output encoder
	loggerConfig.Encoding = "json"

	// Create the logger
	logger, err := loggerConfig.Build()
	if err != nil {
		log.Fatalf("FATAL: Can't initialize zap logger: %v", err)
	}
	defer logger.Sync()

	if len(os.Args) > 2 && os.Args[2] == "respawn" {
		respawn(ctx, logger.Sugar(), client)
		return
	}

	memory := botPkg.BotMemory{}
	for {
		logger.Info("Fetching game state for NEW TICK ...")
		// Use the client to make API requests
		gameResp, httpResp, err := client.DungeonsAndTrollsApi.DungeonsAndTrollsGame(ctx, nil)
		if err != nil {
			log.Println(err)
			log.Println("HTTP error when fetching game state")
			logger.Error("HTTP error when fetching game state",
				zap.Error(err),
				zap.Any("response", fmt.Sprintf("%+v", httpResp)),
			)
		}
		loggerWTick := logger.Sugar().With(zap.String("tick", gameResp.Tick))
		loggerWTick.Info("============= Game state fetched for NEW TICK =============")
		loggerWTick.Debug("Running bot ...")
		id := "TODO"
		bot := botPkg.New(&gameResp, id, memory, loggerWTick)
		command := bot.Run3()
		prettyprint.Command(loggerWTick, command)

		_, httpResp, err = client.DungeonsAndTrollsApi.DungeonsAndTrollsCommands(ctx, *command)
		if err != nil {
			loggerWTick.Errorw("HTTP error when sending commands",
				zap.Error(err),
				zap.Any("response", fmt.Sprintf("%+v", httpResp)),
			)
		}
		duration := 2 * time.Second
		loggerWTick.Warnw("Sleeping ... TODO: only sleep till end of tick",
			zap.Duration("duration", duration),
		)
		time.Sleep(duration)
	}
}

func respawn(ctx context.Context, logger *zap.SugaredLogger, client *swagger.APIClient) {
	dummyPayload := ctx
	logger.Warn("Respawning ...")
	_, httpResp, err := client.DungeonsAndTrollsApi.DungeonsAndTrollsRespawn(ctx, dummyPayload)
	if err != nil {
		logger.Errorw("HTTP error when respawning",
			zap.Error(err),
			zap.Any("response", fmt.Sprintf("%+v", httpResp)),
		)
	}
}
