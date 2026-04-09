package main

import (
	"context"
	"log"

	"github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/app"
	"github.com/amirhossein-shakeri/zhinux-platform/shutdown"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func main() {
	rootCtx, stopSignal := shutdown.NotifyContext(context.Background())
	defer stopSignal()

	application, err := app.New(rootCtx, app.NormalizeBuildInfo(app.BuildInfo{
		Version:   version,
		Commit:    commit,
		BuildDate: buildDate,
	}))
	if err != nil {
		log.Fatalf("bootstrap failed: %v", err)
	}

	if err := application.Start(); err != nil {
		log.Fatalf("start application: %v", err)
	}

	if err := application.Wait(rootCtx); err != nil {
		log.Printf("runtime error: %v", err)
	}

	if err := application.Shutdown(); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
