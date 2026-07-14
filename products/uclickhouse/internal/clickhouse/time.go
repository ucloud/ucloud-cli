package clickhouse

import "github.com/ucloud/ucloud-cli/internal/common"

func formatUnixDate(timestamp int) string {
	if timestamp <= 0 {
		return ""
	}
	if timestamp > 1000000000000 {
		timestamp = timestamp / 1000
	}
	return common.FormatDate(timestamp)
}
