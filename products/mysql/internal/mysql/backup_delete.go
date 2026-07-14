package mysql

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newUDBBackupDelete ucloud udb backup delete
func newUDBBackupDelete(ctx *cli.Context) *cobra.Command {
	ids := []int{}
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewDeleteUDBBackupRequest()
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete backups of MySQL instance",
		Long:    "Delete backups of MySQL instance",
		Example: "ucloud udb backup delete --backup-id 65534,65535",
		Run: func(c *cobra.Command, args []string) {
			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, id := range ids {
				req.BackupId = sdk.Int(id)
				_, err := client.DeleteUDBBackup(req)
				if err != nil {
					ctx.HandleError(err)
					continue
				}
				fmt.Fprintf(w, "backup[%d] deleted\n", id)
				results = append(results, cli.OpResultRow{ResourceID: strconv.Itoa(id), Action: "delete", Status: "Deleted"})
			}
			ctx.EmitResult(results...)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	flags.IntSliceVar(&ids, "backup-id", nil, "Required. BackupID of backups to delete")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)

	cmd.MarkFlagRequired("backup-id")
	return cmd
}
