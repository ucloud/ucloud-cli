package cmd

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/base"
)

var ulhostSpoller = base.NewSpoller(sdescribeULHostByID, base.Cxt.GetWriter())

func sdescribeULHostByID(ulhostID string, common *request.CommonBase) (interface{}, error) {
	req := base.BizClient.UCompShareClient.NewDescribeULHostInstanceRequest()
	req.ULHostIds = []string{ulhostID}
	if common != nil {
		req.CommonBase = *common
	}
	resp, err := base.BizClient.UCompShareClient.DescribeULHostInstance(req)
	if err != nil {
		return nil, err
	}
	if len(resp.ULHostInstanceSets) < 1 {
		return nil, nil
	}

	return &resp.ULHostInstanceSets[0], nil
}
