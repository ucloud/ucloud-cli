// Package ip ...
//
// @Brief  删除国内高防IP命令
//
// @File   delete.go
//
// @Author leas.li(cc)
//
// @Email  leas.li@ucloud.cn
//
// @Date   2026/07/11
//
// @CopyRights(C) UCloud All rights reserved.
package ip

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ucloud/ucloud-sdk-go/services/uaccount"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newDelete 构建 uddos mainland ip delete 命令
//
// @Brief  构建国内高防 ip delete 子命令，调用 DeleteBGPServiceIP
//
// @Param  ctx *cli.Context
//
// @Return *cobra.Command
//
// @Author leas.li(cc)
//
// @Email  leas.li@ucloud.cn
//
// @Date   2026/07/11
func newDelete(ctx *cli.Context) *cobra.Command {
	var resourceID, defenceIP string
	var yes bool

	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete a mainland BGP high-protection IP",
		Long:    "Delete a BGP DDoS protection IP from the specified mainland service",
		Example: "  ucloud uddos mainland ip delete --resource-id ghp-xxxxx --defence-ip 1.2.3.4",
		Run: func(cmd *cobra.Command, args []string) {
			confirmed, err := ctx.Confirm(yes, fmt.Sprintf("Are you sure to delete defence IP %s from service %s?", defenceIP, resourceID))
			if err != nil {
				ctx.HandleError(fmt.Errorf("confirm: %w", err))
				return
			}
			if !confirmed {
				return
			}
			client := cli.NewServiceClient(ctx, uaccount.NewClient)
			params := map[string]interface{}{
				"Action":     "DeleteBGPServiceIP",
				"ResourceId": resourceID,
				"DefenceIp":  defenceIP,
			}
			req := client.NewGenericRequest()
			if err := req.SetPayload(params); err != nil {
				ctx.HandleError(fmt.Errorf("set payload: %w", err))
				return
			}
			_, invokeErr := client.GenericInvoke(req)
			if invokeErr != nil {
				ctx.HandleError(fmt.Errorf("DeleteBGPServiceIP: %w", invokeErr))
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "BGP IP deleted: %s\n", defenceIP)
			ctx.EmitResult(cli.OpResultRow{ResourceID: defenceIP, Action: "delete", Status: "Deleted"})
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&resourceID, "resource-id", "", "Required. Service resource ID")
	flags.StringVar(&defenceIP, "defence-ip", "", "Required. BGP defence IP to delete")
	flags.BoolVarP(&yes, "yes", "y", false, "Optional. Skip confirmation prompt")
	cmd.MarkFlagRequired("resource-id")
	cmd.MarkFlagRequired("defence-ip")
	return cmd
}
