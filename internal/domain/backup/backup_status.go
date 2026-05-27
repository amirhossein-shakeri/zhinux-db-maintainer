package backup

type BackupStatus string

const (
	BackupStatusPending    BackupStatus = "pending"
	BackupStatusInProgress BackupStatus = "in_progress"
	BackupStatusSucceeded  BackupStatus = "success"
	BackupStatusSuccess    BackupStatus = BackupStatusSucceeded
	BackupStatusFailed     BackupStatus = "failed"
	BackupStatusCanceled   BackupStatus = "canceled"
)

func (s BackupStatus) IsTerminal() bool {
	return s == BackupStatusSucceeded ||
		s == BackupStatusSuccess ||
		s == BackupStatusFailed ||
		s == BackupStatusCanceled
}

func CanTransitionBackupStatus(from BackupStatus, to BackupStatus) bool {
	switch from {
	case "":
		return to == BackupStatusPending
	case BackupStatusPending:
		return to == BackupStatusInProgress || to == BackupStatusCanceled || to == BackupStatusFailed
	case BackupStatusInProgress:
		return to == BackupStatusSucceeded || to == BackupStatusSuccess || to == BackupStatusFailed || to == BackupStatusCanceled
	default:
		return false
	}
}
