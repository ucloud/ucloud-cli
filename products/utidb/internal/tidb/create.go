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

// parseCreateNodeConfig parses a CLI node-config string into the SDK type.
// Format: ConfigId=xxx,DiskSize=N,NodeCount=N,ServerType=tidb
func parseCreateNodeConfig(s string) (tidb.CreateTiDBClusterServiceParamNodeConfig, error) {
	var cfg tidb.CreateTiDBClusterServiceParamNodeConfig
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
		case "DiskSize":
			n, err := strconv.Atoi(val)
			if err != nil {
				return cfg, fmt.Errorf("invalid DiskSize %q: %w", val, err)
			}
			cfg.DiskSize = sdk.Int(n)
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
	if cfg.ConfigId == nil || cfg.DiskSize == nil || cfg.NodeCount == nil || cfg.ServerType == nil {
		return cfg, fmt.Errorf("node-config must include ConfigId, DiskSize, NodeCount and ServerType")
	}
	if err := validateServerType(*cfg.ServerType); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func parseCreateLabels(ss []string) []tidb.CreateTiDBClusterServiceParamLabels {
	var out []tidb.CreateTiDBClusterServiceParamLabels
	for _, s := range ss {
		parts := strings.SplitN(s, "=", 2)
		if len(parts) == 2 {
			out = append(out, tidb.CreateTiDBClusterServiceParamLabels{
				Key:   sdk.String(strings.TrimSpace(parts[0])),
				Value: sdk.String(strings.TrimSpace(parts[1])),
			})
		}
	}
	return out
}

func parseCreateSecGroupInfo(ss []string) ([]tidb.CreateTiDBClusterServiceParamSecGroupInfo, error) {
	var out []tidb.CreateTiDBClusterServiceParamSecGroupInfo
	for _, s := range ss {
		var item tidb.CreateTiDBClusterServiceParamSecGroupInfo
		parts := strings.Split(s, ",")
		for _, part := range parts {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) != 2 {
				continue
			}
			key := strings.TrimSpace(kv[0])
			val := strings.TrimSpace(kv[1])
			switch key {
			case "SecGroupId":
				item.SecGroupId = sdk.String(val)
			case "Priority":
				n, err := strconv.Atoi(val)
				if err != nil {
					return nil, fmt.Errorf("invalid Priority %q: %w", val, err)
				}
				item.Priority = sdk.Int(n)
			}
		}
		out = append(out, item)
	}
	return out, nil
}

