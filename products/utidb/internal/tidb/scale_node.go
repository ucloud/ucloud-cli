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
	if err := validateServerType(*cfg.ServerType); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// newScaleNode ucloud utidb scale-node
func newScaleNode(ctx *cli.Context) *cobra.Command {
	var id, scaleType, nodeConfig, serverID string
	var startTime int
	var async bool

	client := cli.NewServiceClient(ctx, tidb.NewClient)
	req := client.NewModifyTiDBClusterNodeRequest()

	cmd := &cobra.Command{
		Use:   "scale-node",
		Short: "Scale nodes of a UTiDB instance (SCALEOUT/SCALEIN)",
		Long:  helpScaleNodeLong,
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
			params := mergeCommonParams(req.GetRegion(), req.GetZone(), req.GetProjectId(), map[string]interface{}{
				"Id":        pickedID,
				"ScaleType": scaleType,
			})
			params["NodeConfig"] = scaleNodeConfigToMap(cfg)
			if serverID != "" {
				params["ServerId"] = serverID
			}
			if startTime != 0 {
				params["StartTime"] = startTime
			}

			_, err = invokeAPI(ctx, "ModifyTiDBClusterNode", params)
			if err != nil {
				handleAPIError(ctx, err)
				return
			}

			w := ctx.ProgressWriter()
			if async {
				fmt.Fprintf(w, "utidb[%s] is scaling nodes\n", pickedID)
			} else {
				text := fmt.Sprintf("utidb[%s] is scaling nodes", pickedID)
				spollUpgrade(ctx, w, req.GetRegion(), req.GetZone(), req.GetProjectId(), pickedID, text)
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: pickedID, Action: "scale-node", Status: "Scaling"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&id, "utidb-id", "", "Required. Resource ID of the UTiDB instance")
	flags.StringVar(&scaleType, "scale-type", "", "Required. SCALEOUT (expand) or SCALEIN (shrink; requires --server-id)")
	flags.StringVar(&nodeConfig, "node-config", "", "Required. ConfigId=xxx,NodeCount=N,ServerType=tidb|tikv|pd|tiflash (target count after scale)")
	flags.StringVar(&serverID, "server-id", "", "Required for SCALEIN. Server ID of the node to remove (tab completion lists cluster nodes)")
	flags.IntVar(&startTime, "start-time", 0, "Optional. Task start time")
	flags.BoolVar(&async, "async", false, "Optional. Do not wait for scaling to finish")

	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("utidb-id")
	cmd.MarkFlagRequired("scale-type")
	cmd.MarkFlagRequired("node-config")
	command.SetFlagValues(cmd, "scale-type", "SCALEOUT", "SCALEIN")
	command.SetCompletion(cmd, "utidb-id", func() []string {
		return listResourceIDs(ctx, nil, req.GetRegion(), req.GetZone(), req.GetProjectId())
	})
	command.SetCompletion(cmd, "server-id", func() []string {
		return listServerIDs(ctx, req.GetRegion(), req.GetZone(), req.GetProjectId(), ctx.PickResourceID(id))
	})

	return cmd
}
