package ulb

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	ulbsdk "github.com/ucloud/ucloud-sdk-go/services/ulb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newSSLDescribe returns ucloud ulb ssl describe.
func newSSLDescribe(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ulbsdk.NewClient)
	req := client.NewDescribeSSLRequest()
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Display all data associated with SSL Certificate",
		Long:  "Display all data associated with SSL Certificate",
		Run: func(c *cobra.Command, args []string) {
			req.SSLId = sdk.String(ctx.PickResourceID(*req.SSLId))
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			resp, err := client.DescribeSSL(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if len(resp.DataSet) <= 0 {
				fmt.Fprintf(ctx.ProgressWriter(), "ssl certificate[%s] is not exists\n", *req.SSLId)
				return
			}

			sslcf := resp.DataSet[0]
			targets := []string{}
			for _, t := range sslcf.BindedTargetSet {
				item := fmt.Sprintf("%s/%s-%s/%s", t.ULBId, t.ULBName, t.VServerId, t.VServerName)
				targets = append(targets, item)
			}
			rows := []cli.DescribeRow{
				{Attribute: "ResourceID", Content: sslcf.SSLId},
				{Attribute: "Name", Content: sslcf.SSLName},
				{Attribute: "Type", Content: sslcf.SSLType},
				{Attribute: "UploadTime", Content: common.FormatDateTime(sslcf.CreateTime)},
				{Attribute: "BindResource", Content: strings.Join(targets, ",")},
				{Attribute: "MD5", Content: sslcf.HashValue},
				{Attribute: "Content", Content: sslcf.SSLContent},
			}
			printDescribe(ctx, rows)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	req.SSLId = flags.String("ssl-id", "", "Required. ResouceID of ssl certificate to describe")
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	command.SetCompletion(cmd, "ssl-id", func() []string {
		return getAllSSLCertIDNames(ctx, *req.ProjectId, *req.Region)
	})
	cmd.MarkFlagRequired("ssl-id")
	return cmd
}

func printDescribe(ctx *cli.Context, rows []cli.DescribeRow) {
	if ctx.Format() != cli.OutputTable {
		ctx.PrintList(rows)
		return
	}
	for _, row := range rows {
		fmt.Fprintln(ctx.Out(), row.Attribute)
		fmt.Fprintln(ctx.Out(), row.Content)
		fmt.Fprintln(ctx.Out())
	}
}
