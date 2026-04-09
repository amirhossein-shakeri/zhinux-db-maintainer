package backup

import (
	"time"

	zhinuxtypes "github.com/amirhossein-shakeri/zhinux-platform/types"
)

type BackupArtifact struct {
	ID       string
	PublicID string

	DatabaseID      zhinuxtypes.ID
	BackupJobID     string
	StorageLocation string // todo: Use more detailed address? struct {Provider, Path, Bucket}
	Size            int
	Checksum        string

	CreatedAt time.Time
	DeletedAt *time.Time
}
