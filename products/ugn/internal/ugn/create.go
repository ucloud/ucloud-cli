package ugn

import (
	"fmt"

	"github.com/spf13/cobra"

	ugnsdk "github.com/ucloud/ucloud-sdk-go/services/ugn"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newCreate ucloud ugn create
func newCreate(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ugnsdk.NewClient)
	req := client.NewCreateUGNRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a ugn instance",
		Long:  "Create a ugn instance",
		Run: func(c *cobra.Command, args []string) {
			resp, err := client.CreateUGN(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "ugn[%s] created\n", resp.UGNID)
			ctx.EmitResult(cli.OpResultRow{ResourceID: resp.UGNID, Action: "create", Status: "Created"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	req.Name = flags.String("name", "", "Required. Name of the ugn instance to create")
	req.Remark = flags.String("remark", "", "Optional. Remark")

	ctx.BindProjectID(cmd, req)
	ctx.SetCompletion(cmd, "project-id", ctx.ProjectList)

	cmd.MarkFlagRequired("name")

	return cmd
}
