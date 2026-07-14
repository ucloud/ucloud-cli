package pgsql

import (
	"fmt"

	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newUSupabaseClient returns an authed generic SDK client whose requests are
// JSON-encoded, reusing the platform credential/handlers (so AK/SK and OAuth
// profiles both sign correctly, and project-id normalization / OAuth retry
// apply uniformly). ucloud-sdk-go has no typed USupabase methods, so supabase
// actions are invoked generically: a simple map payload signed and POSTed as a
// JSON body. JSON (not form) is mandatory — the USupabase gateway does
// json.Unmarshal on the request and fails on string-into-int fields (same
// 214001 bug as UPgSQL).
func newUSupabaseClient(ctx *cli.Context) *uaccount.UAccountClient {
	client := cli.NewServiceClient(ctx, uaccount.NewClient)
	_ = client.AddRequestHandler(func(c *sdk.Client, req request.Common) (request.Common, error) {
		req.SetEncoder(request.NewJSONEncoder(c.GetConfig(), c.GetCredential()))
		return req, nil
	})
	return client
}

// invokeSupabase calls a USupabase action with a simple map payload and returns
// the response payload map. Region/Zone/ProjectId/IsMemoryDB and business fields
// are all part of params; the caller is responsible for putting them in. The
// SDK signs the map (cred.Apply adds PublicKey + Signature) and the JSONEncoder
// marshals it with native types. A non-zero RetCode is returned as an error by
// GenericInvoke (RetCodePatcher), so callers forward it to ctx.HandleError.
func invokeSupabase(ctx *cli.Context, action string, params map[string]interface{}) (map[string]interface{}, error) {
	client := newUSupabaseClient(ctx)
	req := client.NewGenericRequest()
	payload := make(map[string]interface{}, len(params)+1)
	payload["Action"] = action
	for k, v := range params {
		payload[k] = v
	}
	if err := req.SetPayload(payload); err != nil {
		return nil, fmt.Errorf("set payload: %w", err)
	}
	resp, err := client.GenericInvoke(req)
	if err != nil {
		return nil, err
	}
	return resp.GetPayload(), nil
}

// supabaseState is a typed view over a generic response's State field, used to
// avoid repeated map[string]interface{} casts in command code.
type supabaseState struct {
	State string
}
