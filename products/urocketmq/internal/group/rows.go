package group

// groupRow is the full-field row (json/yaml mode).
type groupRow struct {
	GroupName  string
	Id         string
	Remark     string
	CreateTime int
}

// groupRowDefault is the default curated columns in table mode: GroupName, Id, Remark, CreateTime.
type groupRowDefault struct {
	GroupName  string
	Id         string
	Remark     string
	CreateTime string
}
