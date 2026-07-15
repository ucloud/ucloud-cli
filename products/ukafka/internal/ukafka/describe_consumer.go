package ukafka

import (
	"fmt"

	"github.com/spf13/cobra"

	ukafkasdk "github.com/ucloud/ucloud-sdk-go/services/ukafka"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newDescribeConsumer ucloud ukafka describe-consumer
func newDescribeConsumer(ctx *cli.Context) *cobra.Command {
	var instanceID *string
	var consumerGroup *string
	var consumerType *string

	client := cli.NewServiceClient(ctx, ukafkasdk.NewClient)
	req := client.NewDescribeUKafkaConsumerRequest()

	cmd := &cobra.Command{
		Use:   "describe-consumer",
		Short: "Describe Kafka consumer group details",
		Long:  "Describe Kafka consumer group details",
		Run: func(cmd *cobra.Command, args []string) {
			req.ClusterInstanceId = sdk.String(*instanceID)
			req.ConsumerGroup = sdk.String(*consumerGroup)
			req.Type = sdk.String(*consumerType)

			resp, err := client.DescribeUKafkaConsumer(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			rows := []cli.DescribeRow{
				{Attribute: "GroupName", Content: resp.GroupName},
				{Attribute: "Type", Content: resp.Type},
			}
			if len(resp.Topics) > 0 {
				for i, topic := range resp.Topics {
					rows = append(rows, cli.DescribeRow{
						Attribute: fmt.Sprintf("Topic[%d]", i),
						Content:   topic,
					})
				}
			}
			ctx.PrintList(rows)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	instanceID = flags.String("ukafka-id", "", "Required. Instance ID")
	consumerGroup = flags.String("consumer-group", "", "Required. Consumer group name")
	consumerType = flags.String("type", "", "Required. Consumer group type (e.g. ZK, KF)")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")

	cmd.MarkFlagRequired("ukafka-id")
	cmd.MarkFlagRequired("consumer-group")
	cmd.MarkFlagRequired("type")

	return cmd
}
