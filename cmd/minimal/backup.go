package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type backupResult struct {
	Task                backupTask   `json:"task"`
	OutputPath          string       `json:"outputPath"`
	StartedAt           time.Time    `json:"startedAt"`
	EndedAt             time.Time    `json:"endedAt"`
	DumpDuration        durationJSON `json:"dumpDuration"`
	CompressionDuration durationJSON `json:"compressionDuration"`
	TotalDuration       durationJSON `json:"totalDuration"`
	UncompressedBytes   int64        `json:"uncompressedBytes"`
	CompressedBytes     int64        `json:"compressedBytes"`
	UncompressedHuman   string       `json:"uncompressedHuman"`
	CompressedHuman     string       `json:"compressedHuman"`
	CompressionRatio    float64      `json:"compressionRatio"`
	CompressionPercent  float64      `json:"compressionPercent"`
	PgDumpBin           string       `json:"pgDumpBin"`
	ZstdBin             string       `json:"zstdBin"`
	ZstdArgs            []string     `json:"zstdArgs,omitempty"`
	PgDumpArgs          []string     `json:"pgDumpArgs"`
	CompressionEnabled  bool         `json:"compressionEnabled"`
	ExecutionPath       string       `json:"executionPath"`
	Success             bool         `json:"success"`
	Error               string       `json:"error,omitempty"`
	PgDumpStderr        string       `json:"pgDumpStderr,omitempty"`
	ZstdStderr          string       `json:"zstdStderr,omitempty"`
	ReportPath          string       `json:"reportPath,omitempty"`
	Profile             *profileInfo `json:"profile,omitempty"`
}

type runSummary struct {
	StartedAt                time.Time    `json:"startedAt"`
	EndedAt                  time.Time    `json:"endedAt"`
	TotalDuration            durationJSON `json:"totalDuration"`
	TotalDumpDuration        durationJSON `json:"totalDumpDuration"`
	TotalCompressionDuration durationJSON `json:"totalCompressionDuration"`
	Total                    int          `json:"total"`
	Succeeded                int          `json:"succeeded"`
	Failed                   int          `json:"failed"`
	TotalInputBytes          int64        `json:"totalInputBytes"`
	TotalOutputBytes         int64        `json:"totalOutputBytes"`
	TotalInputHuman          string       `json:"totalInputHuman"`
	TotalOutputHuman         string       `json:"totalOutputHuman"`
	OverallRatio             float64      `json:"overallRatio"`
	OverallPercent           float64      `json:"overallPercent"`
	ProcessMode              string       `json:"processMode"`
	Concurrency              int          `json:"concurrency"`
	PgDumpBin                string       `json:"pgDumpBin"`
	ZstdBin                  string       `json:"zstdBin"`
	SourceMode               string       `json:"sourceMode"`
	SourceFiles              string       `json:"sourceFiles"`
	RunReportPath            string       `json:"runReportPath,omitempty"`
	Profile                  *profileInfo `json:"profile,omitempty"`
}

type profileInfo struct {
	Goroutines      int `json:"goroutines"`
	ConcurrencyUsed int `json:"concurrencyUsed"`
}

type countingWriter struct {
	count int64
}

func (writer *countingWriter) Write(p []byte) (int, error) {
	writer.count += int64(len(p))
	return len(p), nil
}

