// hack/t0/main.go
// T0 门禁验证程序：空 AK/SK + Authorization: Bearer 头，跨产品线抽样。
// 读覆盖 uhost/unet/udb，写覆盖 uhost/unet（udb 写涉及计费实例，用户裁定豁免）。
// 用法：
//   export UCLOUD_T0_TOKEN=...   # OAuth access_token
//   export UCLOUD_T0_REGION=cn-bj2
//   export UCLOUD_T0_ZONE=cn-bj2-04
//   export UCLOUD_T0_PROJECT=org-xxxxx
//   go run -mod=vendor ./hack/t0
package main

import (
	"fmt"
	"os"
	"time"

	uhttp "github.com/ucloud/ucloud-sdk-go/private/protocol/http"
	"github.com/ucloud/ucloud-sdk-go/services/uaccount"
	"github.com/ucloud/ucloud-sdk-go/services/udb"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

func main() {
	token := os.Getenv("UCLOUD_T0_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "UCLOUD_T0_TOKEN is required")
		os.Exit(1)
	}
	cfg := &sdk.Config{
		BaseUrl:   "https://api.ucloud.cn/",
		Timeout:   15 * time.Second,
		Region:    os.Getenv("UCLOUD_T0_REGION"),
		Zone:      os.Getenv("UCLOUD_T0_ZONE"),
		ProjectId: os.Getenv("UCLOUD_T0_PROJECT"),
	}
	cred := &auth.Credential{} // 空 AK/SK：PublicKey 为空时 SDK 不注入签名参数
	bearer := func(c *sdk.Client, req *uhttp.HttpRequest) (*uhttp.HttpRequest, error) {
		return req, req.SetHeader("Authorization", "Bearer "+token)
	}

	failed := false
	check := func(name string, err error) {
		if err != nil {
			failed = true
			fmt.Printf("[FAIL] %s: %v\n", name, err)
		} else {
			fmt.Printf("[OK]   %s\n", name)
		}
	}

	// 读：uaccount GetRegion
	uaClient := uaccount.NewClient(cfg, cred)
	uaClient.Client.AddHttpRequestHandler(bearer)
	_, err := uaClient.GetRegion(uaClient.NewGetRegionRequest())
	check("read  uaccount.GetRegion", err)

	// 读：uhost DescribeUHostInstance
	uhClient := uhost.NewClient(cfg, cred)
	uhClient.Client.AddHttpRequestHandler(bearer)
	_, err = uhClient.DescribeUHostInstance(uhClient.NewDescribeUHostInstanceRequest())
	check("read  uhost.DescribeUHostInstance", err)

	// 读：unet DescribeEIP
	unClient := unet.NewClient(cfg, cred)
	unClient.Client.AddHttpRequestHandler(bearer)
	_, err = unClient.DescribeEIP(unClient.NewDescribeEIPRequest())
	check("read  unet.DescribeEIP", err)

	// 读：udb DescribeUDBInstance
	udbClient := udb.NewClient(cfg, cred)
	udbClient.Client.AddHttpRequestHandler(bearer)
	udbReq := udbClient.NewDescribeUDBInstanceRequest()
	udbReq.ClassType = sdk.String("sql")
	udbReq.Offset = sdk.Int(0)
	udbReq.Limit = sdk.Int(10)
	_, err = udbClient.DescribeUDBInstance(udbReq)
	check("read  udb.DescribeUDBInstance", err)

	// 写：uhost CreateIsolationGroup + DeleteIsolationGroup（免费、无副作用残留）
	cigReq := uhClient.NewCreateIsolationGroupRequest()
	cigReq.GroupName = sdk.String(fmt.Sprintf("t0-gate-probe-%d", time.Now().Unix()))
	cigResp, err := uhClient.CreateIsolationGroup(cigReq)
	check("write uhost.CreateIsolationGroup", err)
	if err == nil {
		digReq := uhClient.NewDeleteIsolationGroupRequest()
		digReq.GroupId = sdk.String(cigResp.GroupId)
		_, err = uhClient.DeleteIsolationGroup(digReq)
		check("write uhost.DeleteIsolationGroup (cleanup)", err)
	}

	// 写：unet AllocateEIP + ReleaseEIP（按量最小档，立即释放）
	aeReq := unClient.NewAllocateEIPRequest()
	aeReq.OperatorName = sdk.String("Bgp")
	aeReq.Bandwidth = sdk.Int(1)
	aeReq.PayMode = sdk.String("Traffic")
	aeReq.ChargeType = sdk.String("Dynamic")
	aeResp, err := unClient.AllocateEIP(aeReq)
	check("write unet.AllocateEIP", err)
	if err == nil && len(aeResp.EIPSet) == 0 {
		check("write unet.AllocateEIP (empty EIPSet — cannot cleanup)", fmt.Errorf("EIPSet is empty"))
	}
	if err == nil && len(aeResp.EIPSet) > 0 {
		reReq := unClient.NewReleaseEIPRequest()
		reReq.EIPId = sdk.String(aeResp.EIPSet[0].EIPId)
		_, err = unClient.ReleaseEIP(reReq)
		check("write unet.ReleaseEIP (cleanup)", err)
	}

	if failed {
		fmt.Println("\nT0 GATE: FAILED — do NOT start implementation; review credential model in spec")
		os.Exit(1)
	}
	fmt.Println("\nT0 GATE: PASSED")
}
