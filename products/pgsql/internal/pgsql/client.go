package pgsql

import (
	"github.com/ucloud/ucloud-sdk-go/services/upgsql"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// newUPgSQLClient returns an authed UPgSQL client whose requests are encoded as
// JSON bodies instead of form-urlencoded.
//
// The UPgSQL gateway cannot unmarshal form-urlencoded string values into Go
// int/bool request fields (e.g. ListUPgSQLParamTemplate.Count), returning
// RetCode 214001 "json: cannot unmarshal string into Go struct field ... of
// type int". The SDK default form encoder serializes *int/*bool as the strings
// "100"/"false", which trips this. Switching to NewJSONEncoder keeps numeric
// and boolean fields as JSON numbers/booleans, which the gateway accepts.
//
// The encoder is swapped per-request via an SDK request handler (runs before
// buildHTTPRequest) using the SAME config+credential the default form encoder
// would use, so signing is unchanged. This covers every typed call that goes
// through this client, including the completion helpers.
//
// Note: this works for AK/SK profiles (Signature lives in the signed JSON
// body). OAuth profiles additionally need the platform cred-header injector to
// strip Signature/PublicKey from a JSON body (it currently only strips form
// bodies) — that is a separate platform-layer change; until then OAuth+pgsql
// remains non-functional (as it is today).
func newUPgSQLClient(ctx *cli.Context) *upgsql.UPgSQLClient {
	client := cli.NewServiceClient(ctx, upgsql.NewClient)
	_ = client.AddRequestHandler(func(c *sdk.Client, req request.Common) (request.Common, error) {
		req.SetEncoder(request.NewJSONEncoder(c.GetConfig(), c.GetCredential()))
		return req, nil
	})
	return client
}