func runTasks(cfg config, tasks []backupTask) ([]backupResult, runSummary) {
	startedAt := time.Now()
	logger.Infof("starting task runner: tasks=%d process-mode=%s requested-concurrency=%d", len(tasks), cfg.ProcessMode, cfg.Concurrency)
	results := make([]backupResult, len(tasks))

	concurrency := cfg.Concurrency
	if cfg.ProcessMode == "sync" {
		concurrency = 1
	}
	if concurrency > len(tasks) {
		concurrency = len(tasks)
	}
	if concurrency < 1 {
		concurrency = 1
	}
	logger.Infof("effective concurrency: %d", concurrency)

	queue := make(chan int)
	var wg sync.WaitGroup

	for worker := 0; worker < concurrency; worker++ {
		workerID := worker + 1
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			logger.Infof("worker started: id=%d", id)
			for idx := range queue {
				logger.Infof("worker picked task: worker=%d task-index=%d db=%s", id, idx, tasks[idx].Database)
				result := runSingleTask(cfg, tasks[idx])
				if cfg.EnableDBReport && !tasks[idx].DisableReport {
					reportPath, err := writeDBReport(cfg, result)
					if err != nil {
						result.Error = mergeError(result.Error, fmt.Sprintf("write db report: %v", err))
						result.Success = false
						logger.Warnf("db report write failed: worker=%d db=%s err=%v", id, tasks[idx].Database, err)
					} else {
						result.ReportPath = reportPath
						logger.Infof("db report written: worker=%d db=%s path=%s", id, tasks[idx].Database, reportPath)
					}
				}
				results[idx] = result
				logger.Infof("worker finished task: worker=%d task-index=%d db=%s success=%t duration=%s",
					id, idx, tasks[idx].Database, result.Success, time.Duration(result.TotalDuration))
			}
			logger.Infof("worker stopped: id=%d", id)
		}(workerID)
	}

	for i := range tasks {
		logger.Debugf("queue task: index=%d db=%s host=%s", i, tasks[i].Database, tasks[i].Host)
		queue <- i
	}
	close(queue)
	logger.Debugf("task queue closed")
	wg.Wait()
	logger.Infof("all workers completed")

	summary := aggregateSummary(cfg, startedAt, time.Now(), results, concurrency)
	logger.Infof("summary aggregated: total=%d succeeded=%d failed=%d", summary.Total, summary.Succeeded, summary.Failed)
	return results, summary
}

func runSingleTask(cfg config, task backupTask) backupResult {
	startedAt := time.Now()
	outputPath := buildOutputPath(cfg, task.OutputName, task.Host, task.Database, task.Compress, startedAt)
	logger.Infof("starting backup task: db=%s host=%s port=%d user=%s out=%s compress=%t", task.Database, task.Host, task.Port, task.Username, outputPath, task.Compress)

	result := backupResult{
		Task:               task,
		OutputPath:         outputPath,
		StartedAt:          startedAt,
		PgDumpBin:          cfg.PgDumpBin,
		ZstdBin:            cfg.ZstdBin,
		CompressionEnabled: task.Compress,
		Success:            false,
	}
	if err := ensureParentDir(result.OutputPath); err != nil {
		result.Error = fmt.Sprintf("ensure output parent dir: %v", err)
		result.EndedAt = time.Now()
		result.TotalDuration = durationJSON(result.EndedAt.Sub(startedAt))
		return finalizeResult(cfg, result)
	}
	if strings.TrimSpace(task.PrecheckError) != "" {
		result.Error = "source precheck failed: " + task.PrecheckError
		logger.Warnf("skipping task due to precheck error: db=%s err=%s", task.Database, task.PrecheckError)
		result.EndedAt = time.Now()
		result.TotalDuration = durationJSON(result.EndedAt.Sub(startedAt))
		return finalizeResult(cfg, result)
	}

	if task.Compress {
		result.ExecutionPath = "stream-pgdump-to-zstd"
		result = runCompressedStreamingTask(cfg, task, result)
	} else {
		result.ExecutionPath = "direct-plain-pgdump-file"
		result = runPlainTask(cfg, task, result)
	}

	result.EndedAt = time.Now()
	result.TotalDuration = durationJSON(result.EndedAt.Sub(startedAt))
	return finalizeResult(cfg, result)
}

