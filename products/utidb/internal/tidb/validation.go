package tidb

import (
	"fmt"
	"strings"
)

// minNodeCountPerType is the minimum NodeCount per node type enforced by the TiDB API (> 3).
const minNodeCountPerType = 4

// serverTypeValues lists accepted ServerType values (CLI input, case-insensitive).
var serverTypeValues = []string{"tidb", "tikv", "pd", "tiflash"}

func helpServerTypes() string {
	return strings.Join(serverTypeValues, ", ")
}

func validateServerType(serverType string) error {
	st := strings.ToLower(strings.TrimSpace(serverType))
	for _, v := range serverTypeValues {
		if st == v {
			return nil
		}
	}
	return fmt.Errorf("invalid ServerType %q: must be one of %s", serverType, helpServerTypes())
}

func validateNodeCount(count int, serverType string) error {
	if count < minNodeCountPerType {
		return fmt.Errorf(
			"NodeCount for %s must be greater than 3 (minimum %d), got %d; each node type requires more than 3 nodes (各节点类型数量须大于 3，最少 %d 个)",
			serverType, minNodeCountPerType, count, minNodeCountPerType,
		)
	}
	return nil
}
