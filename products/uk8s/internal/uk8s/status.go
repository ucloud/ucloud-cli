package uk8s

// UK8S cluster-domain state constants. Sourced from
// ucloud-sdk-go/services/uk8s/models.go UK8SClusterSet.Status docstring.
// Product-owned (formerly model/status); see §2.5 of the platform spec.
const (
	// Cluster is being initialized after creation.
	CLUSTER_INITIALIZING = "INITIALIZING"
	// Cluster is starting up.
	CLUSTER_STARTING = "STARTING"
	// Cluster creation failed.
	CLUSTER_CREATEFAILED = "CREATEFAILED"
	// Cluster is running normally.
	CLUSTER_RUNNING = "RUNNING"
	// Cluster has an error.
	CLUSTER_ERROR = "ERROR"
	// Cluster is in an abnormal state.
	CLUSTER_ABNORMAL = "ABNORMAL"
)