package tidb

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/tidb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// parseResizeDiskNodeConfig parses the CLI node-config string for resize-disk.
// Format: DiskSize=N,ServerType=tidb
func parseResizeDiskNodeConfig(s string) (tidb.ModifyTiDBClusterUhostDiskParamNodeConfig, error) {
	var cfg tidb.ModifyTiDBClusterUhostDiskParamNodeConfig
	parts := strings.Split(s, ",")
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			return cfg, fmt.Errorf("invalid node-config segment %q, expected key=value", part)
		}
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		switch key {
		case "DiskSize":
			n, err := strconv.Atoi(val)
			if err != nil {
				return cfg, fmt.Errorf("invalid DiskSize %q: %w", val, err)
			}
			cfg.DiskSize = sdk.Int(n)
		case "ServerType":
			cfg.ServerType = sdk.String(val)
		default:
			return cfg, fmt.Errorf("unknown node-config key %q", key)
		}
	}
	if cfg.DiskSize == nil || cfg.ServerType == nil {
		return cfg, fmt.Errorf("node-config must include DiskSize and ServerType")
	}
	return cfg, nil
}

// newResizeDisk ucloud utidb resize-disk
func newResizeDisk(ctx *cli.Context) *cobra.Command {
	var id, scaleType, nodeConfig string
	var startTime int

	client := cli.NewServiceClient(ctx, tidb.NewClient)
	req := client.NewModifyTiDBClusterUhostDiskRequest()

	cmd := &cobra.Command{
		Use:   "resize-disk",
		Short: "Resize disk of a UTiDB instance",
		Long:  "Resize disk of a UTiDB instance",
		Run: func(c *cobra.Command, args []string) {
			cfg, err := parseResizeDiskNodeConfig(nodeConfig)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			pickedID := ctx.PickResourceID(id)
			req.Id = sdk.String(pickedID)
			req.ScaleType = sdk.String(scaleType)
			req.NodeConfig = &cfg
			if startTime != 0 {
				req.StartTime = sdk.Int(startTime)
			}

			_, err = client.ModifyTiDBClusterUhostDisk(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			w := ctx.ProgressWriter()
			text := fmt.Sprintf("utidb[%s] is resizing disk", pickedID)
			ctx.PollerTo(w, describeByID(ctx, req.GetRegion(), req.GetZone(), req.GetProjectId())).Spoll(pickedID, text, []string{stateRunning, stateUpgradeFail})
			ctx.EmitResult(cli.OpResultRow{ResourceID: pickedID, Action: "resize-disk", Status: "Resizing"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&id, "utidb-id", "", "Required. Resource ID of the UTiDB instance")
	flags.StringVar(&scaleType, "scale-type", "", "Required. Scale type: SCALEOUT or SCALEIN")
	flags.StringVar(&nodeConfig, "node-config", "", "Required. Node config, format: DiskSize=N,ServerType=tidb")
	flags.IntVar(&startTime, "start-time", 0, "Optional. Task start time")

	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("utidb-id")
	cmd.MarkFlagRequired("scale-type")
	cmd.MarkFlagRequired("node-config")
	command.SetFlagValues(cmd, "scale-type", "SCALEOUT", "SCALEIN")
	command.SetCompletion(cmd, "utidb-id", func() []string {
		return listResourceIDs(ctx, nil, *req.Region, *req.Zone, *req.ProjectId)
	})

	return cmd
}
