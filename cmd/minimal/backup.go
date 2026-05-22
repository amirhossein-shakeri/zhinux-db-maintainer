package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type backupResult struct {
	Task                   backupTask   `json:"task"`
	OutputPath             string       `json:"outputPath"`
	StartedAt              time.Time    `json:"startedAt"`
	EndedAt                time.Time    `json:"endedAt"`
	DumpDuration           durationJSON `json:"dumpDuration"`
	CompressionDuration    durationJSON `json:"compressionDuration"`
	TotalDuration          durationJSON `json:"totalDuration"`
	UncompressedBytes      int64        `json:"uncompressedBytes"`
	CompressedBytes        int64        `json:"compressedBytes"`
	CompressionRatio       float64      `json:"compressionRatio"`
	CompressionPercent     float64      `json:"compressionPercent"`
	PgDumpBin              string       `json:"pgDumpBin"`
	ZstdBin                string       `json:"zstdBin"`
	ZstdArgs               []string     `json:"zstdArgs"`
	PgDumpArgs             []string     `json:"pgDumpArgs"`
	Success                bool         `json:"success"`
	Error                  string       `json:"error,omitempty"`
	PgDumpStderr           string       `json:"pgDumpStderr,omitempty"`
	ZstdStderr             string       `json:"zstdStderr,omitempty"`
	ReportPath             string       `json:"reportPath,omitempty"`
	Profile                *profileInfo `json:"profile,omitempty"`
}

type runSummary struct {
	StartedAt         time.Time    `json:"startedAt"`
	EndedAt           time.Time    `json:"endedAt"`
	TotalDuration     durationJSON `json:"totalDuration"`
	Total             int          `json:"total"`
	Succeeded         int          `json:"succeeded"`
	Failed            int          `json:"failed"`
	TotalInputBytes   int64        `json:"totalInputBytes"`
	TotalOutputBytes  int64        `json:"totalOutputBytes"`
	OverallRatio      float64      `json:"overallRatio"`
	OverallPercent    float64      `json:"overallPercent"`
	ProcessMode       string       `json:"processMode"`
	Concurrency       int          `json:"concurrency"`
	PgDumpBin         string       `json:"pgDumpBin"`
	ZstdBin           string       `json:"zstdBin"`
	SourceMode        string       `json:"sourceMode"`
	SourceFiles       string       `json:"sourceFiles"`
	RunReportPath     string       `json:"runReportPath,omitempty"`
	Profile           *profileInfo `json:"profile,omitempty"`
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

	queue := make(chan int)
	var wg sync.WaitGroup

	for worker := 0; worker < concurrency; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range queue {
				result := runSingleTask(cfg, tasks[idx])
				if cfg.EnableDBReport && !tasks[idx].DisableReport {
					reportPath, err := writeDBReport(cfg, result)
					if err != nil {
						result.Error = mergeError(result.Error, fmt.Sprintf("write db report: %v", err))
						result.Success = false
					} else {
						result.ReportPath = reportPath
					}
				}
				results[idx] = result
			}
		}()
	}

	for i := range tasks {
		queue <- i
	}
	close(queue)
	wg.Wait()

	summary := aggregateSummary(cfg, startedAt, time.Now(), results, concurrency)
	return results, summary
}

