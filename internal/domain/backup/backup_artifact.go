package backup

import (
	"time"

	"github.com/amirhossein-shakeri/zhinux-platform/types"
)

type BackupArtifact struct {
	ID       types.ID // Internal ID(Fast joins)
	PublicID string   // Exposed UUID at public APIs(Safe public identifiers)

	DatabaseID  types.ID
	BackupJobID types.ID
	Format      BackupFormat
	Storage     StorageLocation
	Size        int
	Checksum    Checksum
	Compression CompressionStats
	Metadata    map[string]string

	CreatedAt time.Time
	DeletedAt *time.Time
}

func (a BackupArtifact) StorageLocation() string {
	return a.Storage.String()
}

func (a BackupArtifact) CompressedSize() int {
	if a.Compression.CompressedBytes > 0 {
		return a.Compression.CompressedBytes
	}
	return a.Size
}

func (a BackupArtifact) UncompressedSize() int {
	return a.Compression.UncompressedBytes
}

func (a BackupArtifact) CompressionRatio() float64 {
	return a.Compression.Ratio()
}
