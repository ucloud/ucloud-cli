package gssh

import (
	"github.com/spf13/cobra"

	pathxsdk "github.com/ucloud/ucloud-sdk-go/services/pathx"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newList ucloud gssh list
func newList(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, pathxsdk.NewClient)
	req := client.NewDescribeGlobalSSHInstanceRequest()
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all GlobalSSH instances",
		Long:    "List all GlobalSSH instances",
		Example: "ucloud gssh list",
		Run: func(cmd *cobra.Command, args []string) {
			areaMap := map[string]string{
				"洛杉矶":  "LosAngeles",
				"新加坡":  "Singapore",
				"香港":   "HongKong",
				"东京":   "Tokyo",
				"华盛顿":  "Washington",
				"法兰克福": "Frankfurt",
				"拉各斯":  "Lagos",
			}
			resp, err := client.DescribeGlobalSSHInstance(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			list := make([]GSSHRow, 0)
			for _, gssh := range resp.InstanceSet {
				row := GSSHRow{
					ResourceID:         gssh.InstanceId,
					SSHServerIP:        gssh.TargetIP,
					AcceleratingDomain: gssh.AcceleratingDomain,
					SSHPort:            gssh.Port,
					GlobalSSHPort:      gssh.GlobalSSHPort,
					Remark:             gssh.Remark,
					InstanceType:       gssh.InstanceType,
				}
				if val, ok := areaMap[gssh.Area]; ok {
					row.SSHServerLocation = val
				} else {
					row.SSHServerLocation = gssh.Area
				}
				list = append(list, row)
			}
			ctx.PrintList(list)
		},
	}
	cmd.Flags().SortFlags = false
	req.Region = cmd.Flags().String("region", ctx.DefaultRegion(), "Optional. Assign region")
	req.ProjectId = cmd.Flags().String("project-id", ctx.DefaultProjectID(), "Optional. Assign project-id")
	return cmd
}
