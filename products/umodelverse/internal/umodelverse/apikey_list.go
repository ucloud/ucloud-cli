package umodelverse

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func newAPIKeyList(ctx *cli.Context) *cobra.Command {
	client := newClient(ctx)
	req := &apiKeyRequest{}
	newRequest(client, req, true)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List uModelVerse API keys",
		Long:  "List uModelVerse API keys.",
		Run: func(c *cobra.Command, args []string) {
			flags := c.Flags()
			clearStringIfUnchanged(flags, "key-id", &req.KeyId)
			clearIntIfUnchanged(flags, "modelverse-disabled", &req.ModelverseDisabled)
			clearIntIfUnchanged(flags, "sandbox-disabled", &req.SandBoxDisabled)
			resp, err := invokeUMAction(client, "ListUMInferAPIKey", req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			printResponse(ctx, resp)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false
	req.KeyId = flags.String("key-id", "", "Optional. API key ID filter.")
	req.Offset = flags.Int("offset", 0, "Optional. The index of API key which start to list.")
	req.Limit = flags.Int("limit", 20, "Optional. The maximum number of API keys per page.")
	req.ModelverseDisabled = flags.Int("modelverse-disabled", 0, "Optional. Whether ModelVerse is disabled: 0 enabled, 1 disabled.")
	req.SandBoxDisabled = flags.Int("sandbox-disabled", 0, "Optional. Whether sandbox is disabled: 0 enabled, 1 disabled.")
	bindProject(cmd, req, ctx.DefaultProjectID())
	return cmd
}
