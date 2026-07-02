package image

// Image-domain state/type constants plus constants this product depends on,
// product-owned copies (formerly model/status + model/cli; the IAMGE_CUSTOM
// typo is preserved verbatim — renaming it is behavior-adjacent cleanup, out
// of scope for the pure copy).
const (
	HOST_RUNNING = "Running"
	HOST_STOPPED = "Stopped"

	IMAGE_MAKING      = "Making"
	IMAGE_AVAILABLE   = "Available"
	IMAGE_UNAVAILABLE = "Unavailable"
	IMAGE_COPYING     = "Copying"

	IAMGE_CUSTOM = "Custom"
	IMAGE_ALL    = "*"
)
