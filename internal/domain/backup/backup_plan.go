package backup

import (
	"github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/domain/shared"
	"github.com/amirhossein-shakeri/zhinux-platform/types"
)

type BackupPlan struct {
	ID       types.ID // Internal ID(Fast joins)
	PublicID string   // Exposed UUID at public APIs(Safe public identifiers)

	DatabaseID types.ID
	Schedule   string // Cron perhaps
	Enabled    bool

	RetentionPolicy    shared.RetentionPolicy
	CompressionEnabled bool
	EncryptionEnabled  bool
	// Compression     CompressionConfig // todo: compression config value object is from compression service or here or shared at platform?
	// Encryption      EncryptionConfig // todo: encryption config value objects is from encryption service or here or shared at platform?
}
