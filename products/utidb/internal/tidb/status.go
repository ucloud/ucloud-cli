package tidb

// UTiDB domain state constants, product-owned copies.
const (
	stateAvailable   = "Available"
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
	stateAltering    = "Altering"  // modify-spec / scale / resize in progress (实测)
	stateDeploying   = "Deploying" // create in progress (实测)
)
