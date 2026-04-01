package main

import (
	"fmt"
	"log"

	"github.com/amirhossein-shakeri/zhinux-platform/config"
	"github.com/amirhossein-shakeri/zhinux-platform/logging"
)

func main() {
	fmt.Println("Hello, World!")

	// Config
	cfg, err := config.LoadBaseFromEnv()
	if err != nil {
		log.Fatalf("Failed to load the config: %v", err)
	}

	logger, err := logging.NewLogger(logging.LoggerOptions{
		Level:       cfg.LogLevel,
		Service:     cfg.ServiceName,
		Development: cfg.Environment == "development",
	})
}
