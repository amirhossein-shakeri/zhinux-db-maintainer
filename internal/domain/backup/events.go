package backup

import (
	"time"

	"github.com/amirhossein-shakeri/zhinux-platform/types"
)

const (
	EventBackupPlanCreated     = "db_maintainer.backup.plan_created"
	EventBackupRequested       = "db_maintainer.backup.requested"
	EventBackupStarted         = "db_maintainer.backup.started"
	EventBackupArtifactCreated = "db_maintainer.backup.artifact_created"
	EventBackupCompleted       = "db_maintainer.backup.completed"
	EventBackupFailed          = "db_maintainer.backup.failed"
)

type DomainEvent struct {
	EventName string
	Data      any
	At        time.Time
}

func (e DomainEvent) Name() string {
	return e.EventName
}

func (e DomainEvent) Payload() any {
	return e.Data
}

func (e DomainEvent) OccuredAt() time.Time {
	return e.At
}

type BackupPlanCreatedPayload struct {
	PlanID     types.ID
	DatabaseID types.ID
}

type BackupRequestedPayload struct {
	JobID         types.ID
	DatabaseID    types.ID
	PlanID        *types.ID
	TriggerType   BackupTrigger
	CorrelationID string
}

type BackupStartedPayload struct {
	JobID     types.ID
	StartedAt time.Time
}

type BackupArtifactCreatedPayload struct {
	ArtifactID types.ID
	JobID      types.ID
	DatabaseID types.ID
	Storage    StorageLocation
	Size       int
}

type BackupCompletedPayload struct {
	JobID      types.ID
	ArtifactID types.ID
	Report     BackupReport
}

type BackupFailedPayload struct {
	JobID  types.ID
	Reason string
}
