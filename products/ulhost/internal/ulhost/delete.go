package ulhost

import (
	"fmt"

	"github.com/spf13/cobra"

	ucompsharesdk "github.com/ucloud/ucloud-sdk-go/services/ucompshare"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newDelete ucloud ulhost delete
func newDelete(ctx *cli.Context) *cobra.Command {
	var ulhostIDs *[]string
	var yes *bool
	var releaseUDisk bool
	client := cli.NewServiceClient(ctx, ucompsharesdk.NewClient)
	req := client.NewTerminateULHostInstanceRequest()
	cmd := &cobra.Command{
		Use:          "delete",
		Short:        "Delete ULHost instance",
		Long:         "Delete ULHost instance",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ok, err := ctx.Confirm(*yes, "Are you sure you want to delete the ulhost instance(s)?")
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
			req.ReleaseUDisk = sdk.Bool(releaseUDisk)
			w := ctx.ProgressWriter()
			for _, id := range *ulhostIDs {
				id = ctx.PickResourceID(id)
				req.ULHostId = sdk.String(id)
				resp, err := client.TerminateULHostInstance(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				if resp.InRecycle == "Yes" {
					fmt.Fprintf(w, "ulhost[%s] has been moved to recycle bin\n", resp.ULHostId)
				} else {
					fmt.Fprintf(w, "ulhost[%s] deleted\n", resp.ULHostId)
				}
			}
			return nil
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	ulhostIDs = flags.StringSlice("ulhost-id", nil, "Required. ResourceIDs(ULHostIds) of the ulhost instance")
	req.ProjectId = flags.String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	req.Region = flags.String("region", ctx.DefaultRegion(), "Optional. Assign region")
	flags.BoolVar(&releaseUDisk, "delete-cloud-disk", true, "Optional. false, detach cloud disk only; true, detach cloud disk and delete it")
	yes = flags.BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")
	command.SetFlagValues(cmd, "delete-cloud-disk", "true", "false")
	command.SetCompletion(cmd, "ulhost-id", func() []string {
		return getULHostList(ctx, []string{HOST_RUNNING, HOST_STOPPED, HOST_FAIL}, *req.ProjectId, *req.Region)
	})
	cmd.MarkFlagRequired("ulhost-id")

	return cmd
}
