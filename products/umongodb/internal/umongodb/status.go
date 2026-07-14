package umongodb

// Terminal states for the Poller. UMongoDB uses "Stopped" (not UDB's "Shutoff").
const (
	stateRunning = "Running"
	stateStopped = "Stopped"
	stateFail    = "InitFailed"
)
