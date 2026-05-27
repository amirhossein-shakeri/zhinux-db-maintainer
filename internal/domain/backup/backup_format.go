package backup

type BackupFormat string

const (
	BackupFormatPlainSQL  BackupFormat = "plain_sql"
	BackupFormatCustom    BackupFormat = "custom"
	BackupFormatTar       BackupFormat = "tar"
	BackupFormatDirectory BackupFormat = "directory"
)
