package tidb

import (
	"fmt"
	"strings"
)

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
