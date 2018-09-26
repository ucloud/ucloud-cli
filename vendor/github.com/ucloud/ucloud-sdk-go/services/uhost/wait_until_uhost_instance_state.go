package uhost

import (
	"time"

	"github.com/ucloud/ucloud-sdk-go/sdk"
	uerr "github.com/ucloud/ucloud-sdk-go/sdk/error"
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/utils"
)

type WaitUntilUHostInstanceStateRequest struct {
	request.CommonBase

	Interval        *time.Duration
	MaxAttempts     *int
	DescribeRequest *DescribeUHostInstanceRequest
	State           State
	IgnoreError     *bool
}

// NewWaitUntilUHostInstanceStateRequest will create request of WaitUntilUHostInstanceState action.
func (c *UHostClient) NewWaitUntilUHostInstanceStateRequest() *WaitUntilUHostInstanceStateRequest {
	cfg := c.client.GetConfig()

	return &WaitUntilUHostInstanceStateRequest{
		CommonBase: request.CommonBase{
			Region:    sdk.String(cfg.Region),
			ProjectId: sdk.String(cfg.ProjectId),
		},
	}
}

// WaitUntilUHostInstanceState will pending current goroutine until the state has changed to expected state.
func (c *UHostClient) WaitUntilUHostInstanceState(req *WaitUntilUHostInstanceStateRequest) error {
	waiter := utils.FuncWaiter{
		Interval:    sdk.TimeDurationValue(req.Interval),
		MaxAttempts: sdk.IntValue(req.MaxAttempts),
		IgnoreError: sdk.BoolValue(req.IgnoreError),
		Checker: func() (bool, error) {
			resp, err := c.DescribeUHostInstance(req.DescribeRequest)

			if err != nil {
				skipErrors := []string{uerr.ErrNetwork, uerr.ErrHTTPStatus, uerr.ErrRetCode}
				if uErr, ok := err.(uerr.Error); ok && utils.IsStringIn(uErr.Name(), skipErrors) {
					return false, nil
				}
				return false, err
			}

			// TODO: Ensure if it is any data consistency problem?
			// Such as creating a new uhost, but cannot describe it's correct state immediately ...
			for _, uhost := range resp.UHostSet {
				if val, _ := req.State.MarshalValue(); uhost.State != val {
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
