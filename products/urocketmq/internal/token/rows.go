package token

// tokenRow is the full-field row (json/yaml mode), without SecretKey (secure default).
type tokenRow struct {
	TokenId          string
	Name             string
	TopicConsumePerm string
	TopicProducePerm string
	Type             string
	CreateTime       string
	ModifyTime       string
	AccessKey        string
}

// tokenRowWithSecret is used for get --display, appending SecretKey on top of tokenRow.
type tokenRowWithSecret struct {
	TokenId          string
	Name             string
	TopicConsumePerm string
	TopicProducePerm string
	Type             string
	CreateTime       string
	ModifyTime       string
	AccessKey        string
	SecretKey        string
}

// tokenRowDefault is the curated columns in table mode, containing no AKSK key information.
type tokenRowDefault struct {
	TokenId          string
	Name             string
	TopicConsumePerm string
	TopicProducePerm string
	Type             string
	CreateTime       string
}
