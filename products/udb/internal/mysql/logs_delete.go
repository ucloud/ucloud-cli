package mysql

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newUDBLogArchiveDelete ucloud udb log archive delete
func newUDBLogArchiveDelete(ctx *cli.Context) *cobra.Command {
	var ids []int
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewDeleteUDBLogPackageRequest()
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete log archives(log files)",
		Long:    "Delete log archives(log files)",
		Example: "ucloud mysql logs delete --archive-id 35025",
		Run: func(c *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, id := range ids {
				req.BackupId = sdk.Int(id)
				_, err := client.DeleteUDBLogPackage(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(w, "archive[%d] deleted\n", id)
				results = append(results, cli.OpResultRow{ResourceID: strconv.Itoa(id), Action: "delete", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.IntSliceVar(&ids, "archive-id", nil, "Optional. ArchiveID of log archives to delete")
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("archive-id")

	return cmd
}
