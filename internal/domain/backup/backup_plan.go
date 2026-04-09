package backup

import (
	"github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/domain/shared"
	zhinuxtypes "github.com/amirhossein-shakeri/zhinux-platform/types"
)

type BackupPlan struct {
	ID       string
	PublicID string

	DatabaseID zhinuxtypes.ID
	Schedule   string // Cron perhaps
	Enabled    bool

	RetentionPolicy    shared.RetentionPolicy
	CompressionEnabled bool
	EncryptionEnabled  bool
	// Compression     CompressionConfig
	// Encryption      EncryptionConfig
}
