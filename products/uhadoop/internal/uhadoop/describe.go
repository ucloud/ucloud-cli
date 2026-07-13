package uhadoop

import (
	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/ucloud/response"

	uhadoopsdk "github.com/ucloud/ucloud-sdk-go/services/uhadoop"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// describeClusterResponse mirrors the real API response for DescribeUHadoopInstance.
// The SDK's ClusterInfo has CreateTime/ExpireTime as string but the API returns
// Unix timestamps (int), and many fields are missing or wrongly typed.
type describeClusterResponse struct {
	response.CommonBase
	ClusterSet []describeClusterInfo `json:"ClusterSet"`
}

type describeClusterInfo struct {
	InstanceId         string `json:"InstanceId"`
	ClusterInstanceId  string `json:"ClusterInstanceId"`
	InstanceName       string `json:"InstanceName"`
	ClusterInstanceName string `json:"ClusterInstanceName"`
	FlinkResourceId    string `json:"FlinkResourceId"`
	Framework           string `json:"Framework"`
	FrameworkVersion   string `json:"FrameworkVersion"`
	ReleaseVersion     string `json:"ReleaseVersion"`
	HadoopVersion      string `json:"HadoopVersion"`
	State              string `json:"State"`
	Zone               string `json:"Zone"`
	VPCId              string `json:"VPCId"`
	SubnetId           string `json:"SubnetId"`
	BusinessId         string `json:"BusinessId"`
	ChargeType         string `json:"ChargeType"`
	Tag                string `json:"Tag"`
	CreateTime         int64  `json:"CreateTime"`
	ExpireTime         int64  `json:"ExpireTime"`
	RunningTime        int64  `json:"RunningTime"`
	MasterCount        int    `json:"MasterCount"`
	CoreCount          int    `json:"CoreCount"`
	TaskCount          int    `json:"TaskCount"`
	NodeCount          int    `json:"NodeCount"`
	RedundantCount     int    `json:"RedundantCount"`
	AppConfigCount     int    `json:"AppConfigCount"`
	IsOpenSecGroup     bool   `json:"IsOpenSecGroup"`
	HdfsTotal          int    `json:"HdfsTotal"`
	HdfsUsed           int    `json:"HdfsUsed"`
	// Complex nested fields (NodeSet/AppConfigSet) kept as interface{} to avoid
	// deep struct definitions; omitted in table mode, full in JSON.
	NodeSet      []interface{} `json:"NodeSet"`
	AppConfigSet []interface{} `json:"AppConfigSet"`
}

// newDescribe ucloud uhadoop describe
func newDescribe(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uhadoopsdk.NewClient)
	req := client.NewDescribeUHadoopInstanceRequest()
	cmd := &cobra.Command{
		Use:   "describe <instance-id>",
		Short: "Describe a UHadoop cluster",
		Long:  `Describe a UHadoop cluster with detailed information`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			req.InstanceId = sdkStr(args[0])
			var resp describeClusterResponse
			err := client.InvokeAction("DescribeUHadoopInstance", req, &resp)
			if err != nil {
				ctx.HandleError(err)
				return
			}
			ctx.PrintList(resp.ClusterSet)
		},
	}
	cmd.Flags().SortFlags = false
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)

	command.SetFlagValues(cmd, "region", ctx.RegionList()...)

	return cmd
}

func sdkStr(s string) *string { return &s }
