package ulhost

// ULHost-domain state/type constants. Product-owned copies (mirrors uhost
// status.go pattern). ULHost states follow the ucompshare SDK
// ULHostInstanceSet.State enumeration.
const (
	HOST_RUNNING      = "Running"
	HOST_STOPPED      = "Stopped"
	HOST_FAIL         = "Install Fail"
	HOST_INITIALIZING = "Initializing"
	HOST_STARTING     = "Starting"
	HOST_STOPPING     = "Stopping"
	HOST_REBOOTING    = "Rebooting"

	IMAGE_AVAILABLE   = "Available"
	IMAGE_UNAVAILABLE = "Unavailable"

	REGEXP_NAME = "^[A-Za-z0-9-_.一-龥]{1,63}$"
)
