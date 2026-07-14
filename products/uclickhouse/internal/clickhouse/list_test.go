package clickhouse

import (
	"encoding/json"
	"strings"
	"testing"

	uclickhousesdk "github.com/ucloud/ucloud-sdk-go/services/uclickhouse"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

func TestListUClickhouseClusterResponseDecodesClusterArray(t *testing.T) {
	body := []byte(`{
		"RetCode": 0,
		"Message": "success",
		"Data": {
			"Clusters": [
				{
					"ClusterId": "uck-1",
					"ClusterName": "test",
					"Status": "RUNNING",
					"CreateTimestamp": 1783700833334
				},
				{
					"ClusterId": "uck-2",
					"ClusterName": "tp_test",
					"Status": "RUNNING",
					"CreateTimestamp": 1783504224814
				}
			],
			"TotalCount": 2
		}
	}`)

	var resp listUClickhouseClusterResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("unmarshal list response: %v", err)
	}
	if got := len(resp.Data.Clusters); got != 2 {
		t.Fatalf("cluster count = %d, want 2", got)
	}
	if got := resp.Data.Clusters[1].ClusterId; got != "uck-2" {
		t.Fatalf("second cluster id = %q, want uck-2", got)
	}
}

func TestDescribeUClickhouseClusterResponseDecodesStringPaymentPrice(t *testing.T) {
	body := []byte(`{
		"RetCode": 0,
		"Message": "success",
		"Data": {
			"Cluster": {
				"ClusterId": "uck-1",
				"ClusterName": "test",
				"Status": "RUNNING",
				"CreateTimestamp": 1783700833334,
				"ExpireTimestamp": null
			},
			"ClickhouseNodes": [],
			"Payment": {
				"ChargeType": "Dynamic",
				"CreateTimestamp": 1783700878,
				"ExpireTimestamp": 1783926000,
				"Price": "1.63",
				"OriginalPrice": "1.63",
				"ResourceId": "uck-1"
			},
			"ZookeeperNodes": []
		}
	}`)

	var resp describeUClickhouseClusterResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("unmarshal describe response: %v", err)
	}
	if got := resp.Data.Payment.Price.String(); got != "1.63" {
		t.Fatalf("payment price = %q, want 1.63", got)
	}
}

func TestCreateAndOpResponsesDecodeMinimalPayloads(t *testing.T) {
	var createResp createUClickhouseClusterResponse
	if err := json.Unmarshal([]byte(`{"RetCode":0,"Message":"success","Data":{"ClusterId":"uck-new"}}`), &createResp); err != nil {
		t.Fatalf("unmarshal create response: %v", err)
	}
	if got := createResp.Data.ClusterId; got != "uck-new" {
		t.Fatalf("cluster id = %q, want uck-new", got)
	}

	var opResp opResponse
	if err := json.Unmarshal([]byte(`{"RetCode":0,"Message":"success"}`), &opResp); err != nil {
		t.Fatalf("unmarshal op response: %v", err)
	}
	if got := opResp.Message; got != "success" {
		t.Fatalf("message = %q, want success", got)
	}
}

func TestClusterRowFormatsMillisecondCreateTimestamp(t *testing.T) {
	row := clusterRow(uclickhousesdk.ClickhouseCluster{
		ClusterId:       "uck-1",
		CreateTimestamp: 1783700833334,
		ExpireTimestamp: 1783926000,
	})

	if row.CreateTime != "2026-07-11" {
		t.Fatalf("CreateTime = %q, want 2026-07-11", row.CreateTime)
	}
	if row.ExpireTime != "2026-07-13" {
		t.Fatalf("ExpireTime = %q, want 2026-07-13", row.ExpireTime)
	}
}

func TestFormatUnixDateReturnsEmptyForMissingTimestamp(t *testing.T) {
	if got := formatUnixDate(0); got != "" {
		t.Fatalf("formatUnixDate(0) = %q, want empty", got)
	}
}

func TestEnrichUClickhouseErrorFormatsEmptyServerMessage(t *testing.T) {
	err := enrichUClickhouseError("CreateUClickhouseCluster", uerr.NewServerCodeError(207803, ""))
	if err == nil {
		t.Fatal("expected enriched error")
	}
	got := err.Error()
	for _, want := range []string{
		"UClickhouse API CreateUClickhouseCluster failed",
		"RetCode:207803",
		"Message:<empty from service>",
		"ucloud uclickhouse create-option",
		"--debug",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("error = %q, want contain %q", got, want)
		}
	}
}

func TestEnrichUClickhouseErrorFormatsServerMessage(t *testing.T) {
	err := enrichUClickhouseError("ListUClickhouseCluster", uerr.NewServerCodeError(123, "boom"))
	if err == nil {
		t.Fatal("expected enriched error")
	}
	got := err.Error()
	if !strings.Contains(got, "UClickhouse API ListUClickhouseCluster failed") || !strings.Contains(got, "Message:boom") {
		t.Fatalf("error = %q, want action and service message", got)
	}
}
