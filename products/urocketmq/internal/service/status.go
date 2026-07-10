package service

// Service status constants, from ServiceBaseInfo.State / ServiceDetail.State enum values.
// See uhost status.go (HOST_RUNNING / HOST_STOPPED / HOST_FAIL).
const (
	SERVICE_AVAILABLE      = "Available"
	SERVICE_INITIALIZING   = "Initializing"
	SERVICE_DELETING       = "Deleting"
	SERVICE_CREATE_FAILED  = "CreateFailed"
	SERVICE_CLOSING        = "Closing"
	SERVICE_CLOSED         = "Closed"
	SERVICE_CLOSE_FAILED   = "CloseFailed"
	SERVICE_RECOVERING     = "Recovering"
	SERVICE_RECOVER_FAILED = "RecoverFailed"
	SERVICE_UPGRADING      = "Upgrading"
	SERVICE_UPGRADE_FAILED = "UpgradeFailed"
	SERVICE_DELETE_FAILED  = "DeleteFailed"
	// SERVICE_DELETED is a pseudo-status used internally by deletion polling: Get returns empty list
	// indicating the instance is gone.
	SERVICE_DELETED = "Deleted"
)
