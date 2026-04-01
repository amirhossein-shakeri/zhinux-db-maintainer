package backup

import "time"

type BackupJob struct {
	ID          string
	DatabaseID  string
	TriggerType BackupTrigger
	Status      BackupStatus
	StartedAt   *time.Time
	FinishedAt  *time.Time
	ArtifactID  *string
}
