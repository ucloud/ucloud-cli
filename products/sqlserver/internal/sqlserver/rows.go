package sqlserver

// MachineTypeRow 计算规格表格行
type MachineTypeRow struct {
	ID          string
	Description string
	Cpu         int
	Memory      int
	Group       string
}

// UDBSQLServerRow 表格行
type UDBSQLServerRow struct {
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
}
