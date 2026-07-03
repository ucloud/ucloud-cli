package uhost

import (
	"errors"
	"fmt"

	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

var errStopDeclined = errors.New("skip, you do not agree to stop uhost")

// stopUhostIns stops a uhost and (unless async) polls it to Stopped. Mirrors
// cmd/uhost.go stopUhostIns (sequential base.NewPoller → ctx.PollerTo.Spoll).
func stopUhostIns(ctx *cli.Context, client *uhostsdk.UHostClient, req *uhostsdk.StopUHostInstanceRequest, async bool) bool {
	w := ctx.ProgressWriter()
	resp, err := client.StopUHostInstance(req)
	if err != nil {
		ctx.HandleError(err)
		return false
	}

	text := fmt.Sprintf("uhost[%v] is shutting down", resp.UHostId)
	if async {
		fmt.Fprintln(w, text)
		return false
	}
	// base.Poller.Poll returned a bool (reached target state) that cmd/uhost.go
	// fed back into resize (inst.State = Stopped). The platform Spoll narrates to
	// the writer but returns nothing, so we return true here: a successful
	// (non-async) stop request that we then polled is treated as "stopped" for
	// the resize state-transition, which matches the original intent.
	ctx.PollerTo(w, describeUHostByID(ctx, *req.ProjectId, *req.Region, *req.Zone)).Spoll(resp.UHostId, text, []string{HOST_STOPPED, HOST_FAIL})
	return true
}

// checkAndCloseUhost stops the uhost (with optional prompt) if it is running.
// Mirrors cmd/uhost.go checkAndCloseUhost.
func checkAndCloseUhost(ctx *cli.Context, client *uhostsdk.UHostClient, yes, async bool, uhostID, project, region, zone string) error {
	host, err := describeUHostByID(ctx, project, region, zone)(uhostID, nil)
	if err != nil {
		return err
	}
	inst, ok := host.(*uhostsdk.UHostInstanceSet)
	if ok {
		if inst.State == "Running" {
			ok, err := ctx.Confirm(yes, fmt.Sprintf("uhost[%s] will be stopped, can we do this?", uhostID))
			if err != nil {
				return err
			}
			if !ok {
				return errStopDeclined
			}
			_req := client.NewStopUHostInstanceRequest()
			_req.ProjectId = &project
			_req.Region = &region
			_req.Zone = &zone
			_req.UHostId = &uhostID
			stopUhostIns(ctx, client, _req, async)
		}
	} else {
		return fmt.Errorf("Something wrong, uhost[%s] may not exist", uhostID)
	}
	return nil
}
