package outbound_ports

import (
	"context"
	"io"
	"time"

	"github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/domain/backup"
	"github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/domain/database"
	"github.com/amirhossein-shakeri/zhinux-platform/types"
)

type BackupIDGenerator interface {
	NewID(ctx context.Context) (types.ID, error)
}

type DumpRunner interface {
	Dump(ctx context.Context, request DumpRequest) (*DumpResult, error)
}

type DumpRequest struct {
	JobID    types.ID
	Database *database.Database
	Format   backup.BackupFormat
	Writer   io.Writer
}

type DumpResult struct {
	BytesWritten int
	Duration     time.Duration
	Stderr       string
}

type Compressor interface {
	Compress(ctx context.Context, request CompressionRequest) (*CompressionResult, error)
}

type CompressionRequest struct {
	JobID  types.ID
	Config backup.CompressionConfig
	Source io.Reader
	Target io.Writer
}

type CompressionResult struct {
	Stats  backup.CompressionStats
	Stderr string
}

type ArtifactStore interface {
	Put(ctx context.Context, request PutArtifactRequest) (*PutArtifactResult, error)
}

type PutArtifactRequest struct {
	JobID       types.ID
	DatabaseID  types.ID
	ObjectName  string
	Content     io.Reader
	ContentType string
	Metadata    map[string]string
}

type PutArtifactResult struct {
	Storage  backup.StorageLocation
	Size     int
	Checksum backup.Checksum
}

type ResourceProbe interface {
	Snapshot(ctx context.Context) (backup.ResourceSnapshot, error)
}

type BackupEventPublisher interface {
	Publish(ctx context.Context, events ...backup.DomainEvent) error
}
