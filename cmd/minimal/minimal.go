package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	defaultPgDumpBin = "pg_dump"
	defaultZstdBin   = "zstd"
)

var filenameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

type config struct {
	Host             string
	Port             string
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
}

func main() {
	cfg := parseFlags()

	if err := cfg.validate(); err != nil {
		exitf("invalid config: %v", err)
	}

	if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
		exitf("create output directory: %v", err)
	}

	outputPath := cfg.outputPath()
	if err := runBackup(cfg, outputPath); err != nil {
		exitf("backup failed: %v", err)
	}

	fmt.Printf("backup created: %s\n", outputPath)
}

func parseFlags() config {
	cfg := config{}

	flag.StringVar(&cfg.Host, "host", "", "PostgreSQL host")
	flag.StringVar(&cfg.Port, "port", "5432", "PostgreSQL port")
	flag.StringVar(&cfg.Username, "username", "", "PostgreSQL username")
	flag.StringVar(&cfg.Password, "password", "", "PostgreSQL password")
	flag.StringVar(&cfg.Database, "db", "", "Database name")
	flag.StringVar(&cfg.OutputDir, "out-dir", "backups", "Output directory for backup files")
	flag.StringVar(&cfg.OutputName, "out-name", "", "Optional output filename (defaults to <db>_<host>_<utc>.sql.zst)")
	flag.StringVar(&cfg.PgDumpBin, "pg-dump-bin", defaultPgDumpBin, "Path to pg_dump binary")
	flag.StringVar(&cfg.ZstdBin, "zstd-bin", defaultZstdBin, "Path to zstd binary")
	flag.IntVar(&cfg.ZstdLevel, "zstd-level", 19, "zstd compression level (1-22)")
	flag.BoolVar(&cfg.UseAllCPUs, "zstd-all-cpus", true, "Use all CPU cores for zstd compression")
	flag.IntVar(&cfg.ZstdThreadCount, "zstd-threads", 0, "zstd thread count when not using all CPUs (must be >= 1)")

	flag.Parse()
	return cfg
}

func (cfg config) validate() error {
	if cfg.Host == "" {
		return errors.New("host is required")
	}
	if cfg.Username == "" {
		return errors.New("username is required")
	}
	if cfg.Database == "" {
		return errors.New("db is required")
	}
	if cfg.ZstdLevel < 1 || cfg.ZstdLevel > 22 {
		return errors.New("zstd-level must be between 1 and 22")
	}
	if !cfg.UseAllCPUs && cfg.ZstdThreadCount < 1 {
		return errors.New("zstd-threads must be >= 1 when zstd-all-cpus=false")
	}

	return nil
}

func (cfg config) outputPath() string {
	if cfg.OutputName != "" {
		return filepath.Join(cfg.OutputDir, cfg.OutputName)
	}

	timestamp := time.Now().UTC().Format("20060102T150405Z")
	host := sanitizeForFilename(cfg.Host)
	database := sanitizeForFilename(cfg.Database)
	name := fmt.Sprintf("%s_%s_%s.sql.zst", database, host, timestamp)

	return filepath.Join(cfg.OutputDir, name)
}

func runBackup(cfg config, outputPath string) error {
	pgDumpArgs := []string{
		"-h", cfg.Host,
		"-p", cfg.Port,
		"-U", cfg.Username,
		"-d", cfg.Database,
		"--format=plain",
	}
	pgDumpCmd := exec.Command(cfg.PgDumpBin, pgDumpArgs...)
	pgDumpCmd.Env = append(os.Environ(), "PGPASSWORD="+cfg.Password)

	zstdArgs := []string{fmt.Sprintf("-%d", cfg.ZstdLevel)}
	if cfg.UseAllCPUs {
		zstdArgs = append(zstdArgs, "-T0")
	} else {
		zstdArgs = append(zstdArgs, "-T", fmt.Sprintf("%d", cfg.ZstdThreadCount))
	}
	zstdArgs = append(zstdArgs, "-o", outputPath)
	zstdCmd := exec.Command(cfg.ZstdBin, zstdArgs...)

	pgDumpStdout, err := pgDumpCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("pg_dump stdout pipe: %w", err)
	}
	pgDumpStderr, err := pgDumpCmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("pg_dump stderr pipe: %w", err)
	}

	zstdCmd.Stdin = pgDumpStdout
	zstdStderr, err := zstdCmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("zstd stderr pipe: %w", err)
	}

	if err := pgDumpCmd.Start(); err != nil {
		return fmt.Errorf("start pg_dump: %w", err)
	}
	if err := zstdCmd.Start(); err != nil {
		_ = pgDumpCmd.Process.Kill()
		_ = pgDumpCmd.Wait()
		return fmt.Errorf("start zstd: %w", err)
	}

	pgDumpErrCh := make(chan string, 1)
	zstdErrCh := make(chan string, 1)

	go copyPipeOutput(pgDumpStderr, pgDumpErrCh)
	go copyPipeOutput(zstdStderr, zstdErrCh)

	pgDumpWaitErr := pgDumpCmd.Wait()
	zstdWaitErr := zstdCmd.Wait()

	pgDumpErrText := <-pgDumpErrCh
	zstdErrText := <-zstdErrCh

	if zstdWaitErr != nil {
		return fmt.Errorf("zstd failed: %w; stderr: %s", zstdWaitErr, trimmedOrEmpty(zstdErrText))
	}
	if pgDumpWaitErr != nil {
		return fmt.Errorf("pg_dump failed: %w; stderr: %s", pgDumpWaitErr, trimmedOrEmpty(pgDumpErrText))
	}

	return nil
}

func copyPipeOutput(reader io.Reader, out chan<- string) {
	data, err := io.ReadAll(reader)
	if err != nil {
		out <- fmt.Sprintf("read stderr error: %v", err)
		return
	}
	out <- string(data)
}

func trimmedOrEmpty(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "(empty)"
	}
	return trimmed
}

func sanitizeForFilename(value string) string {
	sanitized := filenameSanitizer.ReplaceAllString(strings.TrimSpace(value), "_")
	sanitized = strings.Trim(sanitized, "._-")
	if sanitized == "" {
		return "unknown"
	}
	return sanitized
}

func exitf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
