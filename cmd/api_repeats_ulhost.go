package cmd

import (
	"io"

	"github.com/ucloud/ucloud-sdk-go/services/ucompshare"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newULHostPoller supports `ucloud api --Action CreateULHostInstance --repeats`.
func newULHostPoller(out io.Writer) cli.Poller {
	return cli.NewPoller(sdescribeULHostByID, out)
}

func sdescribeULHostByID(ulhostID string, common *request.CommonBase) (interface{}, error) {
	client := newServiceClient(ucompshare.NewClient)
	req := client.NewDescribeULHostInstanceRequest()
	req.ULHostIds = []string{ulhostID}
	if common != nil {
		req.CommonBase = *common
	}
	resp, err := client.DescribeULHostInstance(req)
	if err != nil {
		return nil, err
	}
	if len(resp.ULHostInstanceSets) < 1 {
		return nil, nil
	}

	return &resp.ULHostInstanceSets[0], nil
}
