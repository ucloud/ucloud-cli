package redis

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/umem"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newRestart returns ucloud redis restart.
func newRestart(ctx *cli.Context) *cobra.Command {
	idNames := make([]string, 0)
	client := cli.NewServiceClient(ctx, umem.NewClient)
	req := client.NewRestartURedisGroupRequest()
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart redis instances of master-replica type",
		Long:  "Restart redis instances of master-replica type",
		Run: func(c *cobra.Command, args []string) {
			reqs := make([]request.Common, len(idNames))
			for idx, idname := range idNames {
				id := ctx.PickResourceID(idname)
				next := *req
				next.GroupId = &id
				reqs[idx] = &next
			}
			prog := ctx.NewProgress()
			if len(reqs) > 5 {
				prog.Disable()
			}
			ctx.ConcurrentAction(reqs, 10, restart(ctx, client, prog))
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "umem-id", nil, "Required. Resource ID of redis instances to restart")
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)

	cmd.MarkFlagRequired("umem-id")
	command.SetCompletion(cmd, "umem-id", func() []string {
		return getIDList(ctx, *req.ProjectId, *req.Region)
	})

	return cmd
}

func restart(ctx *cli.Context, client *umem.UMemClient, prog *cli.Progress) func(request.Common) (bool, []string) {
	return func(creq request.Common) (bool, []string) {
		req := creq.(*umem.RestartURedisGroupRequest)
		block := prog.NewBlock()
		logs := []string{}
		_, err := client.RestartURedisGroup(req)
		if err != nil {
			msg := fmt.Sprintf("restart redis[%s] failed: %s", *req.GroupId, cli.ParseError(err))
			block.Append(cli.ParseError(err))
			logs = append(logs, msg)
			return false, logs
		}
		text := fmt.Sprintf("redis[%s] is restarting", *req.GroupId)
		ret := ctx.PollerTo(ctx.ProgressWriter(), describeByID(ctx)).Sspoll(*req.GroupId, text, []string{UMEM_RUNNING, UMEM_FAIL}, block, nil)
		if ret.Err != nil {
			block.Append(cli.ParseError(ret.Err))
			logs = append(logs, ret.Err.Error())
		}
		if ret.Timeout {
			logs = append(logs, fmt.Sprintf("poll redis[%s] timeout", *req.GroupId))
		}
		return ret.Done, logs
	}
}
