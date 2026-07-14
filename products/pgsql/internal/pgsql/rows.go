package pgsql

// PgsqlInstanceRow is the table row for `pgsql db list`.
type PgsqlInstanceRow struct {
	Name         string
	InstanceID   string
	Zone         string
	State        string
	IP           string
	VPC          string
	Subnet       string
	InstanceMode string
	DBVersion    string
	Port         int
	DiskSpace    int
	Memory       int
}

// PgsqlMachineTypeRow is the table row for `pgsql db list-machine-type`.
type PgsqlMachineTypeRow struct {
	ID          string
	Description string
	Cpu         int
	Memory      int
	Os          string
}

// PgsqlVersionRow is the table row for `pgsql db list-version`.
type PgsqlVersionRow struct {
	DBVersion string
	Available string
}

// PgsqlConfRow is the table row for `pgsql conf list`.
type PgsqlConfRow struct {
	GroupID     int
	GroupName   string
	DBVersion   string
	Description string
	Modifiable  bool
}

// PgsqlConfParamRow is the table row for `pgsql conf describe` parameter list.
type PgsqlConfParamRow struct {
	Key        string
	Value      string
	Modifiable bool
}

// PgsqlBackupRow is the table row for `pgsql backup list`.
type PgsqlBackupRow struct {
	BackupID        string
	BackupName      string
	InstanceID      string
	State           string
	BackupType      string
	BackupSize      string
	BackupStartTime string
	BackupEndTime   string
}

// PgsqlBackupURLRow is the table row for `pgsql backup download`.
type PgsqlBackupURLRow struct {
	BackupPath      string
	InnerBackupPath string
}

// PgsqlBackupStrategyRow is the table row for `pgsql backup strategy`.
type PgsqlBackupStrategyRow struct {
	BackupMethod    string
	BackupTimeRange string
	BackupWeek      string
}

// PgsqlLogRow is the table row for `pgsql log list`.
type PgsqlLogRow struct {
	Name      string
	Size      string
	BeginTime string
	EndTime   string
}

// PgsqlPriceRow is the table row for `pgsql db price` / `upgrade-price`.
type PgsqlPriceRow struct {
	ChargeType    string
	Price         float64
	OriginalPrice float64
}
