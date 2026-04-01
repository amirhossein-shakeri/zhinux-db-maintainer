package restore

type RestoreStatus string

const (
	RestoreStatusPending    RestoreStatus = "pending"
	RestoreStatusInProgress RestoreStatus = "in_progress"
	RestoreStatusSuccess    RestoreStatus = "success" // todo: Feelds inconsistent?
	RestoreStatusFailed     RestoreStatus = "failed"
	RestoreStatusCanceled   RestoreStatus = "canceled"
)
