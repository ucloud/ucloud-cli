package ugn

import (
	"github.com/spf13/cobra"

	ugnsdk "github.com/ucloud/ucloud-sdk-go/services/ugn"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// bwRow is the table row for `ucloud ugn bw list`.
type bwRow struct {
	ResourceID    string
	Name          string
	BandwidthMbps float64
	RegionA       string
	RegionB       string
	Path          string
	QoS           string
	ChargeType    string
	CreateTime    string
	ExpireTime    string
}

// newBWList ucloud ugn bw list
func newBWList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ugnsdk.NewClient)
	req := client.NewGetSimpleUGNBwPackagesRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List ugn bandwidth packages",
		Long:  "List ugn bandwidth packages",
		Run: func(c *cobra.Command, args []string) {
			*req.UGNID = ctx.PickResourceID(*req.UGNID)
			resp, err := client.GetSimpleUGNBwPackages(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			rows := make([]bwRow, 0, len(resp.BwPackages))
			for _, bw := range resp.BwPackages {
				path := bw.Path
				if path == "None" {
					path = "IGP"
				}
				expireTime := ""
				if bw.ExpireTime > 0 {
					expireTime = common.FormatDate(bw.ExpireTime)
				}
				rows = append(rows, bwRow{
					ResourceID:    bw.PackageID,
					Name:          bw.Name,
					BandwidthMbps: bw.BandWidth,
					RegionA:       bw.RegionA,
					RegionB:       bw.RegionB,
					Path:          path,
					QoS:           bw.Qos,
					ChargeType:    bw.PayMode,
					CreateTime:    common.FormatDate(bw.CreateTime),
					ExpireTime:    expireTime,
				})
			}
			ctx.PrintList(rows)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.UGNID = flags.String("ugn-id", "", "Required. Resource ID of the ugn instance")
	ctx.BindOffset(cmd, req)
	req.Limit = flags.Int("limit", 50, "Optional. Limit")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Project-id, see 'ucloud project list'")

	cmd.MarkFlagRequired("ugn-id")
	ctx.SetCompletion(cmd, "ugn-id", func() []string {
		return getAllUGNIdNames(ctx, *req.ProjectId)
	})
	ctx.SetCompletion(cmd, "project-id", ctx.ProjectList)

	return cmd
}