func runPlainTask(cfg config, task backupTask, result backupResult) backupResult {
	removeFileIfExists(result.OutputPath)
	pgDumpArgs := []string{
		"-h", task.Host,
		"-p", fmt.Sprintf("%d", task.Port),
		"-U", task.Username,
		"-d", task.Database,
		"--format=plain",
		"--file", result.OutputPath,
	}
	result.PgDumpArgs = append([]string{}, pgDumpArgs...)
	logger.Debugf("pg_dump args (plain path) db=%s: %v", task.Database, pgDumpArgs)

	pgDumpCmd := exec.Command(cfg.PgDumpBin, pgDumpArgs...)
	pgDumpCmd.Env = append(os.Environ(), "PGPASSWORD="+task.Password)
	pgDumpStderr, err := pgDumpCmd.StderrPipe()
	if err != nil {
		result.Error = fmt.Sprintf("pg_dump stderr pipe: %v", err)
		return result
	}

	dumpStart := time.Now()
	if err := pgDumpCmd.Start(); err != nil {
		result.Error = fmt.Sprintf("start pg_dump: %v", err)
		logger.Errorf("pg_dump start failed (plain path): db=%s err=%v", task.Database, err)
		return result
	}
	logger.Debugf("waiting for pg_dump completion (plain path): db=%s", task.Database)
	pgDumpWaitErr := pgDumpCmd.Wait()
	dumpEnd := time.Now()
	result.DumpDuration = durationJSON(dumpEnd.Sub(dumpStart))
	result.PgDumpStderr = strings.TrimSpace(<-readPipeText(pgDumpStderr))

	if pgDumpWaitErr != nil {
		result.Error = mergeError(result.Error, fmt.Sprintf("pg_dump failed: %v", pgDumpWaitErr))
		_ = os.Remove(result.OutputPath)
		logger.Warnf("plain backup failed, output removed: db=%s path=%s err=%s", task.Database, result.OutputPath, result.Error)
		return result
	}

	stat, statErr := os.Stat(result.OutputPath)
	if statErr != nil {
		result.Error = mergeError(result.Error, fmt.Sprintf("stat output file: %v", statErr))
		return result
	}

	result.UncompressedBytes = stat.Size()
	result.CompressedBytes = result.UncompressedBytes
	result.Success = true
	logger.Infof("plain backup succeeded: db=%s path=%s", task.Database, result.OutputPath)
	return result
}

