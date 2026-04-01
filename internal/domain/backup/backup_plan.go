package backup

import "github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/domain/shared"

type BackupPlan struct {
	ID         string
	DatabaseID string
	Schedule   string // Cron perhaps
	Enabled    bool

	RetentionPolicy    shared.RetentionPolicy
	CompressionEnabled bool
	EncryptionEnabled  bool
	// Compression     CompressionConfig
	// Encryption      EncryptionConfig
}
