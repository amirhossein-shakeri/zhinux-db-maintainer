package backup

import (
	"time"

	"github.com/amirhossein-shakeri/zhinux-platform/types"
)

type BackupJob struct {
	ID       types.ID // Internal ID(Fast joins)
	PublicID string   // Exposed UUID at public APIs(Safe public identifiers)

	DatabaseID  types.ID
	TriggerType BackupTrigger
	Status      BackupStatus
	StartedAt   *time.Time
	FinishedAt  *time.Time
	ArtifactID  *types.ID
}
