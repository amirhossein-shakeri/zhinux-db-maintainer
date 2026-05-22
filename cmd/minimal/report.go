package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type durationJSON time.Duration

func (duration durationJSON) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(duration).String())
}

func writeDBReport(cfg config, result backupResult) (string, error) {
	reportPath := result.OutputPath + cfg.ReportExt
	logger.Debugf("writing db report: db=%s path=%s", result.Task.Database, reportPath)
	payload, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal db report: %w", err)
	}
	if err := os.WriteFile(reportPath, payload, 0o644); err != nil {
		return "", fmt.Errorf("write db report: %w", err)
	}
	logger.Infof("db report written: db=%s path=%s", result.Task.Database, reportPath)
	return reportPath, nil
}

func writeRunReportIfEnabled(cfg config, summary *runSummary, results []backupResult) error {
	if !cfg.EnableRunReport {
		logger.Infof("run report disabled by config")
		return nil
	}

	timestamp := time.Now().UTC().Format("20060102T150405Z")
	runName := fmt.Sprintf("backup_run_%s%s", timestamp, cfg.ReportExt)
	runPath := filepath.Join(cfg.OutputDir, runName)
	logger.Debugf("writing run report: path=%s", runPath)

	payload := struct {
		Summary runSummary     `json:"summary"`
		Results []backupResult `json:"results"`
	}{
		Summary: *summary,
		Results: results,
	}

	bytes, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal run report: %w", err)
	}
	if err := os.WriteFile(runPath, bytes, 0o644); err != nil {
		return fmt.Errorf("write run report: %w", err)
	}

	summary.RunReportPath = runPath
	logger.Infof("run report written: path=%s", runPath)
	return nil
}
