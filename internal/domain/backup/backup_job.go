package backup

import (
	"time"

	"github.com/amirhossein-shakeri/zhinux-platform/types"
)

type BackupJob struct {
	ID       types.ID // Internal ID(Fast joins)
	PublicID string   // Exposed UUID at public APIs(Safe public identifiers)

	DatabaseID     types.ID
	PlanID         *types.ID
	TriggerType    BackupTrigger
	Status         BackupStatus
	Format         BackupFormat
	Compression    CompressionConfig
	Execution      ExecutionTimings
	Report         *BackupReport
	FailureReason  string
	RetryCount     int
	Actor          string
	CorrelationID  string
	IdempotencyKey string
	ArtifactID     *types.ID
	StartedAt      *time.Time
	FinishedAt     *time.Time
}

func (j *BackupJob) MarkStarted(at time.Time) bool {
	if j == nil || !CanTransitionBackupStatus(j.Status, BackupStatusInProgress) {
		return false
	}
	j.Status = BackupStatusInProgress
	j.StartedAt = &at
	j.Execution.StartedAt = &at
	return true
}

func (j *BackupJob) MarkSucceeded(at time.Time, artifactID types.ID, report BackupReport) bool {
	if j == nil || !CanTransitionBackupStatus(j.Status, BackupStatusSucceeded) {
		return false
	}
	j.Status = BackupStatusSucceeded
	j.FinishedAt = &at
	j.Execution.FinishedAt = &at
	j.ArtifactID = &artifactID
	j.Report = &report
	return true
}

func (j *BackupJob) MarkFailed(at time.Time, reason string) bool {
	if j == nil || !CanTransitionBackupStatus(j.Status, BackupStatusFailed) {
		return false
	}
	j.Status = BackupStatusFailed
	j.FinishedAt = &at
	j.Execution.FinishedAt = &at
	j.FailureReason = reason
	return true
}
