package umongodb

import (
	"fmt"

	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-sdk-go/services/umongodb"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// deleteOpts captures the configuration that differs between delete-replset
// and delete-sharded so the shared newDeleteCmd can handle both.
type deleteOpts struct {
	use        string
	short      string
	long       string
	action     string // GenericInvoke action name
	idParam    string // GenericInvoke cluster ID param name ("ClusterId" or "ShardedClusterId")
	idFlagDesc string // help text for the --umongodb-id flag
}

// newDeleteCmd returns a cobra.Command that stops (optionally) and then
// deletes MongoDB clusters via GenericInvoke. The stop step uses the typed
// SDK; the delete step uses GenericInvoke because the SDK has no typed
// delete methods.
func newDeleteCmd(ctx *cli.Context, opts deleteOpts) *cobra.Command {
	var async bool
	var skipStop bool
	var yes bool
	var ids []string

	stopClient := cli.NewServiceClient(ctx, umongodb.NewClient)
	stopReq := stopClient.NewStopUMongoDBClusterRequest()

	var common request.CommonBase

	cmd := &cobra.Command{
		Use:   opts.use,
		Short: opts.short,
		Long:  opts.long,
		Run: func(c *cobra.Command, args []string) {
			// Confirm before destructive operation
			ok, err := ctx.Confirm(yes, "Are you sure you want to delete the umongodb cluster(s)?")
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if !ok {
				return
			}

			region := common.GetRegion()
			zone := common.GetZone()
			projectID := common.GetProjectId()

			// Set loop-invariant stop fields once
			if !skipStop {
				stopReq.Region = &region
				if zone != "" {
					stopReq.Zone = &zone
				}
				if projectID != "" {
					stopReq.ProjectId = &projectID
				}
			}

			w := ctx.ProgressWriter()
			results := []cli.OpResultRow{}
			for _, idName := range ids {
				id := ctx.PickResourceID(idName)

				// Step 1: Stop the cluster (skip when --skip-stop is set)
				if !skipStop {
					stopReq.ClusterId = sdk.String(id)
					_, err := stopClient.StopUMongoDBCluster(stopReq)
					if err != nil {
						ctx.HandleError(fmt.Errorf("stop %s before delete: %w", id, err))
						continue
					}

					// Always poll for Stopped before deleting, even in async mode.
					// Only the final delete-step polling is skipped when --async is set.
					text := fmt.Sprintf("%s[%s] is stopping before delete", productName, id)
					ctx.PollerTo(w, describeByID(ctx, region, zone)).Spoll(id, text, []string{stateStopped, stateFail})
				}

				// Step 2: Delete the cluster via GenericInvoke
				params := map[string]interface{}{
					"Action": opts.action,
					"Region": region,
					opts.idParam: id,
				}
				if zone != "" {
					params["Zone"] = zone
				}
				if projectID != "" {
					params["ProjectId"] = projectID
				}

				if _, err := genericCall(ctx, opts.action, params); err != nil {
					ctx.HandleError(fmt.Errorf("delete %s: %w", id, err))
					continue
				}

				text := fmt.Sprintf("%s[%s] is deleting", productName, id)
				if async {
					fmt.Fprintln(w, text)
				} else {
					ctx.PollerTo(w, describeByID(ctx, region, zone)).Spoll(id, text, []string{stateFail})
				}
				results = append(results, cli.OpResultRow{ResourceID: id, Action: "delete", Status: "Deleting"})
			}
			ctx.EmitResult(results...)
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringSliceVar(&ids, resourceIDFlag, nil, "Required. "+opts.idFlagDesc)
	flags.BoolVarP(&async, "async", "a", false, "Optional. Do not wait for the operation to finish.")
	flags.BoolVar(&skipStop, "skip-stop", false, "Optional. Skip the stop-before-delete step.")
	flags.BoolVarP(&yes, "yes", "y", false, "Optional. Do not prompt for confirmation.")

	ctx.BindRegion(cmd, &common)
	ctx.BindZone(cmd, &common)
	ctx.BindProjectID(cmd, &common)

	cmd.MarkFlagRequired(resourceIDFlag)
	command.SetCompletion(cmd, resourceIDFlag, func() []string {
		return getMongoDBIDList(ctx, nil, common.GetRegion(), common.GetZone(), common.GetProjectId())
	})

	return cmd
}
