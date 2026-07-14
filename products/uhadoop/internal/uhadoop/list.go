package uhadoop

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/ucloud/response"

	uhadoopsdk "github.com/ucloud/ucloud-sdk-go/services/uhadoop"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	sdkerror "github.com/ucloud/ucloud-sdk-go/ucloud/error"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// listClusterResponse mirrors the real API response for ListUHadoopInstance.
// The SDK's ListClusterInfo has CreateTime/ExpireTime as string but the API
// returns Unix timestamps (int), and misses fields like AutoRenew/HdfsTotal.
type listClusterResponse struct {
	response.CommonBase
	TotalCount int               `json:"TotalCount"`
	ClusterSet []listClusterInfo `json:"ClusterSet"`
}

type listClusterInfo struct {
	Zone               string `json:"Zone"`
	InstanceId         string `json:"InstanceId"`
	ClusterInstanceId  string `json:"ClusterInstanceId"`
	InstanceName       string `json:"InstanceName"`
	ClusterInstanceName string `json:"ClusterInstanceName"`
	FlinkResourceId    string `json:"FlinkResourceId"`
	Framework          string `json:"Framework"`
	FrameworkVersion   string `json:"FrameworkVersion"`
	Remark             string `json:"Remark"`
	CreateTime         int64  `json:"CreateTime"`
	ExpireTime         int64  `json:"ExpireTime"`
	AutoRenew          int    `json:"AutoRenew"`
	ChargeType         string `json:"ChargeType"`
	MasterCount        int    `json:"MasterCount"`
	CoreCount          int    `json:"CoreCount"`
	TaskCount          int    `json:"TaskCount"`
	UHostCount         int    `json:"UHostCount"`
	RedundantCount     int    `json:"RedundantCount"`
	State              string `json:"State"`
	ReleaseVersion     string `json:"ReleaseVersion"`
	HadoopVersion      string `json:"HadoopVersion"`
	VPCId              string `json:"VPCId"`
	SubnetId           string `json:"SubnetId"`
	BusinessId         string `json:"BusinessId"`
	HdfsTotal          int    `json:"HdfsTotal"`
	HdfsUsed           int    `json:"HdfsUsed"`
}

// newList ucloud uhadoop list
func newList(ctx *cli.Context) *cobra.Command {
	var allRegion, idOnly bool
	client := cli.NewServiceClient(ctx, uhadoopsdk.NewClient)
	req := client.NewListUHadoopInstanceRequest()
	cmd := &cobra.Command{
		Use:          "list",
		Short:        "List all UHadoop clusters",
		Long:         `List all UHadoop clusters`,
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			clusters, err := getAllClusters(ctx, client, req, allRegion)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			if idOnly {
				listClusterID(ctx, clusters)
			} else {
				listClusters(ctx, clusters, allRegion)
			}
		},
	}
	cmd.Flags().SortFlags = false

	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	req.Zone = cmd.Flags().String("zone", "", "Optional. Assign availability zone")
	req.Limit = cmd.Flags().Int("limit", 60, "Optional. Limit default 60")
	req.Offset = cmd.Flags().Int("offset", 0, "Optional. Offset default 0")
	cmd.Flags().BoolVar(&allRegion, "all-region", false, "Optional. Accept values: true or false. List clusters of all regions when assigned true")
	cmd.Flags().BoolVar(&idOnly, "id-only", false, "Optional. Just display resource id of clusters")

	command.SetFlagValues(cmd, "all-region", "true", "false")
	command.SetFlagValues(cmd, "id-only", "true", "false")

	return cmd
}

func getAllClusters(ctx *cli.Context, client *uhadoopsdk.UHadoopClient, req *uhadoopsdk.ListUHadoopInstanceRequest, allRegion bool) ([]listClusterInfo, error) {
	if allRegion {
		result := make([]listClusterInfo, 0)
		regions, err := ctx.AllRegions()
		if err != nil {
			return nil, err
		}
		for _, region := range regions {
			_req := *req
			_req.Region = sdk.String(region)
			clusters, err := fetchClustersPageOff(client, &_req)
			if e, ok := err.(sdkerror.Error); ok && e.Code() == _RetCodeRegionNoPermission {
				continue
			}
			if err != nil {
				return nil, err
			}
			result = append(result, clusters...)
		}
		return result, nil
	}

	var resp listClusterResponse
	// Use InvokeAction directly with our custom response struct because the
	// SDK's ListClusterInfo types CreateTime/ExpireTime as string (wrong: int).
	err := client.InvokeAction("ListUHadoopInstance", req, &resp)
	if err != nil {
		return nil, err
	}
	return resp.ClusterSet, nil
}

func fetchClustersPageOff(client *uhadoopsdk.UHadoopClient, req *uhadoopsdk.ListUHadoopInstanceRequest) ([]listClusterInfo, error) {
	_req := *req
	result := make([]listClusterInfo, 0)
	for limit, offset := 60, 0; ; offset += limit {
		_req.Offset = sdk.Int(offset)
		_req.Limit = sdk.Int(limit)
		var resp listClusterResponse
		err := client.InvokeAction("ListUHadoopInstance", &_req, &resp)
		if err != nil {
			return nil, err
		}
		result = append(result, resp.ClusterSet...)
		if len(resp.ClusterSet) < limit {
			break
		}
	}
	return result, nil
}

func listClusters(ctx *cli.Context, clusters []listClusterInfo, listAllRegion bool) {
	list := make([]listRow, 0, len(clusters))
	for _, c := range clusters {
		list = append(list, toListRow(c))
	}

	if ctx.Format() != cli.OutputTable {
		ctx.PrintList(list)
		return
	}

	rows := make([]listRowDefault, 0, len(list))
	for _, r := range list {
		rows = append(rows, listRowDefault{
			InstanceId: r.InstanceId, InstanceName: r.InstanceName,
			Framework: r.Framework, ReleaseVersion: r.ReleaseVersion,
			HadoopVersion: r.HadoopVersion, State: r.State,
			Zone: r.Zone, CreateTime: r.CreateTime, ExpireTime: r.ExpireTime,
		})
	}
	ctx.PrintList(rows)
}

func toListRow(c listClusterInfo) listRow {
	return listRow{
		InstanceId:     c.InstanceId,
		InstanceName:   c.InstanceName,
		Framework:      c.Framework,
		ReleaseVersion: c.ReleaseVersion,
		HadoopVersion:  c.HadoopVersion,
		State:          c.State,
		Zone:           c.Zone,
		VPCId:          c.VPCId,
		SubnetId:       c.SubnetId,
		ChargeType:     c.ChargeType,
		CreateTime:     formatUnixTime(c.CreateTime),
		ExpireTime:     formatUnixTime(c.ExpireTime),
	}
}

func formatUnixTime(ts int64) string {
	if ts <= 0 {
		return ""
	}
	return time.Unix(ts, 0).Format("2006-01-02")
}

func listClusterID(ctx *cli.Context, clusters []listClusterInfo) {
	ids := make([]string, 0, len(clusters))
	for _, c := range clusters {
		ids = append(ids, c.InstanceId)
	}
	fmt.Fprintln(ctx.Out(), strings.Join(ids, ","))
}

// _RetCodeRegionNoPermission is the SDK RetCode when account lacks permission in the current region.
const _RetCodeRegionNoPermission = 230
