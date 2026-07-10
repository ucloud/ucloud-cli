package tidb

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/services/tidb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// parseModifySpecNodeConfig parses the CLI node-config string for modify-spec.
// Format: ConfigId=xxx,ServerType=tidb
func parseModifySpecNodeConfig(s string) (tidb.ModifyTiDBClusterUhostSpecsParamNodeConfig, error) {
	var cfg tidb.ModifyTiDBClusterUhostSpecsParamNodeConfig
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
		case "ServerType":
			cfg.ServerType = sdk.String(val)
		default:
			return cfg, fmt.Errorf("unknown node-config key %q", key)
		}
	}
	if cfg.ConfigId == nil || cfg.ServerType == nil {
		return cfg, fmt.Errorf("node-config must include ConfigId and ServerType")
	}
	return cfg, nil
}

// newModifySpec ucloud utidb modify-spec
func newModifySpec(ctx *cli.Context) *cobra.Command {
	var id, nodeConfig string
	var startTime int

	client := cli.NewServiceClient(ctx, tidb.NewClient)
	req := client.NewModifyTiDBClusterUhostSpecsRequest()

	cmd := &cobra.Command{
		Use:   "modify-spec",
		Short: "Modify uhost specs of a UTiDB instance",
		Long:  "Modify uhost specs of a UTiDB instance",
		Run: func(c *cobra.Command, args []string) {
			cfg, err := parseModifySpecNodeConfig(nodeConfig)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			pickedID := ctx.PickResourceID(id)
			req.Id = sdk.String(pickedID)
			req.NodeConfig = &cfg
			if startTime != 0 {
				req.StartTime = sdk.Int(startTime)
			}

			_, err = client.ModifyTiDBClusterUhostSpecs(req)
			if err != nil {
				ctx.HandleError(err)
				return
			}

			w := ctx.ProgressWriter()
			text := fmt.Sprintf("utidb[%s] is modifying spec", pickedID)
			ctx.PollerTo(w, describeByID(ctx, req.GetRegion(), req.GetZone(), req.GetProjectId())).Spoll(pickedID, text, []string{stateRunning, stateUpgradeFail})
			ctx.EmitResult(cli.OpResultRow{ResourceID: pickedID, Action: "modify-spec", Status: "Modifying"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&id, "utidb-id", "", "Required. Resource ID of the UTiDB instance")
	flags.StringVar(&nodeConfig, "node-config", "", "Required. Node config, format: ConfigId=xxx,ServerType=tidb")
	flags.IntVar(&startTime, "start-time", 0, "Optional. Task start time")

	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("utidb-id")
	cmd.MarkFlagRequired("node-config")
	command.SetCompletion(cmd, "utidb-id", func() []string {
		return listResourceIDs(ctx, nil, *req.Region, *req.Zone, *req.ProjectId)
	})

	return cmd
}
