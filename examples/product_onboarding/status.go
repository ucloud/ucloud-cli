package onboarding

// Terminal states a Poller waits on. A real product imports these from its own
// status table; the example defines them locally so it depends only on the
// platform packages and the SDK, never on another product's internals.
const (
	stateRunning = "Running"
	stateShutoff = "Shutoff"
	stateFail    = "Fail"
)