// newCreate ucloud utidb create
func newCreate(ctx *cli.Context) *cobra.Command {
	var name, password, chargeType, dtType, pubUlbID, vpcID, subnetID string
	var dbVersion, ip, port, coupon, promotionID, templateID string
	var quantity float64
	var activityID, ruleID int
	var alertStrategyIDs []int
	var labels, secGroupInfo []string
	var nodeConfigs []string
	var async bool

	client := cli.NewServiceClient(ctx, tidb.NewClient)
	req := client.NewCreateTiDBClusterServiceRequest()

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a UTiDB instance",
		Long:  helpCreateLong,
		Run: func(c *cobra.Command, args []string) {
			var configs []tidb.CreateTiDBClusterServiceParamNodeConfig
			for _, s := range nodeConfigs {
				cfg, err := parseCreateNodeConfig(s)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				configs = append(configs, cfg)
			}

			params := mergeCommonParams(req.GetRegion(), req.GetZone(), req.GetProjectId(), map[string]interface{}{
				"Name":       name,
				"Password":   password,
				"ChargeType": chargeType,
				"DTType":     dtType,
				"VPCId":      vpcID,
				"SubnetId":   subnetID,
				"Quantity":   quantity,
			})
			if pubUlbID != "" {
				params["PubUlbId"] = pubUlbID
			}
			if dbVersion != "" {
				params["DbVersion"] = dbVersion
			}
			if ip != "" {
				params["Ip"] = ip
			}
			if port != "" {
				params["Port"] = port
			}
			if promotionID != "" {
				params["PromotionId"] = promotionID
			}
			if templateID != "" {
				params["TemplateId"] = templateID
			}
			if activityID != 0 {
				params["ActivityId"] = activityID
			}
			if ruleID != 0 {
				params["RuleId"] = ruleID
			}
			if coupon != "" {
				params["Coupon"] = coupon
			}
			if len(alertStrategyIDs) > 0 {
				params["AlertStrategyIds"] = alertStrategyIDs
			}
			if len(labels) > 0 {
				labelMaps := make([]map[string]interface{}, 0, len(labels))
				for _, l := range parseCreateLabels(labels) {
					labelMaps = append(labelMaps, labelToMap(l))
				}
				params["Labels"] = labelMaps
			}
			if len(secGroupInfo) > 0 {
				infos, err := parseCreateSecGroupInfo(secGroupInfo)
				if err != nil {
					ctx.HandleError(err)
					return
				}
				secMaps := make([]map[string]interface{}, 0, len(infos))
				for _, info := range infos {
					secMaps = append(secMaps, secGroupToMap(info))
				}
				params["SecGroupInfo"] = secMaps
			}
			nodeConfigMaps := make([]map[string]interface{}, 0, len(configs))
			for _, cfg := range configs {
				nodeConfigMaps = append(nodeConfigMaps, createNodeConfigToMap(cfg))
			}
			params["NodeConfig"] = nodeConfigMaps

			payload, err := invokeAPI(ctx, "CreateTiDBClusterService", params)
			if err != nil {
				handleAPIError(ctx, err)
				return
			}
			data, _ := payload["Data"].(map[string]interface{})
			clusterID := stringVal(data["Id"])
			if clusterID == "" {
				ctx.HandleError(fmt.Errorf("empty cluster ID in response"))
				return
			}

			w := ctx.ProgressWriter()
			if async {
				fmt.Fprintf(w, "utidb[%s] is creating\n", clusterID)
			} else {
				text := fmt.Sprintf("utidb[%s] is creating", clusterID)
				spollCreate(ctx, w, req.GetRegion(), req.GetZone(), req.GetProjectId(), clusterID, text)
			}
			ctx.EmitResult(cli.OpResultRow{ResourceID: clusterID, Action: "create", Status: "Creating"})
		},
	}

	flags := cmd.Flags()
	flags.SortFlags = false

	flags.StringVar(&name, "name", "", "Required. Instance name")
	flags.StringVar(&password, "password", "", "Required. Admin password")
	flags.StringVar(&chargeType, "charge-type", "", "Required. Charge type: Month, Year, Dynamic, Trial")
	flags.StringVar(&dtType, "dt-type", "", "Required. Disaster tolerance: 10 (same AZ), 20 (cross AZ)")
	flags.StringVar(&pubUlbID, "pub-ulb-id", "", "Optional. Public ULB ID")
	flags.StringVar(&vpcID, "vpc-id", "", "Required. VPC ID")
	flags.StringVar(&subnetID, "subnet-id", "", "Required. Subnet ID")
	flags.Float64Var(&quantity, "quantity", 1, "Required. Purchase duration")
	flags.StringArrayVar(&nodeConfigs, "node-config", nil, "Required. Per node type: ConfigId=xxx,DiskSize=N,NodeCount=N,ServerType=tidb|tikv|pd|tiflash")

	flags.StringVar(&dbVersion, "db-version", "", "Optional. Database version, e.g. v8.5.1, v8.5.6")
	flags.StringVar(&ip, "ip", "", "Optional. Specified IP address")
	flags.StringVar(&port, "port", "", "Optional. Specified port")
	flags.StringVar(&coupon, "coupon", "", "Optional. Coupon ID")
	flags.StringVar(&promotionID, "promotion-id", "", "Optional. Promotion ID")
	flags.StringVar(&templateID, "template-id", "", "Optional. Parameter template ID")
	flags.IntVar(&activityID, "activity-id", 0, "Optional. Activity ID")
	flags.IntVar(&ruleID, "rule-id", 0, "Optional. Rule ID")
	flags.IntSliceVar(&alertStrategyIDs, "alert-strategy-ids", nil, "Optional. Alert strategy IDs")
	flags.StringSliceVar(&labels, "labels", nil, "Optional. Resource labels, format: key=value, repeatable")
	flags.StringArrayVar(&secGroupInfo, "sec-group-info", nil, "Optional. Security group info, format: SecGroupId=xxx,Priority=N, repeatable")
	flags.BoolVar(&async, "async", false, "Optional. Do not wait for creation to finish")

	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)
	ctx.BindProjectID(cmd, req)

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("charge-type")
	cmd.MarkFlagRequired("dt-type")
	cmd.MarkFlagRequired("vpc-id")
	cmd.MarkFlagRequired("subnet-id")
	cmd.MarkFlagRequired("quantity")
	cmd.MarkFlagRequired("node-config")

	command.SetFlagValues(cmd, "charge-type", "Month", "Year", "Dynamic", "Trial")
	command.SetFlagValues(cmd, "dt-type", "10", "20")

	return cmd
}
