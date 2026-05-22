package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	defaultZstdBin      = "zstd"
	defaultReportExt    = ".report.json"
	defaultSourceMode   = "auto"
	defaultProcessMode  = "sync"
	defaultSourcesGlob  = "db-sources/*"
	defaultBackupSuffix = ".sql.zst"
)

var filenameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

type config struct {
	Host             string
	Port             int
	Username         string
	Password         string
	Database         string
	OutputDir        string
	OutputName       string
	PgDumpBin        string
	ZstdBin          string
	ZstdLevel        int
	UseAllCPUs       bool
	ZstdThreadCount  int
	SourceMode       string
	SourceFiles      string
	ProcessMode      string
	Concurrency      int
	EnableDBReport   bool
	EnableRunReport  bool
	ReportExt        string
	ProfileSummary   bool
	Source           string
	Verbose          bool
	VerboseShort     bool
}

func parseFlags() config {
	cfg := config{}

	flag.StringVar(&cfg.Host, "host", "localhost", "PostgreSQL host")
	flag.IntVar(&cfg.Port, "port", 5432, "PostgreSQL port")
	flag.StringVar(&cfg.Username, "username", "postgres", "PostgreSQL username")
	flag.StringVar(&cfg.Password, "password", "", "PostgreSQL password")
	flag.StringVar(&cfg.Database, "db", "", "Database name (single quick-run mode)")
	flag.StringVar(&cfg.OutputDir, "out-dir", "backups", "Output directory for backup files")
	flag.StringVar(&cfg.OutputName, "out-name", "", "Optional output filename for single-db mode")
	flag.StringVar(&cfg.PgDumpBin, "pg-dump-bin", "", "Path to pg_dump binary (auto-discover when empty)")
	flag.StringVar(&cfg.ZstdBin, "zstd-bin", defaultZstdBin, "Path to zstd binary")
	flag.IntVar(&cfg.ZstdLevel, "zstd-level", 19, "zstd compression level (1-22)")
	flag.BoolVar(&cfg.UseAllCPUs, "zstd-all-cpus", true, "Use all CPU cores for zstd compression")
	flag.IntVar(&cfg.ZstdThreadCount, "zstd-threads", 0, "zstd threads when zstd-all-cpus=false")
	flag.StringVar(&cfg.SourceMode, "source-mode", defaultSourceMode, "Backup source mode: auto|single|files")
	flag.StringVar(&cfg.SourceFiles, "source-files", defaultSourcesGlob, "Comma-separated file list or glob for db source files")
	flag.StringVar(&cfg.Source, "source", "", "Single source file path (equivalent to source-mode=files with one file)")
	flag.StringVar(&cfg.ProcessMode, "process-mode", defaultProcessMode, "Processing mode: sync|async")
	flag.IntVar(&cfg.Concurrency, "concurrency", runtime.NumCPU(), "Max concurrent backups when process-mode=async")
	flag.BoolVar(&cfg.EnableDBReport, "report-db", true, "Write per-database report next to each backup")
	flag.BoolVar(&cfg.EnableRunReport, "report-run", true, "Write aggregate run report in output directory")
	flag.StringVar(&cfg.ReportExt, "report-ext", defaultReportExt, "Report file extension")
	flag.BoolVar(&cfg.ProfileSummary, "profile-summary", true, "Include runtime profile summary in reports")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "Enable verbose debug logs")
	flag.BoolVar(&cfg.VerboseShort, "v", false, "Enable verbose debug logs (short)")

	flag.Parse()
	return cfg
}

func (cfg config) validate() error {
	if cfg.ZstdLevel < 1 || cfg.ZstdLevel > 22 {
		return errors.New("zstd-level must be between 1 and 22")
	}
	if !cfg.UseAllCPUs && cfg.ZstdThreadCount < 1 {
		return errors.New("zstd-threads must be >= 1 when zstd-all-cpus=false")
	}
	if cfg.ProcessMode != "sync" && cfg.ProcessMode != "async" {
		return fmt.Errorf("invalid process-mode %q", cfg.ProcessMode)
	}
	if cfg.Concurrency < 1 {
		return errors.New("concurrency must be >= 1")
	}
	if cfg.ReportExt == "" {
		return errors.New("report-ext cannot be empty")
	}
	if !strings.HasPrefix(cfg.ReportExt, ".") {
		return errors.New("report-ext must start with '.'")
	}
	if cfg.SourceMode != "auto" && cfg.SourceMode != "single" && cfg.SourceMode != "files" {
		return fmt.Errorf("invalid source-mode %q", cfg.SourceMode)
	}
	if cfg.SourceMode == "single" && strings.TrimSpace(cfg.Database) == "" {
		return errors.New("db is required when source-mode=single")
	}

	return nil
}

