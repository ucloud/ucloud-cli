package uhost

// State is the state of UHost instance
type State = string

// Enum values for State
const (
	StateInitializing State = "Initializing"
	StateStarting     State = "Starting"
	StateRunning      State = "Running"
	StateStopping     State = "Stopping"
	StateStopped      State = "Stopped"
	StateInstallFail  State = "InstallFail"
	StateRebooting    State = "Rebooting"
)
