package main

import (
	"os"
)

func main() {
	cfg := parseFlags()
	if cfg.VerboseShort {
		cfg.Verbose = true
	}
	logger.setVerbose(cfg.Verbose)
	logger.Infof("minimal backup app started")

	if err := cfg.validate(); err != nil {
		exitf("invalid config: %v", err)
	}
	logger.Infof("config validation passed")

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
	logger.Infof("backup tasks prepared: %d", len(tasks))

	results, runSummary := runTasks(cfg, tasks)
	logger.Infof("task execution finished: succeeded=%d failed=%d", runSummary.Succeeded, runSummary.Failed)
	if err := writeRunReportIfEnabled(cfg, &runSummary, results); err != nil {
		logger.Warnf("write run report failed: %v", err)
	}

	printRunSummary(runSummary)
	if runSummary.Failed > 0 {
		logger.Errorf("run completed with failures")
		os.Exit(1)
	}
	logger.Infof("run completed successfully")
}
