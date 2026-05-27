package backup

import (
	"fmt"
	"strings"
)

type StorageProvider string

const (
	StorageProviderLocal StorageProvider = "local"
	StorageProviderS3    StorageProvider = "s3"
)

type StorageLocation struct {
	Provider StorageProvider
	Bucket   string
	Key      string
	Region   string
	URI      string
}

func (l StorageLocation) String() string {
	if strings.TrimSpace(l.URI) != "" {
		return l.URI
	}
	if l.Provider == StorageProviderS3 {
		return fmt.Sprintf("s3://%s/%s", l.Bucket, strings.TrimPrefix(l.Key, "/"))
	}
	return l.Key
}

func StorageLocationFromURI(uri string) StorageLocation {
	return StorageLocation{URI: strings.TrimSpace(uri)}
}
