package tidb

// UTiDB domain state constants used by pollers / EmitResult.
// Terminal states validated on prod (cn-bj2): Available after create/scale/resize;
// delete removes the instance (poll may hit Deleted or describe error depending on timing).
const (
	stateAvailable   = "Available"
	stateRunning     = "Running"
	stateDeleted     = "Deleted"
	stateCreateFail  = "CreateFailed"
	stateDeleteFail  = "DeleteFailed"
	stateBackingUp   = "BackingUp"
	stateUpgradeFail = "UpgradeFailed"
)
