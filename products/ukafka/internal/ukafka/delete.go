package ukafka

import (
	"fmt"

	"github.com/spf13/cobra"

	ukafkasdk "github.com/ucloud/ucloud-sdk-go/services/ukafka"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newDelete ucloud ukafka delete
func newDelete(ctx *cli.Context) *cobra.Command {
	var yes *bool
	var instanceIDs *[]string
	client := cli.NewServiceClient(ctx, ukafkasdk.NewClient)
	req := client.NewDeleteUKafkaInstanceRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete UKafka instances",
		Long:  "Delete UKafka instances",
		Run: func(cmd *cobra.Command, args []string) {
			ok, err := ctx.Confirm(*yes, "Are you sure to delete UKafka instance(s)?")
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, idName := range *instanceIDs {
				id := ctx.PickResourceID(idName)
				req.InstanceId = &id
				_, err := client.DeleteUKafkaInstance(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(w, "ukafka[%s] deleted\n", id)
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "delete", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	instanceIDs = flags.StringSlice("ukafka-id", nil, "Required. Instance ID(s) to delete")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.Zone = flags.String("zone", ctx.DefaultZone(), "Optional. Assign availability zone")
	yes = flags.BoolP("yes", "y", false, "Optional. Skip confirmation prompt")

	cmd.MarkFlagRequired("ukafka-id")

	return cmd
}
