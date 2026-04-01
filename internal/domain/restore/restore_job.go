package restore

import "time"

type RestoreJob struct {
	ID             string
	ArtifactID     string
	TargetDatabase string // todo: DB ID?
	Status         RestoreStatus
	StartedAt      *time.Time
	FinishedAt     *time.Time
}
