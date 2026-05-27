package backup

import (
	"time"
)

type ExecutionTimings struct {
	QueuedAt              *time.Time
	StartedAt             *time.Time
	DumpStartedAt         *time.Time
	DumpFinishedAt        *time.Time
	CompressionStartedAt  *time.Time
	CompressionFinishedAt *time.Time
	StorageCommittedAt    *time.Time
	FinishedAt            *time.Time
}

type BackupReport struct {
	DumpDuration        time.Duration
	CompressionDuration time.Duration
	TotalDuration       time.Duration
	CompressionStats    CompressionStats
	ResourceSnapshot    ResourceSnapshot
}

type ResourceSnapshot struct {
	MemoryBytes      int
	CPUPercent       float64
	SystemPressure   string
	StorageQuotaUsed int
	EstimatedCost    float64
	CostCurrency     string
}
