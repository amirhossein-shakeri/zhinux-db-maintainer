package backup

type BackupTrigger string

const (
	BackupTriggerManual    BackupTrigger = "manual"
	BackupTriggerScheduled BackupTrigger = "scheduled"
)
