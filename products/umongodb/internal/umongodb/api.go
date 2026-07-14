package umongodb

import (
	"fmt"
	"time"

	"github.com/ucloud/ucloud-sdk-go/services/uaccount"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// defaultCallTimeout is the HTTP timeout for GenericInvoke calls.
// MongoDB creation APIs are slow; 5 minutes avoids premature timeouts.
const defaultCallTimeout = 5 * time.Minute

// genericCall invokes a UMongoDB API via GenericInvoke and returns the
// raw payload map. All MongoDB commands use this instead of typed SDK methods.
func genericCall(ctx *cli.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
	client := cli.NewServiceClient(ctx, uaccount.NewClient)
	req := client.NewGenericRequest()
	if err := req.SetPayload(params); err != nil {
		return nil, fmt.Errorf("set payload for %s: %w", action, err)
	}
	req.WithTimeout(defaultCallTimeout)
	resp, err := client.GenericInvoke(req)
	if err != nil {
		return nil, err
	}
	payload := resp.GetPayload()
	retCode, _ := payload["RetCode"].(float64)
	if retCode != 0 {
		msg, _ := payload["Message"].(string)
		if msg == "" {
			msg = fmt.Sprintf("API %s returned RetCode %v", action, retCode)
		}
		return nil, fmt.Errorf("%s: %s", action, msg)
	}
	return payload, nil
}
