package mysql

// MachineTypeRow 计算规格表格行
type MachineTypeRow struct {
	ID          string
	Description string
	Cpu         int
	Memory      int
	Group       string
}

// OpResultRow 写命令(create/delete/start/stop/...)的结构化结果行。
// 仅在 --output json/yaml 下输出(见 output.go emitResult),供脚本/agent 提取
// 操作的资源 id 与状态。字段不加 json tag,与本包其它 *Row 风格保持一致
// (键名即 Go 字段名)。
type OpResultRow struct {
	ResourceID string
	Action     string
	Status     string
}

// UDBMysqlRow 表格行
type UDBMysqlRow struct {
	Name       string
	ResourceID string
	Role       string
	Status     string
	Config     string
	Mode       string
	DiskType   string
	IP         string
	Group      string
	Zone       string
	VPC        string
	Subnet     string
	// CreateTime string
}
