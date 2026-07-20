package ugn

import "github.com/ucloud/ucloud-cli/internal/common"

// UGNRow is the table row for `ucloud ugn list`.
type UGNRow struct {
	ResourceID     string
	Name           string
	Remark         string
	NetworkCount   int
	BwPackageCount int
	CreateTime     string
}

func formatUGNCreateTime(t int) string {
	return common.FormatDate(t)
}
