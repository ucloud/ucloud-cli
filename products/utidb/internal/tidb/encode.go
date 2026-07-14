package tidb

import (
	"strings"

	"github.com/ucloud/ucloud-sdk-go/services/tidb"
)

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
