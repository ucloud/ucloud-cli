package topic

// topicRow is the full-field row (json/yaml mode). Fields correspond to SDK TopicInfo.
type topicRow struct {
	TopicId     string
	TopicName   string
	MessageType string
	Remark      string
	CreateTime  int
}

// topicRowDefault is the default curated columns in table mode: TopicName/MessageType/Remark/CreateTime.
type topicRowDefault struct {
	TopicName   string
	MessageType string
	Remark      string
	CreateTime  string
}
