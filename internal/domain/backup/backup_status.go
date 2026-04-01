package backup

type BackupStatus string

const (
	BackupStatusPending    BackupStatus = "pending"
	BackupStatusInProgress BackupStatus = "in_progress"
	BackupStatusSuccess    BackupStatus = "success" // todo: Feelds inconsistent?
	BackupStatusFailed     BackupStatus = "failed"
	BackupStatusCanceled   BackupStatus = "canceled"
)
