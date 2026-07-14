package redis

import (
	"fmt"

	"github.com/ucloud/ucloud-sdk-go/services/umem"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

type redisMode string

const (
	redisModeUnknown       redisMode = ""
	redisModeMasterReplica redisMode = "master-replica"
	redisModeDistributed   redisMode = "distributed"
)

func resourceTypeToMode(resourceType string) redisMode {
	switch resourceType {
	case "g4v6", "single":
		return redisModeMasterReplica
	case "performance", "cluster":
		return redisModeDistributed
	}
	return redisModeUnknown
}

func describeRedisMode(ctx *cli.Context, id string) (redisMode, error) {
	client := cli.NewServiceClient(ctx, umem.NewClient)
	req := client.NewDescribeUMemRequest()
	req.Protocol = sdk.String("redis")
	req.ResourceId = &id
	resp, err := client.DescribeUMem(req)
	if err != nil {
		return redisModeUnknown, err
	}
	if len(resp.DataSet) < 1 {
		return redisModeUnknown, fmt.Errorf("resource [%s] may not exist", id)
	}
	return resourceTypeToMode(resp.DataSet[0].ResourceType), nil
}