func runCompressedStreamingTask(cfg config, task backupTask, result backupResult) backupResult {
	finalTempOutput := result.OutputPath + ".part"
	removeFileIfExists(finalTempOutput)
	pgDumpArgs := []string{
		"-h", task.Host,
		"-p", fmt.Sprintf("%d", task.Port),
		"-U", task.Username,
		"-d", task.Database,
		"--format=plain",
	}
	result.PgDumpArgs = append([]string{}, pgDumpArgs...)
	logger.Debugf("pg_dump args (stream path) db=%s: %v", task.Database, pgDumpArgs)

	zstdArgs := []string{fmt.Sprintf("-%d", cfg.ZstdLevel)}
	if cfg.UseAllCPUs {
		zstdArgs = append(zstdArgs, "-T0")
	} else {
		zstdArgs = append(zstdArgs, "-T", fmt.Sprintf("%d", cfg.ZstdThreadCount))
	}
	zstdArgs = append(zstdArgs, "-o", finalTempOutput)
	result.ZstdArgs = append([]string{}, zstdArgs...)
	logger.Debugf("zstd args (stream path) db=%s: %v", task.Database, zstdArgs)

	pgDumpCmd := exec.Command(cfg.PgDumpBin, pgDumpArgs...)
	pgDumpCmd.Env = append(os.Environ(), "PGPASSWORD="+task.Password)
	zstdCmd := exec.Command(cfg.ZstdBin, zstdArgs...)

	pgDumpStdout, err := pgDumpCmd.StdoutPipe()
	if err != nil {
		result.Error = fmt.Sprintf("pg_dump stdout pipe: %v", err)
		return result
	}
	pgDumpStderr, err := pgDumpCmd.StderrPipe()
	if err != nil {
		result.Error = fmt.Sprintf("pg_dump stderr pipe: %v", err)
		return result
	}
	zstdStderr, err := zstdCmd.StderrPipe()
	if err != nil {
		result.Error = fmt.Sprintf("zstd stderr pipe: %v", err)
		return result
	}

	countWriter := &countingWriter{}
	pipeReader, pipeWriter := io.Pipe()
	zstdCmd.Stdin = pipeReader

	if err := pgDumpCmd.Start(); err != nil {
		result.Error = fmt.Sprintf("start pg_dump: %v", err)
		logger.Errorf("pg_dump start failed (stream path): db=%s err=%v", task.Database, err)
		_ = pipeReader.Close()
		_ = pipeWriter.Close()
		return result
	}
	if err := zstdCmd.Start(); err != nil {
		_ = pgDumpCmd.Process.Kill()
		_ = pgDumpCmd.Wait()
		result.Error = fmt.Sprintf("start zstd: %v", err)
		logger.Errorf("zstd start failed (stream path): db=%s err=%v", task.Database, err)
		_ = pipeReader.Close()
		_ = pipeWriter.Close()
		return result
	}

	var copyErr atomic.Value
	copyDone := make(chan struct{})
	go func() {
		_, err := io.Copy(io.MultiWriter(pipeWriter, countWriter), pgDumpStdout)
		if err != nil && !errors.Is(err, io.ErrClosedPipe) {
			copyErr.Store(err)
		}
		_ = pipeWriter.Close()
		close(copyDone)
	}()

	pgDumpErrCh := readPipeText(pgDumpStderr)
	zstdErrCh := readPipeText(zstdStderr)

	dumpStart := time.Now()
	logger.Debugf("waiting for pg_dump completion (stream path): db=%s", task.Database)
	pgDumpWaitErr := pgDumpCmd.Wait()
	dumpEnd := time.Now()
	result.DumpDuration = durationJSON(dumpEnd.Sub(dumpStart))
	result.PgDumpStderr = strings.TrimSpace(<-pgDumpErrCh)
	logger.Infof("pg_dump completed (stream path): db=%s duration=%s err=%v", task.Database, time.Duration(result.DumpDuration), pgDumpWaitErr)

	<-copyDone
	_ = pipeReader.Close()

	compStart := dumpEnd
	logger.Debugf("waiting for zstd completion (stream path): db=%s", task.Database)
	zstdWaitErr := zstdCmd.Wait()
	compEnd := time.Now()
	result.CompressionDuration = durationJSON(compEnd.Sub(compStart))
	result.ZstdStderr = strings.TrimSpace(<-zstdErrCh)
	logger.Infof("zstd completed (stream path): db=%s duration=%s err=%v", task.Database, time.Duration(result.CompressionDuration), zstdWaitErr)

	result.UncompressedBytes = countWriter.count
	if errAny := copyErr.Load(); errAny != nil {
		result.Error = mergeError(result.Error, fmt.Sprintf("stream copy: %v", errAny))
	}
	if pgDumpWaitErr != nil {
		result.Error = mergeError(result.Error, fmt.Sprintf("pg_dump failed: %v", pgDumpWaitErr))
	}
	if zstdWaitErr != nil {
		result.Error = mergeError(result.Error, fmt.Sprintf("zstd failed: %v", zstdWaitErr))
	}

	if result.Error != "" {
		_ = os.Remove(finalTempOutput)
		logger.Warnf("compressed backup failed, partial archive removed: db=%s path=%s err=%s", task.Database, finalTempOutput, result.Error)
		return result
	}

	if stat, statErr := os.Stat(finalTempOutput); statErr == nil {
		result.CompressedBytes = stat.Size()
	} else {
		result.Error = mergeError(result.Error, fmt.Sprintf("stat compressed output: %v", statErr))
		_ = os.Remove(finalTempOutput)
		return result
	}

	if err := os.Rename(finalTempOutput, result.OutputPath); err != nil {
		result.Error = mergeError(result.Error, fmt.Sprintf("promote archive: %v", err))
		_ = os.Remove(finalTempOutput)
		return result
	}

	result.Success = true
	logger.Infof("compressed backup succeeded: db=%s path=%s", task.Database, result.OutputPath)
	return result
}

func finalizeResult(cfg config, result backupResult) backupResult {
	computeRatios(&result)
	result.UncompressedHuman = formatBytesIEC(result.UncompressedBytes)
	result.CompressedHuman = formatBytesIEC(result.CompressedBytes)
	if cfg.ProfileSummary {
		result.Profile = &profileInfo{Goroutines: runtime.NumGoroutine(), ConcurrencyUsed: 1}
		logger.Debugf("profile summary captured: db=%s goroutines=%d", result.Task.Database, result.Profile.Goroutines)
	}
	logger.Infof("size stats: db=%s input=%s(%d) output=%s(%d) ratio=%.4f saved=%.2f%%",
		result.Task.Database,
		result.UncompressedHuman, result.UncompressedBytes,
		result.CompressedHuman, result.CompressedBytes,
		result.CompressionRatio,
		result.CompressionPercent,
	)
	logger.Infof("backup task completed: db=%s total-duration=%s path=%s", result.Task.Database, time.Duration(result.TotalDuration), result.ExecutionPath)
	return result
}

