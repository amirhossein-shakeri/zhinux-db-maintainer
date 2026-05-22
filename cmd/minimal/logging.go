package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type appLogger struct {
	verbose bool
	mu      sync.Mutex
}

var logger = &appLogger{}

func (log *appLogger) setVerbose(enabled bool) {
	log.verbose = enabled
}

func (log *appLogger) Infof(format string, args ...any) {
	log.print("INFO", format, args...)
}

func (log *appLogger) Debugf(format string, args ...any) {
	if !log.verbose {
		return
	}
	log.print("DEBUG", format, args...)
}

func (log *appLogger) Warnf(format string, args ...any) {
	log.print("WARN", format, args...)
}

func (log *appLogger) Errorf(format string, args ...any) {
	log.print("ERROR", format, args...)
}

func (log *appLogger) print(level, format string, args ...any) {
	log.mu.Lock()
	defer log.mu.Unlock()
	ts := time.Now().UTC().Format(time.RFC3339)
	fmt.Fprintf(os.Stderr, "%s [%s] %s\n", ts, level, fmt.Sprintf(format, args...))
}

