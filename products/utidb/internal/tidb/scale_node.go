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

// parseScaleNodeConfig parses the CLI node-config string for scale-node.
// Format: ConfigId=xxx,NodeCount=N,ServerType=tidb
func parseScaleNodeConfig(s string) (tidb.ModifyTiDBClusterNodeParamNodeConfig, error) {
	var cfg tidb.ModifyTiDBClusterNodeParamNodeConfig
	parts := strings.Split(s, ",")
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			return cfg, fmt.Errorf("invalid node-config segment %q, expected key=value", part)
		}
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		switch key {
		case "ConfigId":
			cfg.ConfigId = sdk.String(val)
		case "NodeCount":
			n, err := strconv.Atoi(val)
			if err != nil {
				return cfg, fmt.Errorf("invalid NodeCount %q: %w", val, err)
			}
			cfg.NodeCount = sdk.Int(n)
		case "ServerType":
			cfg.ServerType = sdk.String(val)
		default:
			return cfg, fmt.Errorf("unknown node-config key %q", key)
		}
	}
	if cfg.ConfigId == nil || cfg.NodeCount == nil || cfg.ServerType == nil {
		return cfg, fmt.Errorf("node-config must include ConfigId, NodeCount and ServerType")
	}
	return cfg, nil
}

// newScaleNode ucloud utidb scale-node
func newScaleNode(ctx *cli.Context) *cobra.Command {
	var id, scaleType, nodeConfig, serverID string
	var startTime int

	client := cli.NewServiceClient(ctx, tidb.NewClient)
	req := client.NewModifyTiDBClusterNodeRequest()

	cmd := &cobra.Command{
		Use:   "scale-node",
		Short: "Scale nodes of a UTiDB instance (SCALEOUT/SCALEIN)",
		Long:  "Scale nodes of a UTiDB instance (SCALEOUT/SCALEIN)",
		Run: func(c *cobra.Command, args []string) {
			if scaleType == "SCALEIN" && serverID == "" {
				ctx.HandleError(fmt.Errorf("server-id is required when scale-type is SCALEIN"))
				return
			}

			cfg, err := parseScaleNodeConfig(nodeConfig)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			pickedID := ctx.PickResourceID(id)
			req.Id = sdk.String(pickedID)
			req.ScaleType = sdk.String(scaleType)
			req.NodeConfig = &cfg
			if serverID != "" {
				req.ServerId = sdk.String(serverID)
			}
			if startTime != 0 {
				req.StartTime = sdk.Int(startTime)
			}

			_, err = client.ModifyTiDBClusterNode(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			w := ctx.ProgressWriter()
			text := fmt.Sprintf("utidb[%s] is scaling nodes", pickedID)
			ctx.PollerTo(w, describeByID(ctx, req.GetRegion(), req.GetZone(), req.GetProjectId())).Spoll(pickedID, text, []string{stateRunning, stateUpgradeFail})
			ctx.EmitResult(cli.OpResultRow{ResourceID: pickedID, Action: "scale-node", Status: "Scaling"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&id, "utidb-id", "", "Required. Resource ID of the UTiDB instance")
	flags.StringVar(&scaleType, "scale-type", "", "Required. Scale type: SCALEOUT or SCALEIN")
	flags.StringVar(&nodeConfig, "node-config", "", "Required. Node config, format: ConfigId=xxx,NodeCount=N,ServerType=tidb")
	flags.StringVar(&serverID, "server-id", "", "Optional. Server ID to scale in, required when scale-type=SCALEIN")
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
	command.SetCompletion(cmd, "server-id", func() []string {
		return listServerIDs(ctx, ctx.PickResourceID(id))
	})

	return cmd
}
