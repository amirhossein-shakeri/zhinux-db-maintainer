package restore

import (
	"time"

	zhinuxtypes "github.com/amirhossein-shakeri/zhinux-platform/types"
)

type RestoreJob struct {
	ID       string
	PublicID string

	ArtifactID       string
	TargetDatabaseID zhinuxtypes.ID
	Status           RestoreStatus
	StartedAt        *time.Time
	FinishedAt       *time.Time
}
