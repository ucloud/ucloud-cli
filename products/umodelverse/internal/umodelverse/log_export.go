package umodelverse

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func newLogExport(ctx *cli.Context) *cobra.Command {
	client := newClient(ctx)
	req := &requestLogRequest{}
	newRequest(client, req, false)
	var modelNames []string
	var apiKeyIds []string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export uModelVerse inference request logs",
		Long:  "Export uModelVerse inference request logs. Time flags use Unix milliseconds.",
		Run: func(c *cobra.Command, args []string) {
			clearStringIfUnchanged(c.Flags(), "request-id", &req.RequestId)
			req.ModelNames = stringSliceJSONRef(modelNames)
			req.ApiKeyIds = stringSliceJSONRef(apiKeyIds)
			resp, err := invokeUMAction(client, "DownloadUMInferRequestLog", req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintln(ctx.ProgressWriter(), "umodelverse log export task created")
			printResponse(ctx, resp)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	bindProject(cmd, req, ctx.DefaultProjectID())
	req.StartTime = flags.Int64("start-time-ms", 0, "Required. Export start time, Unix timestamp in milliseconds.")
	req.EndTime = flags.Int64("end-time-ms", 0, "Required. Export end time, Unix timestamp in milliseconds.")
	req.Email = flags.String("email", "", "Required. Email address to receive export result.")
	flags.StringSliceVar(&modelNames, "model-name", nil, "Optional. Model name filter. Can be repeated or comma-separated.")
	flags.StringSliceVar(&apiKeyIds, "key-id", nil, "Optional. API key ID filter. Can be repeated or comma-separated.")
	req.RequestId = flags.String("request-id", "", "Optional. Request ID filter.")
	cmd.MarkFlagRequired("region")
	cmd.MarkFlagRequired("zone")
	cmd.MarkFlagRequired("start-time-ms")
	cmd.MarkFlagRequired("end-time-ms")
	cmd.MarkFlagRequired("email")
	return cmd
}
