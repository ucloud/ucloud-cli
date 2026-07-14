package uhadoop

import (
	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"

	uhadoopsdk "github.com/ucloud/ucloud-sdk-go/services/uhadoop"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

type describeClusterResponse struct {
	response.CommonBase
	ClusterSet []describeClusterInfo `json:"ClusterSet"`
}

type describeClusterInfo struct {
	InstanceId          string        `json:"InstanceId"`
	ClusterInstanceId   string        `json:"ClusterInstanceId"`
	InstanceName        string        `json:"InstanceName"`
	ClusterInstanceName string        `json:"ClusterInstanceName"`
	FlinkResourceId     string        `json:"FlinkResourceId"`
	Framework           string        `json:"Framework"`
	FrameworkVersion    string        `json:"FrameworkVersion"`
	ReleaseVersion      string        `json:"ReleaseVersion"`
	HadoopVersion       string        `json:"HadoopVersion"`
	State               string        `json:"State"`
	Zone                string        `json:"Zone"`
	VPCId               string        `json:"VPCId"`
	SubnetId            string        `json:"SubnetId"`
	BusinessId          string        `json:"BusinessId"`
	ChargeType          string        `json:"ChargeType"`
	Tag                 string        `json:"Tag"`
	CreateTime          int64         `json:"CreateTime"`
	ExpireTime          int64         `json:"ExpireTime"`
	RunningTime         int64         `json:"RunningTime"`
	MasterCount         int           `json:"MasterCount"`
	CoreCount           int           `json:"CoreCount"`
	TaskCount           int           `json:"TaskCount"`
	NodeCount           int           `json:"NodeCount"`
	RedundantCount      int           `json:"RedundantCount"`
	AppConfigCount      int           `json:"AppConfigCount"`
	IsOpenSecGroup      bool          `json:"IsOpenSecGroup"`
	HdfsTotal           int           `json:"HdfsTotal"`
	HdfsUsed            int           `json:"HdfsUsed"`
	NodeSet             []interface{} `json:"NodeSet"`
	AppConfigSet        []interface{} `json:"AppConfigSet"`
}

func newDescribe(ctx *cli.Context) *cobra.Command {
	client := cli.NewServiceClient(ctx, uhadoopsdk.NewClient)
	req := client.NewDescribeUHadoopInstanceRequest()
	cmd := &cobra.Command{
		Use:          "describe <instance-id>",
		Short:        "Describe a UHadoop cluster",
		Long:         `Describe a UHadoop cluster with detailed information`,
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			req.InstanceId = sdk.String(args[0])
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
	cmd.MarkFlagRequired("region")
	cmd.MarkFlagRequired("zone")
	return cmd
}
