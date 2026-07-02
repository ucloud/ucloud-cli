package uhost

// UHost-domain state/type constants plus constants this product depends on,
// product-owned copies (formerly model/status + model/cli; domain constants
// live with the product). REGEXP_NAME is the resource-name validation pattern
// (formerly model/cli).
const (
	HOST_RUNNING = "Running"
	HOST_STOPPED = "Stopped"
	HOST_FAIL    = "Install Fail"

	IMAGE_AVAILABLE   = "Available"
	IMAGE_UNAVAILABLE = "Unavailable"

	DISK_AVAILABLE = "Available"
	DISK_FAILED    = "Failed"

	EIP_FREE = "free"

	IMAGE_BASE  = "Base"
	IMAGE_ALL   = "*"
	REGEXP_NAME = "^[A-Za-z0-9-_.一-龥]{1,63}$"
)
