package cmd

import (
	"io"

	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/base"
)

// newULHostPoller supports `ucloud api --Action CreateULHostInstance --repeats`.
// It lives on the platform side (cmd/) because product packages must not import
// the legacy base package (hack/check-product rule2); the concurrent spoller's
// describe func uses base.BizClient, which is platform-owned glue.
func newULHostPoller(out io.Writer) *base.Poller {
	return base.NewSpoller(sdescribeULHostByID, out)
}

// sdescribeULHostByID is the concurrent-spoller describe func for ulhost
// creation polling. Nil-on-not-found and CommonBase-aware. Mirrors the legacy
// cmd/ulhost.go helper that was moved out of the product package.
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
