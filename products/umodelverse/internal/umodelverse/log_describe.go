package umodelverse

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func newLogDescribe(ctx *cli.Context) *cobra.Command {
	client := newClient(ctx)
	req := &logDetailRequest{}
	newRequest(client, req, true)

	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe a uModelVerse inference request log",
		Long:  "Describe a uModelVerse inference request log by request ID.",
		Run: func(c *cobra.Command, args []string) {
			resp, err := invokeUMAction(client, "GetUMInferRequestLogDetail", req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			printResponse(ctx, resp)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	bindProject(cmd, req, ctx.DefaultProjectID())
	req.RequestId = flags.String("request-id", "", "Required. Request ID.")
	cmd.MarkFlagRequired("region")
	cmd.MarkFlagRequired("zone")
	cmd.MarkFlagRequired("request-id")
	return cmd
}
