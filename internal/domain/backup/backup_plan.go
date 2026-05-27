package backup

import (
	"github.com/amirhossein-shakeri/zhinux-db-maintainer/internal/domain/shared"
	"github.com/amirhossein-shakeri/zhinux-platform/types"
)

type BackupPlan struct {
	ID       types.ID // Internal ID(Fast joins)
	PublicID string   // Exposed UUID at public APIs(Safe public identifiers)

	DatabaseID types.ID
	Schedule   shared.Schedule
	Enabled    bool

	Retention   shared.RetentionConfig
	Compression CompressionConfig
	Encryption  EncryptionConfig
}

func (p BackupPlan) CompressionEnabled() bool {
	return p.Compression.Enabled
}

func (p BackupPlan) EncryptionEnabled() bool {
	return p.Encryption.Enabled
}
