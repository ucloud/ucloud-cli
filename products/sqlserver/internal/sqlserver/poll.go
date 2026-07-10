package sqlserver

import (
	"fmt"
	"io"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// stopUdbIns stops the instance and narrates progress to out (the caller passes
// ctx.ProgressWriter(): stdout in table mode, stderr in json/yaml). Returns the
// stop error so callers can decide whether to record a structured result.
func stopUdbIns(ctx *cli.Context, req *udb.StopUDBInstanceRequest, async bool, out io.Writer) error {
	client := cli.NewServiceClient(ctx, udb.NewClient)
	_, err := client.StopUDBInstance(req)
	if err != nil {
		ctx.HandleError(err)
		return err
	}
	text := fmt.Sprintf("udb[%s] is stopping", *req.DBId)
	if async {
		fmt.Fprintln(out, text)
	} else {
		ctx.PollerTo(out, describeUdbByID(ctx)).Spoll(*req.DBId, text, []string{UDB_SHUTOFF, UDB_FAIL})
	}
	return nil
}

// describeUdbByID returns the poller's describe func, closing over ctx so it
// can build an authed udb client.
func describeUdbByID(ctx *cli.Context) func(udbID string, commonBase *request.CommonBase) (interface{}, error) {
	return func(udbID string, commonBase *request.CommonBase) (interface{}, error) {
		client := cli.NewServiceClient(ctx, udb.NewClient)
		req := client.NewDescribeUDBInstanceRequest()
		if commonBase != nil {
			req.CommonBase = *commonBase
		}
		req.DBId = sdk.String(udbID)
		resp, err := client.DescribeUDBInstance(req)
		if err != nil {
			return nil, err
		}
		if len(resp.DataSet) < 1 {
			return nil, fmt.Errorf("udb[%s] may not exist", udbID)
		}
		return &resp.DataSet[0], nil
	}
}
