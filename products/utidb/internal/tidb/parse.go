package tidb

import (
	"fmt"
	"strings"

	"github.com/ucloud/ucloud-sdk-go/services/tidb"
)

func parseUhostSpecsFromPayload(payload map[string]interface{}) []tidb.UhostSpecs {
	data, ok := payload["Data"].([]interface{})
	if !ok {
		return nil
	}
	specs := make([]tidb.UhostSpecs, 0, len(data))
	for _, item := range data {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		specs = append(specs, tidb.UhostSpecs{
			ConfigId:        stringVal(m["ConfigId"]),
			ConfigName:      stringVal(m["ConfigName"]),
			NodeType:        stringVal(m["NodeType"]),
			CoreNum:         intVal(m["CoreNum"]),
			Memory:          intVal(m["Memory"]),
			MinDiskCapacity: intVal(m["MinDiskCapacity"]),
			MaxDiskCapacity: intVal(m["MaxDiskCapacity"]),
			DiskStep:        intVal(m["DiskStep"]),
		})
	}
	return specs
}

func parseBackupRowsFromPayload(payload map[string]interface{}) []backupRow {
	data, ok := payload["Data"].([]interface{})
	if !ok {
		return nil
	}
	rows := make([]backupRow, 0, len(data))
	for _, item := range data {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		rows = append(rows, backupRow{
			BackupID:        stringVal(m["BackupId"]),
			BackupType:      stringVal(m["BackupType"]),
			State:           stringVal(m["State"]),
			BackupSize:      intVal(m["BackupSize"]),
			BackupStartTime: intVal(m["BackupStartTime"]),
			BackupEndTime:   intVal(m["BackupEndTime"]),
		})
	}
	return rows
}

var clusterServerKeys = []string{"TiDBServers", "TiKVServers", "PDServers", "TiFlashServers"}

func extractServerIDs(payload map[string]interface{}) []string {
	data, _ := payload["Data"].(map[string]interface{})
	if data == nil {
		return nil
	}
	cluster, _ := data["TiDBCluster"].(map[string]interface{})
	if cluster == nil {
		cluster = data
	}
	var ids []string
	for _, key := range clusterServerKeys {
		servers, _ := cluster[key].([]interface{})
		for _, item := range servers {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			serverID := stringVal(m["ServerId"])
			if serverID == "" {
				continue
			}
			host := stringVal(m["HostIp"])
			nodeType := strings.TrimSuffix(key, "Servers")
			if host != "" {
				ids = append(ids, fmt.Sprintf("%s/%s@%s", serverID, strings.ToLower(nodeType), host))
			} else {
				ids = append(ids, fmt.Sprintf("%s/%s", serverID, strings.ToLower(nodeType)))
			}
		}
	}
	return ids
}

func stringVal(v interface{}) string {
	s, _ := v.(string)
	return s
}

func intVal(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	default:
		return 0
	}
}
