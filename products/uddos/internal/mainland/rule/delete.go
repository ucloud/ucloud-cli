// Package rule ...
//
// @Brief  删除国内高防转发规则命令
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
package rule

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ucloud/ucloud-sdk-go/services/uaccount"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newDelete 构建 uddos mainland rule delete 命令
//
// @Brief  构建国内高防 rule delete 子命令，调用 DeleteBGPServiceFwdRule
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
	var resourceID string
	var ruleIndex int
	var yes bool

	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete a BGP forwarding rule",
		Long:    "Delete a forwarding rule from a mainland BGP DDoS protection service",
		Example: "  ucloud uddos mainland rule delete --resource-id ghp-xxxxx --rule-index 0",
		Run: func(cmd *cobra.Command, args []string) {
			confirmed, err := ctx.Confirm(yes, fmt.Sprintf("Are you sure to delete rule index %d from service %s?", ruleIndex, resourceID))
			if err != nil {
				ctx.HandleError(fmt.Errorf("confirm: %w", err))
				return
			}
			if !confirmed {
				return
			}
			client := cli.NewServiceClient(ctx, uaccount.NewClient)
			params := map[string]interface{}{
				"Action":     "DeleteBGPServiceFwdRule",
				"ResourceId": resourceID,
				"RuleIndex":  ruleIndex,
			}
			req := client.NewGenericRequest()
			if err := req.SetPayload(params); err != nil {
				ctx.HandleError(fmt.Errorf("set payload: %w", err))
				return
			}
			_, invokeErr := client.GenericInvoke(req)
			if invokeErr != nil {
				ctx.HandleError(fmt.Errorf("DeleteBGPServiceFwdRule: %w", invokeErr))
				return
			}
			fmt.Fprintf(ctx.ProgressWriter(), "rule[%d] deleted from service[%s]\n", ruleIndex, resourceID)
			ctx.EmitResult(cli.OpResultRow{
				ResourceID: fmt.Sprintf("%d", ruleIndex),
				Action:     "delete",
				Status:     "Deleted",
			})
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&resourceID, "resource-id", "", "Required. Service resource ID")
	flags.IntVar(&ruleIndex, "rule-index", 0, "Required. Rule index to delete")
	flags.BoolVarP(&yes, "yes", "y", false, "Optional. Skip confirmation prompt")
	cmd.MarkFlagRequired("resource-id")
	cmd.MarkFlagRequired("rule-index")
	return cmd
}
