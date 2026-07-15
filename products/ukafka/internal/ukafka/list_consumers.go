package ukafka

import (
	"fmt"

	"github.com/spf13/cobra"

	ukafkasdk "github.com/ucloud/ucloud-sdk-go/services/ukafka"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newListConsumers ucloud ukafka list-consumers
func newListConsumers(ctx *cli.Context) *cobra.Command {
	var instanceID *string

	client := cli.NewServiceClient(ctx, ukafkasdk.NewClient)
	req := client.NewListUKafkaConsumersRequest()

	cmd := &cobra.Command{
		Use:   "list-consumers",
		Short: "List Kafka consumer groups",
		Long:  "List Kafka consumer groups in UKafka instance",
		Run: func(cmd *cobra.Command, args []string) {
			req.ClusterInstanceId = sdk.String(*instanceID)

			resp, err := client.ListUKafkaConsumers(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			list := []ConsumerGroupRow{}
			for _, g := range resp.Groups {
				row := ConsumerGroupRow{
					GroupName:   g.GroupName,
					Type:        g.Type,
					NumOfTopics: fmt.Sprintf("%d", g.NumOfTopics),
					GroupID:     g.GroupId,
				}
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	instanceID = flags.String("ukafka-id", "", "Required. Instance ID")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")

	cmd.MarkFlagRequired("ukafka-id")

	return cmd
}
