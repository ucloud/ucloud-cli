package ulb

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	ulbsdk "github.com/ucloud/ucloud-sdk-go/services/ulb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newSSLList returns ucloud ulb ssl list.
func newSSLList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, ulbsdk.NewClient)
	req := client.NewDescribeSSLRequest()
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List SSL Certificates",
		Long:  "List SSL Certificates",
		Run: func(c *cobra.Command, args []string) {
			req.ProjectId = sdk.String(ctx.PickResourceID(*req.ProjectId))
			resp, err := client.DescribeSSL(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			rows := []SSLCertificate{}
			for _, ssl := range resp.DataSet {
				row := SSLCertificate{}
				row.Name = ssl.SSLName
				row.ResourceID = ssl.SSLId
				row.MD5 = ssl.HashValue
				row.UploadTime = common.FormatDateTime(ssl.CreateTime)
				targets := []string{}
				for _, t := range ssl.BindedTargetSet {
					item := fmt.Sprintf("%s/%s(%s/%s)", t.VServerId, t.VServerName, t.ULBId, t.ULBName)
					targets = append(targets, item)
				}
				row.BindResource = strings.Join(targets, ",")
				rows = append(rows, row)
			}
			ctx.PrintList(rows)
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	req.SSLId = flags.String("ssl-id", "", "Optional. ResouceID of ssl certificate to list")
	ctx.BindLimit(cmd, req)
	ctx.BindOffset(cmd, req)

	return cmd
}
