package main

import (
	"fmt"
	"os"
)

func main() {
	cfg := parseFlags()

	if err := cfg.validate(); err != nil {
		exitf("invalid config: %v", err)
	}

	if err := cfg.prepare(); err != nil {
		exitf("prepare config: %v", err)
	}

	tasks, err := buildTasks(cfg)
	if err != nil {
		exitf("build tasks: %v", err)
	}
	if len(tasks) == 0 {
		exitf("no backup tasks found")
	}

	results, runSummary := runTasks(cfg, tasks)
	if err := writeRunReportIfEnabled(cfg, &runSummary, results); err != nil {
		fmt.Fprintf(os.Stderr, "warn: write run report failed: %v\n", err)
	}

	printRunSummary(runSummary)
	if runSummary.Failed > 0 {
		os.Exit(1)
	}
}
