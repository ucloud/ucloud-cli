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

type BlockRow struct {
	BlockID     string
	BlockName   string
	BlockVip    string
	BlockPort   int
	BlockType   string
	BlockState  string
	BlockSize   int
	UsedSize    int
	SlotBegin   int
	SlotEnd     int
	ReadWeight  int
}

type ProxyRow struct {
	ProxyID    string
	ResourceID string
	State      string
	Vip        string
}

var redisTypeMap = map[string]string{
	"single":      "master-replica",
	"distributed": "distributed",
}
