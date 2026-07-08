package umodelverse

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func newLogList(ctx *cli.Context) *cobra.Command {
	client := newClient(ctx)
	req := &requestLogRequest{}
	newRequest(client, req, true)
	var modelNames []string
	var apiKeyIds []string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List uModelVerse inference request logs",
		Long:  "List uModelVerse inference request logs.",
		PreRunE: func(c *cobra.Command, args []string) error {
			if req.Limit != nil && !isAllowedLogLimit(*req.Limit) {
				return fmt.Errorf("limit must be one of [10, 20, 50, 100]")
			}
			return nil
		},
		Run: func(c *cobra.Command, args []string) {
			clearStringIfUnchanged(c.Flags(), "request-id", &req.RequestId)
			req.ModelNames = stringSliceJSONRef(modelNames)
			req.ApiKeyIds = stringSliceJSONRef(apiKeyIds)
			resp, err := invokeUMAction(client, "ListUMInferRequestLogs", req)
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
	req.StartTime = flags.Int64("start-time-ms", 0, "Required. Query start time, Unix timestamp in milliseconds.")
	req.EndTime = flags.Int64("end-time-ms", 0, "Required. Query end time, Unix timestamp in milliseconds.")
	flags.StringSliceVar(&modelNames, "model-name", nil, "Optional. Model name filter. Can be repeated or comma-separated.")
	flags.StringSliceVar(&apiKeyIds, "key-id", nil, "Optional. API key ID filter. Can be repeated or comma-separated.")
	req.RequestId = flags.String("request-id", "", "Optional. Request ID filter.")
	req.Offset = flags.Int("offset", 0, "Optional. The index of log which start to list.")
	req.Limit = flags.Int("limit", 20, "Optional. The maximum number of logs per page. Allowed values: 10, 20, 50, 100.")
	cmd.MarkFlagRequired("region")
	cmd.MarkFlagRequired("zone")
	cmd.MarkFlagRequired("start-time-ms")
	cmd.MarkFlagRequired("end-time-ms")
	return cmd
}

func isAllowedLogLimit(limit int) bool {
	switch limit {
	case 10, 20, 50, 100:
		return true
	default:
		return false
	}
}
