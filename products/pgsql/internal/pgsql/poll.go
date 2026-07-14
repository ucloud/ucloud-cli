package pgsql

import (
	"fmt"
	"io"

	"github.com/ucloud/ucloud-sdk-go/services/upgsql"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// stopPgsqlIns stops the instance and narrates progress to out (the caller passes
// ctx.ProgressWriter(): stdout in table mode, stderr in json/yaml). Returns the
// stop error so callers can decide whether to record a structured result.
// Mirrors products/mysql/internal/mysql/poll.go stopUdbIns.
func stopPgsqlIns(ctx *cli.Context, req *upgsql.StopUPgSQLInstanceRequest, async bool, out io.Writer) error {
	client := newUPgSQLClient(ctx)
	_, err := client.StopUPgSQLInstance(req)
	if err != nil {
		ctx.HandleError(err)
		return err
	}
	text := fmt.Sprintf("pgsql[%s] is stopping", *req.InstanceID)
	if async {
		fmt.Fprintln(out, text)
	} else {
		ctx.PollerTo(out, describePgsqlByID(ctx)).Spoll(*req.InstanceID, text, []string{PGSQL_STOPPED, PGSQL_SHUTDOWN_FAILED})
	}
	return nil
}

// describePgsqlByID returns the poller's describe func, closing over ctx so it
// can build an authed upgsql client. Mirrors products/mysql/internal/mysql/poll.go
// describeUdbByID, but uses GetUPgSQLInstance (single-instance describe).
func describePgsqlByID(ctx *cli.Context) func(instanceID string, commonBase *request.CommonBase) (interface{}, error) {
	return func(instanceID string, commonBase *request.CommonBase) (interface{}, error) {
		client := newUPgSQLClient(ctx)
		req := client.NewGetUPgSQLInstanceRequest()
		if commonBase != nil {
			req.CommonBase = *commonBase
		}
		req.InstanceID = sdk.String(instanceID)
		resp, err := client.GetUPgSQLInstance(req)
		if err != nil {
			return nil, err
		}
		if resp.DataSet.InstanceID == "" {
			return nil, fmt.Errorf("pgsql[%s] may not exist", instanceID)
		}
		return &resp.DataSet, nil
	}
}
