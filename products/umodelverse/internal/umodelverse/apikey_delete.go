package umodelverse

import (
	"fmt"

	"github.com/spf13/cobra"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func newAPIKeyDelete(ctx *cli.Context) *cobra.Command {
	client := newClient(ctx)
	req := &apiKeyRequest{}
	newRequest(client, req, true)

	var yes bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a uModelVerse API key",
		Long:  "Delete a uModelVerse API key by key ID.",
		Run: func(c *cobra.Command, args []string) {
			id := ctx.PickResourceID(*req.KeyId)
			ok, err := ctx.Confirm(yes, fmt.Sprintf("Are you sure you want to delete API key %s?", id))
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}
			req.KeyId = sdk.String(id)
			if _, err := invokeUMAction(client, "DeleteUMInferAPIKey", req); err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "umodelverse apikey[%s] deleted\n", id)
			ctx.EmitResult(cli.OpResultRow{ResourceID: id, Action: "delete", Status: "Deleted"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	req.KeyId = flags.String("key-id", "", "Required. API key ID to delete.")
	flags.BoolVarP(&yes, "yes", "y", false, "Optional. Skip the confirmation prompt.")
	bindProject(cmd, req, ctx.DefaultProjectID())

	cmd.MarkFlagRequired("key-id")
	return cmd
}
