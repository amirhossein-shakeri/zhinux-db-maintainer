package backup

import (
	"fmt"
	"strings"
	"time"
)

type CompressionAlgorithm string

const (
	CompressionAlgorithmNone CompressionAlgorithm = "none"
	CompressionAlgorithmZstd CompressionAlgorithm = "zstd"
)

type CompressionMode string

const (
	CompressionModeDisabled  CompressionMode = "disabled"
	CompressionModeStreaming CompressionMode = "streaming"
	CompressionModeFile      CompressionMode = "file"
)

type CompressionEngine string

const (
	CompressionEngineLibrary CompressionEngine = "library"
	CompressionEngineCLI     CompressionEngine = "cli"
)

type CompressionLevel int

func NewCompressionLevel(value int) (CompressionLevel, error) {
	if value < 1 || value > 22 {
		return 0, fmt.Errorf("compression level must be between 1 and 22")
	}
	return CompressionLevel(value), nil
}

type CompressionConfig struct {
	Enabled   bool
	Algorithm CompressionAlgorithm
	Level     CompressionLevel
	Mode      CompressionMode
	Engine    CompressionEngine
	Threads   int
}

func DefaultZstdCompression() CompressionConfig {
	return CompressionConfig{
		Enabled:   true,
		Algorithm: CompressionAlgorithmZstd,
		Level:     CompressionLevel(19),
		Mode:      CompressionModeStreaming,
		Engine:    CompressionEngineLibrary,
	}
}

func NoCompression() CompressionConfig {
	return CompressionConfig{
		Enabled:   false,
		Algorithm: CompressionAlgorithmNone,
		Mode:      CompressionModeDisabled,
	}
}

type CompressionStats struct {
	Algorithm         CompressionAlgorithm
	Engine            CompressionEngine
	Level             CompressionLevel
	UncompressedBytes int
	CompressedBytes   int
	Duration          time.Duration
}

func (s CompressionStats) Ratio() float64 {
	if s.UncompressedBytes <= 0 {
		return 0
	}
	return float64(s.CompressedBytes) / float64(s.UncompressedBytes)
}

func (s CompressionStats) SavedPercent() float64 {
	if s.UncompressedBytes <= 0 {
		return 0
	}
	return (1 - s.Ratio()) * 100
}

type ChecksumAlgorithm string

const (
	ChecksumAlgorithmSHA256 ChecksumAlgorithm = "sha256"
)

type Checksum struct {
	Algorithm ChecksumAlgorithm
	Value     string
}

func (c Checksum) IsZero() bool {
	return c.Algorithm == "" && strings.TrimSpace(c.Value) == ""
}
