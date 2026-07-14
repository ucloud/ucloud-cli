package ukafka

// InstanceRow represents a UKafka instance in list output
type InstanceRow struct {
	InstanceID   string
	InstanceName string
	Framework    string
	Version      string
	Zone         string
	State        string
	NodeCount    string
	VPCId        string
	SubnetId     string
	ChargeType   string
	CreateTime   string
	ExpireTime   string
}

// NodeConfRow represents a node configuration in node-conf output
type NodeConfRow struct {
	NodeType    string
	CPU         string
	Memory      string
	DiskType    string
	MinDiskSize string
	MaxDiskSize string
	SecGroup    string
}

// VersionRow represents a version in app-version output
type VersionRow struct {
	Version string
	Label   string
}

// ConsumerGroupRow represents a consumer group in list-consumers output
type ConsumerGroupRow struct {
	GroupName   string
	Type        string
	NumOfTopics string
	GroupID     string
}

// TopicRow represents a topic in list-topics output
type TopicRow struct {
	Topic             string
	NumOfPartition    string
	NumOfReplica      string
	NumOfOccupyBroker string
	UnderReplicasPer  string
	Status            string
}
