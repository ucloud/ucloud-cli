package uhost

import (
	"fmt"

	"github.com/spf13/cobra"

	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// newDelete ucloud uhost delete
func newDelete(ctx *cli.Context) *cobra.Command {
	var uhostIDs *[]string
	var isDestroy = sdk.Bool(false)
	var yes *bool
	var releaseEIP bool
	var releaseUDisk bool
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	req := client.NewTerminateUHostInstanceRequest()
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete Uhost instance",
		Long:  "Delete Uhost instance",
		// SilenceUsage: a delete that fails at runtime must not dump flag usage.
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !ctx.Confirm(*yes, "Are you sure you want to delete the host(s)?") {
				return nil
			}
			if *isDestroy {
				req.Destroy = sdk.Int(1)
			} else {
				req.Destroy = sdk.Int(0)
			}
			req.ReleaseEIP = &releaseEIP
			req.ReleaseUDisk = &releaseUDisk
			reqs := make([]request.Common, len(*uhostIDs))
			for idx, id := range *uhostIDs {
				_req := *req
				id = ctx.PickResourceID(id)
				_req.UHostId = sdk.String(id)
				reqs[idx] = &_req
			}
			prog := ctx.NewProgress()
			// count>5: ctx.ConcurrentAction shows an aggregate counter, so disable
			// per-block animation here (mirrors cmd/util.go concurrentAction.Do
			// calling ux.Doc.Disable()).
			if len(reqs) > 5 {
				prog.Disable()
			}
			fc := &failCounter{}
			rc := &resultCollector{}
			action := deleteUHost(ctx, prog, client, rc)
			ctx.ConcurrentAction(reqs, 50, func(r request.Common) (bool, []string) {
				ok, logs := action(r)
				if !ok {
					fc.inc()
				}
				return ok, logs
			})
			ctx.EmitResult(rc.all()...)
			if n := fc.count(); n > 0 {
				return fmt.Errorf("%d of %d uhost delete operation(s) failed; see the error(s) above or logs in %s", n, len(reqs), ctx.LogFilePath())
			}
			return nil
		},
	}
	flags := cmd.Flags()
	flags.SortFlags = false

	uhostIDs = cmd.Flags().StringSlice("uhost-id", nil, "Requried. ResourceIDs(UhostIds) of the uhost instance")
	// bindRegion/bindProjectID (cmd/uhost.go) → ctx.Bind*: register dynamic
	// region/project completion (golden). --zone stays a raw flag (no completion),
	// matching the original delete.
	ctx.BindRegion(cmd, req)
	ctx.BindProjectID(cmd, req)
	req.Zone = cmd.Flags().String("zone", "", "Optional. availability zone")
	isDestroy = cmd.Flags().Bool("destroy", false, "Optional. false,the uhost instance will be thrown to UHost recycle if you have permission; true,the uhost instance will be deleted directly")
	cmd.Flags().BoolVar(&releaseEIP, "release-eip", true, "Optional. false,Unbind EIP only; true, Unbind EIP and release it")
	cmd.Flags().BoolVar(&releaseUDisk, "delete-cloud-disk", true, "Optional. false, detach cloud disk only; true, detach cloud disk and delete it")
	yes = cmd.Flags().BoolP("yes", "y", false, "Optional. Do not prompt for confirmation.")
	command.SetFlagValues(cmd, "destroy", "true", "false")
	command.SetFlagValues(cmd, "release-eip", "true", "false")
	command.SetFlagValues(cmd, "delete-cloud-disk", "true", "false")
	command.SetCompletion(cmd, "uhost-id", func() []string {
		return getUhostList(ctx, []string{HOST_RUNNING, HOST_STOPPED, HOST_FAIL}, *req.ProjectId, *req.Region, *req.Zone)
	})
	cmd.MarkFlagRequired("uhost-id")

	return cmd
}

// deleteUHost returns the per-uhost delete action for ctx.ConcurrentAction.
// Mirrors cmd/uhost.go deleteUHost (the "====" log-separator + LogInfo are added
// by ctx.ConcurrentAction, not here). The ToQueryMap request-log line is dropped
// (platform handler covers it).
func deleteUHost(ctx *cli.Context, prog *cli.Progress, client *uhostsdk.UHostClient, rc *resultCollector) func(request.Common) (bool, []string) {
	return func(creq request.Common) (bool, []string) {
		req := creq.(*uhostsdk.TerminateUHostInstanceRequest)
		block := prog.NewBlock()
		logs := []string{}
		hostIns, err := sdescribeUHostByID(ctx)(*req.UHostId, nil)
		if err != nil {
			reportFail(ctx, prog, block, fmt.Sprintf("describe uhost[%s] failed: %s", *req.UHostId, cli.ParseError(err)))
			logs = append(logs, fmt.Sprintf("describe uhost[%s] failed: %s", *req.UHostId, cli.ParseError(err)))
			return false, logs
		}

		if hostIns == nil {
			reportFail(ctx, prog, block, fmt.Sprintf("uhost[%s] does not exist", *req.UHostId))
			logs = append(logs, fmt.Sprintf("uhost[%s] does not exist", *req.UHostId))
			return false, logs
		}

		ins := hostIns.(*uhostsdk.UHostInstanceSet)
		if ins.State == "Running" {
			_req := client.NewStopUHostInstanceRequest()
			_req.ProjectId = req.ProjectId
			_req.Region = req.Region
			_req.Zone = req.Zone
			_req.UHostId = req.UHostId
			stopUhostInsV2(ctx, prog, client, _req, false, block)
		}

		resp, err := client.TerminateUHostInstance(req)
		if err != nil {
			reportFail(ctx, prog, block, cli.ParseError(err))
			logs = append(logs, fmt.Sprintf("delete uhost[%s] failed: %s", *req.UHostId, cli.ParseError(err)))
			return false, logs
		}
		text := fmt.Sprintf("uhost[%s] deleted", resp.UHostId)
		logs = append(logs, text)
		block.Append(text)
		rc.add(cli.OpResultRow{ResourceID: resp.UHostId, Action: "delete", Status: "Deleted"})
		return true, logs
	}
}

// stopUhostInsV2 is the concurrent (block-based) stop used by delete. Mirrors
// cmd/uhost.go stopUhostInsV2.
func stopUhostInsV2(ctx *cli.Context, prog *cli.Progress, client *uhostsdk.UHostClient, req *uhostsdk.StopUHostInstanceRequest, async bool, block *cli.Block) {
	resp, err := client.StopUHostInstance(req)
	if err != nil {
		block.Append(cli.ParseError(err))
		return
	}

	text := fmt.Sprintf("uhost[%v] is shutting down", resp.UHostId)
	if async {
		block.Append(text)
	} else {
		prog.Sspoll(sdescribeUHostByID(ctx), resp.UHostId, text, []string{HOST_STOPPED, HOST_FAIL}, block, nil)
	}
}
