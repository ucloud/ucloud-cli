package redis

type Row struct {
	ResourceID string
	Name       string
	Role       string
	Type       string
	Address    string
	Size       string
	UsedSize   string
	State      string
	Group      string
	Zone       string
	CreateTime string
}

var redisTypeMap = map[string]string{
	"single":      "master-replica",
	"distributed": "distributed",
}
