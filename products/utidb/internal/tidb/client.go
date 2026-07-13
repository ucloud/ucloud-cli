package tidb

import (
	"fmt"
	"strings"

	"github.com/ucloud/ucloud-sdk-go/services/tidb"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// invokeAPI calls a TiDB API action via GenericRequest.
// String arrays like NodeTypes use flat keys (NodeTypes0). Nested objects like NodeConfig
// should be passed as maps/slices in params so the SDK encoder emits NodeConfig.0.Field.
func invokeAPI(ctx *cli.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
	client := cli.NewServiceClient(ctx, tidb.NewClient)
	req := client.NewGenericRequest()
	allParams := map[string]interface{}{
		"Action": action,
	}
	for k, v := range params {
		allParams[k] = v
	}
	if err := req.SetPayload(allParams); err != nil {
		return nil, fmt.Errorf("set payload: %w", err)
	}
	resp, err := client.GenericInvoke(req)
	if err != nil {
		return nil, err
	}
	return resp.GetPayload(), nil
}

func mergeCommonParams(region, zone, projectID string, params map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(params)+3)
	for k, v := range params {
		out[k] = v
	}
	if region != "" {
		out["Region"] = region
	}
	if zone != "" {
		out["Zone"] = zone
	}
	if projectID != "" {
		out["ProjectId"] = projectID
	}
	return out
}

func flattenIndexedStrings(params map[string]interface{}, name string, values []string) {
	for i, v := range values {
		params[fmt.Sprintf("%s%d", name, i)] = v
	}
}

func formatNodeType(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case "tidb":
		return "Tidb"
	case "tikv":
		return "Tikv"
	case "pd":
		return "Pd"
	case "tiflash":
		return "Tiflash"
	default:
		if s == "" {
			return s
		}
		return strings.ToUpper(s[:1]) + s[1:]
	}
}

func formatNodeTypes(types []string) []string {
	out := make([]string, len(types))
	for i, t := range types {
		out[i] = formatNodeType(t)
	}
	return out
}

func getTiDBClusterUhostSpecs(ctx *cli.Context, region, zone, projectID string, nodeTypes []string) ([]tidb.UhostSpecs, error) {
	params := mergeCommonParams(region, zone, projectID, map[string]interface{}{})
	flattenIndexedStrings(params, "NodeTypes", formatNodeTypes(nodeTypes))
	payload, err := invokeAPI(ctx, "GetTiDBClusterUhostSpecs", params)
	if err != nil {
		return nil, err
	}
	return parseUhostSpecsFromPayload(payload), nil
}

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

func createNodeConfigToMap(cfg tidb.CreateTiDBClusterServiceParamNodeConfig) map[string]interface{} {
	m := map[string]interface{}{}
	if cfg.ConfigId != nil {
		m["ConfigId"] = *cfg.ConfigId
	}
	if cfg.DiskSize != nil {
		m["DiskSize"] = *cfg.DiskSize
	}
	if cfg.NodeCount != nil {
		m["NodeCount"] = *cfg.NodeCount
	}
	if cfg.ServerType != nil {
		m["ServerType"] = formatNodeType(*cfg.ServerType)
	}
	return m
}

func labelToMap(l tidb.CreateTiDBClusterServiceParamLabels) map[string]interface{} {
	m := map[string]interface{}{}
	if l.Key != nil {
		m["Key"] = *l.Key
	}
	if l.Value != nil {
		m["Value"] = *l.Value
	}
	return m
}

func secGroupToMap(s tidb.CreateTiDBClusterServiceParamSecGroupInfo) map[string]interface{} {
	m := map[string]interface{}{}
	if s.SecGroupId != nil {
		m["SecGroupId"] = *s.SecGroupId
	}
	if s.Priority != nil {
		m["Priority"] = *s.Priority
	}
	return m
}

func scaleNodeConfigToMap(cfg tidb.ModifyTiDBClusterNodeParamNodeConfig) map[string]interface{} {
	m := map[string]interface{}{}
	if cfg.ConfigId != nil {
		m["ConfigId"] = *cfg.ConfigId
	}
	if cfg.NodeCount != nil {
		m["NodeCount"] = *cfg.NodeCount
	}
	if cfg.ServerType != nil {
		m["ServerType"] = formatNodeType(*cfg.ServerType)
	}
	return m
}

func resizeDiskNodeConfigToMap(cfg tidb.ModifyTiDBClusterUhostDiskParamNodeConfig) map[string]interface{} {
	m := map[string]interface{}{}
	if cfg.DiskSize != nil {
		m["DiskSize"] = *cfg.DiskSize
	}
	if cfg.ServerType != nil {
		m["ServerType"] = formatNodeType(*cfg.ServerType)
	}
	return m
}

func modifySpecNodeConfigToMap(cfg tidb.ModifyTiDBClusterUhostSpecsParamNodeConfig) map[string]interface{} {
	m := map[string]interface{}{}
	if cfg.ConfigId != nil {
		m["ConfigId"] = *cfg.ConfigId
	}
	if cfg.ServerType != nil {
		m["ServerType"] = formatNodeType(*cfg.ServerType)
	}
	return m
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

func stringVal(v interface{}) string {
	s, _ := v.(string)
	return s
}

func getTiDBClusterPayload(ctx *cli.Context, region, zone, projectID, id string) (map[string]interface{}, error) {
	params := mergeCommonParams(region, zone, projectID, map[string]interface{}{
		"Id": id,
	})
	return invokeAPI(ctx, "GetTiDBClusterService", params)
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
