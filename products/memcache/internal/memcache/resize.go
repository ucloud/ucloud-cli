package memcache

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/umem"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newResize returns ucloud memcache resize.
func newResize(ctx *cli.Context) *cobra.Command {
	idNames := make([]string, 0)
	client := cli.NewServiceClient(ctx, umem.NewClient)
	req := client.NewResizeUMemcacheGroupRequest()
	cmd := &cobra.Command{
		Use:   "resize",
		Short: "Resize memcache instances",
		Long:  "Resize memcache instances",
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
			ctx.ConcurrentAction(reqs, 10, resize(ctx, client, prog))
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&idNames, "umem-id", nil, "Required. Resource ID of memcache to resize")
	req.Size = flags.Int("size-gb", 0, "Required. Target memory size in GB. Accept values:1,2,4,8,16,32")
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	command.SetCompletion(cmd, "umem-id", func() []string {
		return getIDList(ctx, *req.ProjectId, *req.Region)
	})
	command.SetFlagValues(cmd, "size-gb", "1", "2", "4", "8", "16", "32")

	cmd.MarkFlagRequired("umem-id")
	cmd.MarkFlagRequired("size-gb")
	return cmd
}

func resize(ctx *cli.Context, client *umem.UMemClient, prog *cli.Progress) func(request.Common) (bool, []string) {
	return func(creq request.Common) (bool, []string) {
		req := creq.(*umem.ResizeUMemcacheGroupRequest)
		block := prog.NewBlock()
		logs := []string{}
		_, err := client.ResizeUMemcacheGroup(req)
		if err != nil {
			msg := fmt.Sprintf("resize memcache[%s] failed: %s", *req.GroupId, cli.ParseError(err))
			block.Append(cli.ParseError(err))
			logs = append(logs, msg)
			return false, logs
		}
		text := fmt.Sprintf("memcache[%s] is resizing", *req.GroupId)
		ret := ctx.PollerTo(ctx.ProgressWriter(), describeByID(ctx)).Sspoll(*req.GroupId, text, []string{UMEM_RUNNING, UMEM_FAIL}, block, nil)
		if ret.Err != nil {
			block.Append(cli.ParseError(ret.Err))
			logs = append(logs, ret.Err.Error())
		}
		if ret.Timeout {
			logs = append(logs, fmt.Sprintf("poll memcache[%s] timeout", *req.GroupId))
		}
		return ret.Done, logs
	}
}
