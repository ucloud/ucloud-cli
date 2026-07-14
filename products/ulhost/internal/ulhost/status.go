package ulhost

// ULHost-domain state/type constants. Product-owned copies (mirrors uhost
// status.go pattern). ULHost states follow the ucompshare SDK
// ULHostInstanceSet.State enumeration. Only the states the CLI polls against
// (terminal Running/Stopped/Install Fail) are kept; the SDK also reports
// Initializing/Starting/Stopping/Rebooting, but no ulhost command waits on
// those, so they are omitted to avoid unused-constant drift.
const (
	HOST_RUNNING = "Running"
	HOST_STOPPED = "Stopped"
	HOST_FAIL    = "Install Fail"

	REGEXP_NAME = "^[A-Za-z0-9-_.一-龥]{1,63}$"
)
