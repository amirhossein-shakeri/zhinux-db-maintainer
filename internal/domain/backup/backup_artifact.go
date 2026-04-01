package backup

import "time"

type BackupArtifact struct {
	ID              string
	DatabaseID      string
	BackupJobID     string
	StorageLocation string // todo: Use more detailed address? struct {Provider, Path, Bucket}
	Size            int
	Checksum        string

	CreatedAt time.Time
	DeletedAt *time.Time
}
