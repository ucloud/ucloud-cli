package cloudwatch

import (
	"encoding/json"
	"fmt"

	sdkcloudwatch "github.com/ucloud/ucloud-sdk-go/services/cloudwatch"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newGenericClient returns the authed CloudWatch service client used purely as
// the carrier for GenericInvoke. The SDK does provide a strongly-typed
// CloudWatchClient (services/cloudwatch), but this product intentionally does
// NOT call its typed methods — it only borrows the client (and the platform
// credential/signature/handler chain cli.NewServiceClient wires up) to send
// generic Action requests, exactly like `ucloud api` and the mysql product do
// for actions with no typed SDK method. Using the cloudwatch client (rather
// than uaccount) makes the call site read as "this is a CloudWatch request"
// without coupling to the SDK's generated request/response types.
func newGenericClient(ctx *cli.Context) *sdkcloudwatch.CloudWatchClient {
	return cli.NewServiceClient(ctx, sdkcloudwatch.NewClient)
}

// invoke sends req with the given payload and returns the decoded SkymFlameAPI
// envelope payload (map: {Action, TraceId, RetCode, Message, Data, TotalCount?}).
//
// Business errors (envelope RetCode != 0) do NOT need to be checked here: the
// SDK's built-in errorHandler (ucloud/handlers.go, registered by default on
// every *ucloud.Client) already inspects resp.GetRetCode() after every
// InvokeAction call and converts a non-zero RetCode into a uerr.Error, which
// GenericInvoke returns as err. So `err != nil` from GenericInvoke already
// covers both transport errors (network/timeout/signature) and business
// errors — callers only need ctx.HandleError(err); there is nothing left for
// product code to inspect on the payload for error purposes.
func invoke(client *sdkcloudwatch.CloudWatchClient, req request.GenericRequest, payload map[string]interface{}) (map[string]interface{}, error) {
	if err := req.SetPayload(payload); err != nil {
		return nil, fmt.Errorf("set payload: %w", err)
	}
	resp, err := client.GenericInvoke(req)
	if err != nil {
		return nil, err
	}
	return resp.GetPayload(), nil
}

// decodeData decodes the envelope's Data field into out (a pointer to the
// caller's local response struct) by re-marshaling the interface{} value.
func decodeData(payload map[string]interface{}, out interface{}) error {
	raw, err := json.Marshal(payload["Data"])
	if err != nil {
		return fmt.Errorf("marshal Data: %w", err)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("unmarshal Data: %w", err)
	}
	return nil
}
