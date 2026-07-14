package tidb

import (
	"fmt"

	"github.com/ucloud/ucloud-sdk-go/services/tidb"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// invokeAPI calls a TiDB API action via GenericRequest.
//
// Why GenericInvoke instead of typed NewXxxRequest() field assignment (§5):
// the SDK FormEncoder emits string slices as NodeTypes.0, but TiDB APIs expect
// flat NodeTypes0; nested NodeConfig / Labels / SecGroupInfo must be nested
// maps so the encoder emits NodeConfig.0.Field. Typed requests cannot express
// both encodings correctly for this product.
func invokeAPI(ctx *cli.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
	client := cli.NewServiceClient(ctx, tidb.NewClient)
	req := client.NewGenericRequest()
	allParams := map[string]interface{}{
		"Action": action,
	}
	for k, v := range params {
		allParams[k] = v
	}
	if err := req.SetPayload(allParams); err != nil {
		return nil, fmt.Errorf("set payload: %w", err)
	}
	resp, err := client.GenericInvoke(req)
	if err != nil {
		return nil, err
	}
	return resp.GetPayload(), nil
}

func mergeCommonParams(region, zone, projectID string, params map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(params)+3)
	for k, v := range params {
		out[k] = v
	}
	if region != "" {
		out["Region"] = region
	}
	if zone != "" {
		out["Zone"] = zone
	}
	if projectID != "" {
		out["ProjectId"] = projectID
	}
	return out
}

func flattenIndexedStrings(params map[string]interface{}, name string, values []string) {
	for i, v := range values {
		params[fmt.Sprintf("%s%d", name, i)] = v
	}
}

func getTiDBClusterUhostSpecs(ctx *cli.Context, region, zone, projectID string, nodeTypes []string) ([]tidb.UhostSpecs, error) {
	params := mergeCommonParams(region, zone, projectID, map[string]interface{}{})
	flattenIndexedStrings(params, "NodeTypes", formatNodeTypes(nodeTypes))
	payload, err := invokeAPI(ctx, "GetTiDBClusterUhostSpecs", params)
	if err != nil {
		return nil, err
	}
	return parseUhostSpecsFromPayload(payload), nil
}

func getTiDBClusterPayload(ctx *cli.Context, region, zone, projectID, id string) (map[string]interface{}, error) {
	params := mergeCommonParams(region, zone, projectID, map[string]interface{}{
		"Id": id,
	})
	return invokeAPI(ctx, "GetTiDBClusterService", params)
}
