package umodelverse

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func newModelList(ctx *cli.Context) *cobra.Command {
	client := newClient(ctx)
	req := &squareModelRequest{}
	newRequest(client, req, true)
	var maxModelLen []int
	var language []string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List uModelVerse square models",
		Long:  "List uModelVerse square models.",
		Run: func(c *cobra.Command, args []string) {
			flags := c.Flags()
			clearStringIfUnchanged(flags, "model-type", &req.ModelType)
			clearStringIfUnchanged(flags, "keyword", &req.KeyWord)
			clearStringIfUnchanged(flags, "order-by", &req.OrderBy)
			clearStringIfUnchanged(flags, "order", &req.Order)
			req.MaxModelLen = intSliceJSONRef(maxModelLen)
			req.Language = stringSliceJSONRef(language)
			resp, err := invokeUMAction(client, "ListUFSquareModel", req)
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
	ctx.BindProjectID(cmd, req)
	req.ModelType = flags.String("model-type", "", "Optional. Model type.")
	req.KeyWord = flags.String("keyword", "", "Optional. Keyword filter.")
	req.Offset = flags.Int("offset", 0, "Optional. The index of model which start to list.")
	req.Limit = flags.Int("limit", 20, "Optional. The maximum number of models per page.")
	req.OrderBy = flags.String("order-by", "", "Optional. Sort field.")
	req.Order = flags.String("order", "", "Optional. Sort order.")
	flags.IntSliceVar(&maxModelLen, "max-model-len", nil, "Optional. Context length filter. Can be repeated or comma-separated.")
	flags.StringSliceVar(&language, "language", nil, "Optional. Language filter, e.g. chinese or english. Can be repeated or comma-separated.")
	return cmd
}
