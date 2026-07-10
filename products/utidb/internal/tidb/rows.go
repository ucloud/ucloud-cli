package tidb

import (
	"github.com/ucloud/ucloud-sdk-go/services/tidb"
)

// instanceRow is the table row for a UTiDB instance.
type instanceRow struct {
	ID         string
	Name       string
	State      string
	DTType     int
	Port       int
	IP         string
	Version    string
	CreateTime int
	VPCID      string
	SubnetID   string
}

func newInstanceRowFromData(d tidb.UTiDBServiceData) instanceRow {
	return instanceRow{
		ID:         d.Id,
		Name:       d.Name,
		State:      d.State,
		DTType:     d.DTType,
		Port:       d.Port,
		IP:         d.Ip,
		Version:    d.Version,
		CreateTime: d.CreateTime,
		VPCID:      d.VPCId,
		SubnetID:   d.SubnetId,
	}
}

// backupRow is the table row for a UTiDB backup.
type backupRow struct {
	BackupID        string
	BackupType      string
	State           string
	BackupSize      int
	BackupStartTime int
	BackupEndTime   int
}

func newBackupRowFromData(d tidb.BackupData) backupRow {
	return backupRow{
		BackupID:        d.BackupId,
		BackupType:      d.BackupType,
		State:           d.State,
		BackupSize:      d.BackupSize,
		BackupStartTime: d.BackupStartTime,
		BackupEndTime:   d.BackupEndTime,
	}
}

// specRow is the table row for a UTiDB uhost spec.
type specRow struct {
	ConfigID        string
	ConfigName      string
	NodeType        string
	CoreNum         int
	Memory          int
	MinDiskCapacity int
	MaxDiskCapacity int
	DiskStep        int
}

func newSpecRowFromData(d tidb.UhostSpecs) specRow {
	return specRow{
		ConfigID:        d.ConfigId,
		ConfigName:      d.ConfigName,
		NodeType:        d.NodeType,
		CoreNum:         d.CoreNum,
		Memory:          d.Memory,
		MinDiskCapacity: d.MinDiskCapacity,
		MaxDiskCapacity: d.MaxDiskCapacity,
		DiskStep:        d.DiskStep,
	}
}
