package mysql

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/udb"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newUDBLogArchiveGetDownloadURL ucloud udb log archive get-download-url
func newUDBLogArchiveGetDownloadURL(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewDescribeUDBBinlogBackupURLRequest()
	cmd := &cobra.Command{
		Use:     "download",
		Short:   "Display url of an archive(log file)",
		Long:    "Display url of an archive(log file)",
		Example: "ucloud mysql logs download --udb-id udb-urixxx/test.cli1 --archive-id 35044",
		Run: func(c *cobra.Command, args []string) {
			*req.DBId = ctx.PickResourceID(*req.DBId)
			resp, err := client.DescribeUDBBinlogBackupURL(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			fmt.Fprintln(ctx.Out(), resp.BackupPath)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.BackupId = flags.Int("archive-id", 0, "Required. ArchiveID of archive to download")
	req.DBId = flags.String("udb-id", "", "Required. Resource ID of UDB which the archive belongs to")
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("archive-id")
	cmd.MarkFlagRequired("udb-id")

	command.SetCompletion(cmd, "udb-id", func() []string {
		return getUDBIDList(ctx, nil, "sql", *req.ProjectId, *req.Region, *req.Zone)
	})

	return cmd
}
