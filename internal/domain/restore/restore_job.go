package restore

import "time"

type RestoreJob struct {
	ID               string
	ArtifactID       string
	TargetDatabaseID string
	Status           RestoreStatus
	StartedAt        *time.Time
	FinishedAt       *time.Time
}
