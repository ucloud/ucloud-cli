package mysql

import (
	"fmt"
	"io"

	"github.com/ucloud/ucloud-sdk-go/services/udb"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/model/status"
	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func stopUdbIns(ctx *cli.Context, req *udb.StopUDBInstanceRequest, async bool, out io.Writer) {
	client := cli.NewServiceClient(ctx, udb.NewClient)
	_, err := client.StopUDBInstance(req)
	if err != nil {
		ctx.HandleError(err)
		return
	}
	text := fmt.Sprintf("udb[%s] is stopping", *req.DBId)
	if async {
		fmt.Fprintln(out, text)
	} else {
		ctx.Poller(describeUdbByID(ctx)).Spoll(*req.DBId, text, []string{status.UDB_SHUTOFF, status.UDB_FAIL})
	}
}

func getUDBIDList(ctx *cli.Context, states []string, dbType, project, region, zone string) []string {
	udbs, err := getUDBList(ctx, states, dbType, project, region, zone)
	if err != nil {
		return nil
	}
	list := []string{}
	for _, db := range udbs {
		list = append(list, fmt.Sprintf("%s/%s", db.DBId, db.Name))
	}
	return list
}

func getUDBList(ctx *cli.Context, states []string, dbType, project, region, zone string) ([]udb.UDBInstanceSet, error) {
	client := cli.NewServiceClient(ctx, udb.NewClient)
	req := client.NewDescribeUDBInstanceRequest()
	if dbType == "" {
		dbType = "sql"
	}
	req.ClassType = &dbType
	req.ProjectId = &project
	req.Region = &region
	req.Zone = &zone
	list := []udb.UDBInstanceSet{}
	for offset, limit := 0, 50; ; offset += limit {
		req.Offset = sdk.Int(offset)
		req.Limit = sdk.Int(limit)
		resp, err := client.DescribeUDBInstance(req)
		if err != nil {
			return nil, err
		}
		for _, ins := range resp.DataSet {
			if states != nil {
				for _, s := range states {
					if s == ins.State {
						list = append(list, ins)
					}
				}
			} else {
				list = append(list, ins)
			}
		}
		if offset+limit >= resp.TotalCount {
			break
		}
	}
	return list, nil
}

// describeUdbByID returns the poller's describe func, closing over ctx so it
// can build an authed udb client. Mirrors cmd/mysql.go's describeUdbByID.
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