func computeRatios(result *backupResult) {
	if result.UncompressedBytes <= 0 {
		result.CompressionRatio = 0
		result.CompressionPercent = 0
		return
	}
	ratio := float64(result.CompressedBytes) / float64(result.UncompressedBytes)
	result.CompressionRatio = ratio
	result.CompressionPercent = (1 - ratio) * 100
}

func readPipeText(reader io.Reader) <-chan string {
	out := make(chan string, 1)
	go func() {
		defer close(out)
		data, err := io.ReadAll(reader)
		if err != nil {
			out <- fmt.Sprintf("read stderr error: %v", err)
			return
		}
		out <- string(data)
	}()
	return out
}

func mergeError(current, next string) string {
	if strings.TrimSpace(next) == "" {
		return current
	}
	if strings.TrimSpace(current) == "" {
		return next
	}
	return current + "; " + next
}

func aggregateSummary(cfg config, startedAt, endedAt time.Time, results []backupResult, concurrencyUsed int) runSummary {
	summary := runSummary{
		StartedAt:     startedAt,
		EndedAt:       endedAt,
		TotalDuration: durationJSON(endedAt.Sub(startedAt)),
		Total:         len(results),
		ProcessMode:   cfg.ProcessMode,
		Concurrency:   concurrencyUsed,
		PgDumpBin:     cfg.PgDumpBin,
		ZstdBin:       cfg.ZstdBin,
		SourceMode:    cfg.SourceMode,
		SourceFiles:   cfg.SourceFiles,
	}

	for _, result := range results {
		if result.Success {
			summary.Succeeded++
		} else {
			summary.Failed++
		}
		summary.TotalInputBytes += result.UncompressedBytes
		summary.TotalOutputBytes += result.CompressedBytes
		summary.TotalDumpDuration += result.DumpDuration
		summary.TotalCompressionDuration += result.CompressionDuration
	}
	logger.Debugf("aggregate bytes: input=%d output=%d", summary.TotalInputBytes, summary.TotalOutputBytes)

	if summary.TotalInputBytes > 0 {
		ratio := float64(summary.TotalOutputBytes) / float64(summary.TotalInputBytes)
		summary.OverallRatio = ratio
		summary.OverallPercent = (1 - ratio) * 100
	}
	summary.TotalInputHuman = formatBytesIEC(summary.TotalInputBytes)
	summary.TotalOutputHuman = formatBytesIEC(summary.TotalOutputBytes)
	if cfg.ProfileSummary {
		summary.Profile = &profileInfo{Goroutines: runtime.NumGoroutine(), ConcurrencyUsed: concurrencyUsed}
	}

	return summary
}

func printRunSummary(summary runSummary) {
	fmt.Printf("run completed: total=%d succeeded=%d failed=%d duration=%s\n", summary.Total, summary.Succeeded, summary.Failed, time.Duration(summary.TotalDuration))
	fmt.Printf("timings: total-dump=%s total-compression=%s\n", time.Duration(summary.TotalDumpDuration), time.Duration(summary.TotalCompressionDuration))
	fmt.Printf("size totals: input=%s(%d) output=%s(%d) ratio=%.4f saved=%.2f%%\n",
		summary.TotalInputHuman,
		summary.TotalInputBytes,
		summary.TotalOutputHuman,
		summary.TotalOutputBytes,
		summary.OverallRatio,
		summary.OverallPercent,
	)
	if summary.RunReportPath != "" {
		fmt.Printf("run report: %s\n", summary.RunReportPath)
	}
}

func formatBytesIEC(bytes int64) string {
	if bytes < 0 {
		return "0 B"
	}
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	prefixes := []string{"KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}
	return fmt.Sprintf("%.2f %s", float64(bytes)/float64(div), prefixes[exp])
}

func removeFileIfExists(path string) {
	if strings.TrimSpace(path) == "" {
		return
	}
	if _, err := os.Stat(path); err != nil {
		return
	}
	if err := os.Remove(path); err != nil {
		logger.Debugf("remove file skipped: path=%s err=%v", path, err)
	}
}

func ensureParentDir(path string) error {
	parent := filepath.Dir(path)
	if parent == "." || parent == "" {
		return nil
	}
	return os.MkdirAll(parent, 0o755)
}
