package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type sourceEntry struct {
	Host           any      `json:"host"`
	Port           any      `json:"port"`
	Username       any      `json:"username"`
	Password       any      `json:"password"`
	Database       any      `json:"database"`
	Databases      []string `json:"databases"`
	Disabled       bool     `json:"disabled"`
	Compress       *bool    `json:"compress"`
	ReportDisabled bool     `json:"reportDisabled"`
}

type backupTask struct {
	Host           string
	Port           int
	Username       string
	Password       string
	Database       string
	Compress       bool
	OutputName     string
	DisableReport  bool
	SourceFile     string
	SourceIndex    int
	PrecheckError  string
}

func buildTasks(cfg config) ([]backupTask, error) {
	logger.Infof("building tasks from source-mode=%s", cfg.SourceMode)
	if cfg.SourceMode == "single" {
		logger.Infof("using single source mode with database=%s", cfg.Database)
		return []backupTask{singleTaskFromConfig(cfg)}, nil
	}

	files, err := resolveSourceFiles(cfg.SourceFiles)
	if err != nil {
		return nil, err
	}
	logger.Infof("resolved source files: %d", len(files))
	for _, file := range files {
		logger.Debugf("source file: %s", file)
	}

	var tasks []backupTask
	for _, file := range files {
		parsed, err := parseSourceFile(cfg, file)
		if err != nil {
			logger.Warnf("source parse error: file=%s err=%v", file, err)
			tasks = append(tasks, backupTask{
				Host:          cfg.Host,
				Port:          cfg.Port,
				Username:      cfg.Username,
				Password:      cfg.Password,
				Database:      fmt.Sprintf("invalid-source-%s", sanitizeForFilename(filepath.Base(file))),
				DisableReport: false,
				SourceFile:    file,
				SourceIndex:   -1,
				PrecheckError: err.Error(),
			})
			continue
		}
		logger.Infof("parsed tasks from source file: file=%s tasks=%d", file, len(parsed))
		tasks = append(tasks, parsed...)
	}

	if len(tasks) == 0 {
		return nil, nil
	}
	return tasks, nil
}

func singleTaskFromConfig(cfg config) backupTask {
	return backupTask{
		Host:          defaultIfEmpty(cfg.Host, "localhost"),
		Port:          positiveOrDefault(cfg.Port, 5432),
		Username:      defaultIfEmpty(cfg.Username, "postgres"),
		Password:      cfg.Password,
		Database:      strings.TrimSpace(cfg.Database),
		Compress:      true,
		OutputName:    strings.TrimSpace(cfg.OutputName),
		DisableReport: !cfg.EnableDBReport,
	}
}

func resolveSourceFiles(spec string) ([]string, error) {
	logger.Debugf("resolving source files from spec=%q", spec)
	inputs := splitCSV(spec)
	if len(inputs) == 0 {
		inputs = []string{defaultSourcesGlob}
	}

	seen := map[string]struct{}{}
	files := make([]string, 0)

	for _, in := range inputs {
		matches, err := filepath.Glob(in)
		if err != nil {
			return nil, fmt.Errorf("glob %q: %w", in, err)
		}
		logger.Debugf("source input=%s matches=%d", in, len(matches))
		if len(matches) == 0 {
			if hasJSONExt(in) {
				if _, err := os.Stat(in); err == nil {
					matches = append(matches, in)
				}
			}
		}
		for _, match := range matches {
			if !hasJSONExt(match) {
				continue
			}
			if _, ok := seen[match]; ok {
				continue
			}
			seen[match] = struct{}{}
			files = append(files, match)
		}
	}

	return files, nil
}

func hasJSONExt(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".json" || ext == ".jsonc"
}

func parseSourceFile(cfg config, path string) ([]backupTask, error) {
	logger.Infof("parsing source file: %s", path)
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read source file %s: %w", path, err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".jsonc" {
		raw = stripJSONCComments(raw)
		logger.Debugf("jsonc comments stripped for file=%s", path)
	}

	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		return nil, nil
	}

	if strings.HasPrefix(trimmed, "[") {
		var entries []sourceEntry
		if err := json.Unmarshal(raw, &entries); err != nil {
			return nil, fmt.Errorf("parse json array %s: %w", path, err)
		}
		logger.Debugf("parsed json array entries=%d file=%s", len(entries), path)
		return expandEntries(cfg, entries, path), nil
	}

	var entry sourceEntry
	if err := json.Unmarshal(raw, &entry); err != nil {
		return nil, fmt.Errorf("parse json object %s: %w", path, err)
	}
	logger.Debugf("parsed json object file=%s", path)
	return expandEntries(cfg, []sourceEntry{entry}, path), nil
}

func expandEntries(cfg config, entries []sourceEntry, sourceFile string) []backupTask {
	var tasks []backupTask
	for i, entry := range entries {
		if entry.Disabled {
			logger.Infof("skipping disabled source entry file=%s index=%d", sourceFile, i)
			continue
		}
		host := defaultIfEmpty(anyToString(entry.Host), cfg.Host)
		port := parsePortString(entry.Port, cfg.Port)
		username := defaultIfEmpty(anyToString(entry.Username), cfg.Username)
		password := anyToString(entry.Password)
		if password == "" {
			password = cfg.Password
		}
		disabledReport := entry.ReportDisabled || !cfg.EnableDBReport

		database := strings.TrimSpace(anyToString(entry.Database))
		if database != "" {
			logger.Debugf("adding task file=%s index=%d db=%s host=%s port=%d user=%s", sourceFile, i, database, host, port, username)
			compress := true
			if entry.Compress != nil {
				compress = *entry.Compress
			}
			tasks = append(tasks, backupTask{
				Host:          host,
				Port:          port,
				Username:      username,
				Password:      password,
				Database:      database,
				Compress:      compress,
				DisableReport: disabledReport,
				SourceFile:    sourceFile,
				SourceIndex:   i,
			})
		}

		for _, db := range entry.Databases {
			dbName := strings.TrimSpace(db)
			if dbName == "" {
				continue
			}
			logger.Debugf("adding task file=%s index=%d db=%s host=%s port=%d user=%s", sourceFile, i, dbName, host, port, username)
			compress := true
			if entry.Compress != nil {
				compress = *entry.Compress
			}
			tasks = append(tasks, backupTask{
				Host:          host,
				Port:          port,
				Username:      username,
				Password:      password,
				Database:      dbName,
				Compress:      compress,
				DisableReport: disabledReport,
				SourceFile:    sourceFile,
				SourceIndex:   i,
			})
		}
	}
	return tasks
}

func anyToString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(v)
	case float64:
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%v", v)
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
}

func defaultIfEmpty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func positiveOrDefault(value, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

var jsoncLineCommentPattern = regexp.MustCompile(`(?m)//.*$`)
var jsoncTrailingCommaPattern = regexp.MustCompile(`,(\s*[\]}])`)

func stripJSONCComments(raw []byte) []byte {
	stripped := jsoncLineCommentPattern.ReplaceAllString(string(raw), "")
	stripped = jsoncTrailingCommaPattern.ReplaceAllString(stripped, "$1")
	return []byte(stripped)
}
