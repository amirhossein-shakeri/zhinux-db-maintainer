package backup

import (
	"time"

	zhinuxtypes "github.com/amirhossein-shakeri/zhinux-platform/types"
)

type BackupJob struct {
	ID       string
	PublicID string

	DatabaseID  zhinuxtypes.ID
	TriggerType BackupTrigger
	Status      BackupStatus
	StartedAt   *time.Time
	FinishedAt  *time.Time
	ArtifactID  *string
}