func (cfg *config) prepare() error {
	if cfg.VerboseShort {
		cfg.Verbose = true
	}
	logger.setVerbose(cfg.Verbose)
	logger.Infof("preparing configuration")
	logger.Debugf("raw config: host=%s port=%d username=%s db=%s source-mode=%s process-mode=%s concurrency=%d out-dir=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Database, cfg.SourceMode, cfg.ProcessMode, cfg.Concurrency, cfg.OutputDir)

	if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	logger.Infof("ensured output directory exists: %s", cfg.OutputDir)

	pgDumpPath, err := resolvePgDumpPath(cfg.PgDumpBin)
	if err != nil {
		return err
	}
	cfg.PgDumpBin = pgDumpPath
	logger.Infof("resolved pg_dump binary: %s", cfg.PgDumpBin)

	zstdPath, err := exec.LookPath(cfg.ZstdBin)
	if err != nil {
		return fmt.Errorf("resolve zstd binary %q: %w", cfg.ZstdBin, err)
	}
	cfg.ZstdBin = zstdPath
	logger.Infof("resolved zstd binary: %s", cfg.ZstdBin)

	if cfg.Source != "" {
		cfg.SourceMode = "files"
		cfg.SourceFiles = cfg.Source
		logger.Infof("single source file provided, switching source-mode=files: %s", cfg.Source)
	}

	if cfg.SourceMode == "auto" {
		if strings.TrimSpace(cfg.Database) != "" {
			cfg.SourceMode = "single"
		} else {
			cfg.SourceMode = "files"
		}
		logger.Infof("resolved auto source-mode to: %s", cfg.SourceMode)
	}

	if cfg.ProcessMode == "sync" {
		cfg.Concurrency = 1
		logger.Infof("process-mode=sync forces concurrency=1")
	}
	logger.Infof("configuration prepared successfully")

	return nil
}

func resolvePgDumpPath(explicit string) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		resolved, err := exec.LookPath(explicit)
		if err != nil {
			return "", fmt.Errorf("resolve pg_dump binary %q: %w", explicit, err)
		}
		return resolved, nil
	}

	candidates := []string{
		"pg_dump",
		"/opt/homebrew/opt/postgresql@18/bin/pg_dump",
		"/opt/homebrew/opt/postgresql@17/bin/pg_dump",
		"/opt/homebrew/opt/postgresql@16/bin/pg_dump",
		"/opt/homebrew/bin/pg_dump",
		"/usr/local/bin/pg_dump",
		"/usr/bin/pg_dump",
	}

	for _, candidate := range candidates {
		resolved, err := exec.LookPath(candidate)
		if err == nil {
			logger.Debugf("pg_dump candidate matched: %s -> %s", candidate, resolved)
			return resolved, nil
		}
		logger.Debugf("pg_dump candidate not found: %s", candidate)
	}

	return "", errors.New("pg_dump binary not found in fallback candidates")
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func parsePortString(portRaw any, defaultPort int) int {
	switch value := portRaw.(type) {
	case nil:
		return defaultPort
	case int:
		if value > 0 {
			return value
		}
	case float64:
		if value > 0 {
			return int(value)
		}
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		if err == nil && parsed > 0 {
			return parsed
		}
	}
	return defaultPort
}

func sanitizeForFilename(value string) string {
	sanitized := filenameSanitizer.ReplaceAllString(strings.TrimSpace(value), "_")
	sanitized = strings.Trim(sanitized, "._-")
	if sanitized == "" {
		return "unknown"
	}
	return sanitized
}

func buildOutputPath(outDir, outputName, host, database string, startedAt time.Time) string {
	if strings.TrimSpace(outputName) != "" {
		return filepath.Join(outDir, outputName)
	}
	timestamp := startedAt.UTC().Format("20060102T150405Z")
	name := fmt.Sprintf("%s_%s_%s%s", sanitizeForFilename(database), sanitizeForFilename(host), timestamp, defaultBackupSuffix)
	return filepath.Join(outDir, name)
}
