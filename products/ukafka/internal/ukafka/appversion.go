package ukafka

import (
	"github.com/spf13/cobra"

	ukafkasdk "github.com/ucloud/ucloud-sdk-go/services/ukafka"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newAppVersion ucloud ukafka app-version
func newAppVersion(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ukafkasdk.NewClient)
	req := client.NewListUKafkaFrameworkVersionRequest()
	cmd := &cobra.Command{
		Use:   "app-version",
		Short: "List available Kafka versions",
		Long:  "List available Kafka versions",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := client.ListUKafkaFrameworkVersion(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := []VersionRow{}
			for _, v := range resp.FrameworkVersions {
				row := VersionRow{
					Version: v.Version,
					Label:   v.Label,
				}
				list = append(list, row)
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
