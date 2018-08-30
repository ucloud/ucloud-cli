package uhost

import (
	"time"

	uerr "github.com/ucloud/ucloud-sdk-go/sdk/error"
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/utils"
)

type WaitUntilUHostInstanceStateRequest struct {
	request.CommonBase

	Interval        time.Duration
	MaxAttempts     int
	DescribeRequest *DescribeUHostInstanceRequest
	State           string
	IgnoreError     bool
}

// NewWaitUntilUHostInstanceStateRequest will create request of WaitUntilUHostInstanceState action.
func (c *UHostClient) NewWaitUntilUHostInstanceStateRequest() *WaitUntilUHostInstanceStateRequest {
	cfg := c.client.GetConfig()

	return &WaitUntilUHostInstanceStateRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

// WaitUntilUHostInstanceState will pending current goroutine until the state has changed to expected state.
func (c *UHostClient) WaitUntilUHostInstanceState(req *WaitUntilUHostInstanceStateRequest) error {
	waiter := utils.FuncWaiter{
		Interval:    req.Interval,
		MaxAttempts: req.MaxAttempts,
		IgnoreError: req.IgnoreError,
		Checker: func() (bool, error) {
			resp, err := c.DescribeUHostInstance(req.DescribeRequest)
			if err != nil {
				switch err {
				case uerr.InvalidRequestError:
					return false, err
				default:
					return false, nil
				}
			}

			// TODO: Ensure if it is any data consistency problem?
			// Such as creating a new uhost, but cannot describe it's correct state immediately ...
			for _, uhost := range resp.UHostSet {
				if uhost.State != req.State {
					return false, nil
				}
			}

			if len(resp.UHostSet) > 0 {
				return true, nil
			} else {
				return false, nil
			}
		},
	}
	return waiter.WaitForCompletion()
}
