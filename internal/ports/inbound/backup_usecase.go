package inbound_ports

import (
	"context"

	"github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/domain/backup"
	"github.com/amirhossein-shakeri/zhinux-platform/types"
)

type BackupUseCase interface {
	RequestBackup(ctx context.Context, request RequestBackupCommand) (*RequestBackupResult, error)
	RunBackup(ctx context.Context, request RunBackupCommand) (*RunBackupResult, error)
}

type RequestBackupCommand struct {
	DatabaseID     types.ID
	PlanID         *types.ID
	TriggerType    backup.BackupTrigger
	Format         backup.BackupFormat
	Compression    backup.CompressionConfig
	Actor          string
	CorrelationID  string
	IdempotencyKey string
}

type RequestBackupResult struct {
	Job *backup.BackupJob
}

type RunBackupCommand struct {
	JobID types.ID
}

type RunBackupResult struct {
	Job      *backup.BackupJob
	Artifact *backup.BackupArtifact
	Report   backup.BackupReport
}
