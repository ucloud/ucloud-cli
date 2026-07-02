package mysql

// MachineTypeRow 计算规格表格行
type MachineTypeRow struct {
	ID          string
	Description string
	Cpu         int
	Memory      int
	Group       string
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
