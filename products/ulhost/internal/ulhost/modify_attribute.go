package ulhost

import (
	"fmt"

	"github.com/spf13/cobra"

	ucompsharesdk "github.com/ucloud/ucloud-sdk-go/services/ucompshare"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newModifyAttribute ucloud ulhost modify-attribute
func newModifyAttribute(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ucompsharesdk.NewClient)
	req := client.NewModifyULHostAttributeRequest()
	cmd := &cobra.Command{
		Use:   "modify-attribute",
		Short: "Modify the attribute of ULHost instance",
		Long:  "Modify the attribute (name or remark) of ULHost instance. At least one of Name or Remark must be specified.",
		Run: func(cmd *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			req.ULHostId = sdk.String(ctx.PickResourceID(*req.ULHostId))
			resp, err := client.ModifyULHostAttribute(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintf(w, "ulhost[%s] attribute modified\n", resp.ULHostId)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false
	req.ULHostId = flags.String("ulhost-id", "", "Required. Resource ID of the ulhost instance")
	req.Name = flags.String("name", "", "Optional. New name of the ulhost instance")
	req.Remark = flags.String("remark", "", "Optional. New remark of the ulhost instance")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	command.SetCompletion(cmd, "ulhost-id", func() []string {
		return getULHostList(ctx, nil, *req.ProjectId, *req.Region)
	})
	cmd.MarkFlagRequired("ulhost-id")
	return cmd
}
