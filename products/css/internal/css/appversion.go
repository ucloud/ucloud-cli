package css

import (
	"fmt"

	"github.com/spf13/cobra"

	uessdk "github.com/ucloud/ucloud-sdk-go/services/ues"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newAppVersion ucloud css app-version
func newAppVersion(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uessdk.NewClient)
	req := client.NewGetUESAppVersionRequest()
	cmd := &cobra.Command{
		Use:   "app-version",
		Short: "List available UES application versions",
		Long:  "List available UES application versions",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := client.GetUESAppVersion(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []AppVersionRow{}
			for _, v := range resp.AppVersionList {
				list = append(list, AppVersionRow{
					AppName:     v.AppName,
					AppVersion:  v.AppVersion,
					IsMultiZone: fmt.Sprintf("%t", v.IsMultiZone),
				})
			}
			ctx.PrintList(list)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	return cmd
}
