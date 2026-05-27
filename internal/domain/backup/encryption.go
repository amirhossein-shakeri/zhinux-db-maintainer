package backup

type EncryptionAlgorithm string

const (
	EncryptionAlgorithmNone EncryptionAlgorithm = "none"
)

type EncryptionConfig struct {
	Enabled      bool
	Algorithm    EncryptionAlgorithm
	KeyReference string
}

func NoEncryption() EncryptionConfig {
	return EncryptionConfig{
		Enabled:   false,
		Algorithm: EncryptionAlgorithmNone,
	}
}