func runSingleTask(cfg config, task backupTask) backupResult {
	startedAt := time.Now()
	outputPath := buildOutputPath(cfg.OutputDir, task.OutputName, task.Host, task.Database, startedAt)

	result := backupResult{
		Task:       task,
		OutputPath: outputPath,
		StartedAt:  startedAt,
		PgDumpBin:  cfg.PgDumpBin,
		ZstdBin:    cfg.ZstdBin,
		Success:    false,
	}
	if strings.TrimSpace(task.PrecheckError) != "" {
		result.Error = "source precheck failed: " + task.PrecheckError
		result.EndedAt = time.Now()
		result.TotalDuration = durationJSON(result.EndedAt.Sub(startedAt))
		return result
	}

	pgDumpArgs := []string{
		"-h", task.Host,
		"-p", fmt.Sprintf("%d", task.Port),
		"-U", task.Username,
		"-d", task.Database,
		"--format=plain",
	}
	result.PgDumpArgs = append([]string{}, pgDumpArgs...)

	zstdArgs := []string{fmt.Sprintf("-%d", cfg.ZstdLevel)}
	if cfg.UseAllCPUs {
		zstdArgs = append(zstdArgs, "-T0")
	} else {
		zstdArgs = append(zstdArgs, "-T", fmt.Sprintf("%d", cfg.ZstdThreadCount))
	}
	zstdArgs = append(zstdArgs, "-o", outputPath)
	result.ZstdArgs = append([]string{}, zstdArgs...)

	pgDumpCmd := exec.Command(cfg.PgDumpBin, pgDumpArgs...)
	pgDumpCmd.Env = append(os.Environ(), "PGPASSWORD="+task.Password)

	zstdCmd := exec.Command(cfg.ZstdBin, zstdArgs...)

	pgDumpStdout, err := pgDumpCmd.StdoutPipe()
	if err != nil {
		result.Error = fmt.Sprintf("pg_dump stdout pipe: %v", err)
		result.EndedAt = time.Now()
		result.TotalDuration = durationJSON(result.EndedAt.Sub(startedAt))
		return result
	}
	pgDumpStderr, err := pgDumpCmd.StderrPipe()
	if err != nil {
		result.Error = fmt.Sprintf("pg_dump stderr pipe: %v", err)
		result.EndedAt = time.Now()
		result.TotalDuration = durationJSON(result.EndedAt.Sub(startedAt))
		return result
	}
	zstdStderr, err := zstdCmd.StderrPipe()
	if err != nil {
		result.Error = fmt.Sprintf("zstd stderr pipe: %v", err)
		result.EndedAt = time.Now()
		result.TotalDuration = durationJSON(result.EndedAt.Sub(startedAt))
		return result
	}

	countWriter := &countingWriter{}
	pipeReader, pipeWriter := io.Pipe()

	zstdCmd.Stdin = pipeReader

	if err := pgDumpCmd.Start(); err != nil {
		result.Error = fmt.Sprintf("start pg_dump: %v", err)
		result.EndedAt = time.Now()
		result.TotalDuration = durationJSON(result.EndedAt.Sub(startedAt))
		_ = pipeReader.Close()
		_ = pipeWriter.Close()
		return result
	}
	if err := zstdCmd.Start(); err != nil {
		_ = pgDumpCmd.Process.Kill()
		_ = pgDumpCmd.Wait()
		result.Error = fmt.Sprintf("start zstd: %v", err)
		result.EndedAt = time.Now()
		result.TotalDuration = durationJSON(result.EndedAt.Sub(startedAt))
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

	pgDumpErrCh := make(chan string, 1)
	zstdErrCh := make(chan string, 1)
	go copyPipeOutput(pgDumpStderr, pgDumpErrCh)
	go copyPipeOutput(zstdStderr, zstdErrCh)

	dumpStart := time.Now()
	pgDumpWaitErr := pgDumpCmd.Wait()
	dumpEnd := time.Now()
	result.DumpDuration = durationJSON(dumpEnd.Sub(dumpStart))

	<-copyDone
	_ = pipeReader.Close()

	compStart := dumpEnd
	zstdWaitErr := zstdCmd.Wait()
	compEnd := time.Now()
	result.CompressionDuration = durationJSON(compEnd.Sub(compStart))

	pgDumpErrText := <-pgDumpErrCh
	zstdErrText := <-zstdErrCh
	result.PgDumpStderr = strings.TrimSpace(pgDumpErrText)
	result.ZstdStderr = strings.TrimSpace(zstdErrText)

	result.UncompressedBytes = countWriter.count
	if stat, statErr := os.Stat(outputPath); statErr == nil {
		result.CompressedBytes = stat.Size()
	}
	computeRatios(&result)

	if errAny := copyErr.Load(); errAny != nil {
		result.Error = mergeError(result.Error, fmt.Sprintf("stream copy: %v", errAny))
	}
	if pgDumpWaitErr != nil {
		result.Error = mergeError(result.Error, fmt.Sprintf("pg_dump failed: %v", pgDumpWaitErr))
	}
	if zstdWaitErr != nil {
		result.Error = mergeError(result.Error, fmt.Sprintf("zstd failed: %v", zstdWaitErr))
	}

	if result.Error == "" {
		result.Success = true
	}

	result.EndedAt = time.Now()
	result.TotalDuration = durationJSON(result.EndedAt.Sub(startedAt))
	if cfg.ProfileSummary {
		result.Profile = &profileInfo{Goroutines: runtime.NumGoroutine(), ConcurrencyUsed: 1}
	}

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

func copyPipeOutput(reader io.Reader, out chan<- string) {
	defer close(out)
	data, err := io.ReadAll(reader)
	if err != nil {
		out <- fmt.Sprintf("read stderr error: %v", err)
		return
	}
	out <- string(data)
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
	}

	if summary.TotalInputBytes > 0 {
		ratio := float64(summary.TotalOutputBytes) / float64(summary.TotalInputBytes)
		summary.OverallRatio = ratio
		summary.OverallPercent = (1 - ratio) * 100
	}
	if cfg.ProfileSummary {
		summary.Profile = &profileInfo{Goroutines: runtime.NumGoroutine(), ConcurrencyUsed: concurrencyUsed}
	}

	return summary
}

func printRunSummary(summary runSummary) {
	fmt.Printf("run completed: total=%d succeeded=%d failed=%d duration=%s\n", summary.Total, summary.Succeeded, summary.Failed, time.Duration(summary.TotalDuration))
	fmt.Printf("size totals: input=%d output=%d ratio=%.4f saved=%.2f%%\n", summary.TotalInputBytes, summary.TotalOutputBytes, summary.OverallRatio, summary.OverallPercent)
	if summary.RunReportPath != "" {
		fmt.Printf("run report: %s\n", summary.RunReportPath)
	}
}
