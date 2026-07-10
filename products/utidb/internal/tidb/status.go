package tidb

// UTiDB domain state constants, product-owned copies.
const (
	stateRunning     = "Running"
	stateStopped     = "Stopped"
	stateFailed      = "Failed"
	stateCreating    = "Creating"
	stateDeleting    = "Deleting"
	stateDeleted     = "Deleted"
	stateCreateFail  = "CreateFailed"
	stateDeleteFail  = "DeleteFailed"
	stateBackingUp   = "BackingUp"
	stateBackupFail  = "BackupFailed"
	stateUpgrading   = "Upgrading"
	stateUpgradeFail = "UpgradeFailed"
)
