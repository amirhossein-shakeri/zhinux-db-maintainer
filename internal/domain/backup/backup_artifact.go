package backup

import (
	"time"

	"github.com/amirhossein-shakeri/zhinux-platform/types"
)

type BackupArtifact struct {
	ID       types.ID // Internal ID(Fast joins)
	PublicID string   // Exposed UUID at public APIs(Safe public identifiers)

	DatabaseID      types.ID
	BackupJobID     types.ID
	StorageLocation string // todo: Use more detailed address? struct {Provider, Path, Bucket}
	Size            int    // Final size occupied in the storage
	Checksum        string

	// Compression
	CompressedSize   *int
	UncompressedSize *int
	CompressionRatio *int
	// CompressionAlgorithm // todo: compression alg value object enum from here or from compression service? Why?

	CreatedAt time.Time
	DeletedAt *time.Time
}
