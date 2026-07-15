package uk8s

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

func setupCommandGateway(t *testing.T) (*cli.Context, *[]url.Values) {
	t.Helper()

	requests := &[]url.Values{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse request form: %v", err)
		}
		*requests = append(*requests, r.PostForm)
		action := r.PostForm.Get("Action")
		response := map[string]interface{}{
			"RetCode": 0,
			"Action":  action + "Response",
		}
		switch action {
		case "DescribeUK8SCluster":
			response["ClusterId"] = "uk8s-a"
			response["ClusterName"] = "test-cluster"
			response["KubeProxy"] = map[string]string{"Mode": "iptables"}
			response["MasterList"] = []map[string]interface{}{{
				"NodeId": "uhost-master", "Name": "master-0", "IPSet": []map[string]interface{}{{"IP": "10.0.0.1"}},
			}}
			response["NodeList"] = []map[string]interface{}{{
				"NodeId": "uhost-node", "Name": "node-0", "DiskSet": []map[string]interface{}{{"DiskId": "bsi-node"}},
			}}
		case "GetClusterConfig":
			response["KubeConfig"] = "apiVersion: v1\nclusters: []\n"
			response["ExternalKubeConfig"] = "apiVersion: v1\nclusters: []\n"
		case "GetUK8SVersions":
			response["Data"] = []map[string]string{{
				"K8sVersion":        "1.34.5",
				"ContainerdVersion": "1.7.27",
			}}
		case "AddUK8SNodeGroup":
			response["NodeGroupId"] = "uk8sng-test"
		case "DescribeUK8SImage":
			response["CustomImageSet"] = []map[string]interface{}{{
				"ImageId": "uimage-custom", "ImageName": "custom-ubuntu", "Features": []string{"CloudInit"},
			}}
		case "ListUK8SClusterNodeV2":
			response["NodeSet"] = []map[string]interface{}{{
				"NodeId": "uk8s-node", "UsedCPU": 87, "UsedMemory": 1052389376, "VKCpu": 0, "VKMem": 0,
			}}
		case "AddUK8SUHostNode":
			response["NodeIds"] = []string{"uk8snode-test"}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))

	t.Cleanup(func() {
		server.Close()
	})

	var out, errOut bytes.Buffer
	ctx := cli.NewContext(cli.Deps{
		In: strings.NewReader(""), Out: &out, Err: &errOut,
		Format: cli.OutputTable,
		DefaultsProvider: func() command.Defaults {
			return command.Defaults{Region: "cn-sh2", ProjectID: "org-test"}
		},
		ClientConfig: func() *sdk.Config {
			return &sdk.Config{Region: "cn-sh2", ProjectId: "org-test", BaseUrl: server.URL}
		},
		BuildCredential: func() *auth.Credential {
			return &auth.Credential{PublicKey: "public", PrivateKey: "private"}
		},
		AttachHandlers: func(sdk.ServiceClient) {},
	})
	return ctx, requests
}

func runUK8SCommand(t *testing.T, cmd *cobra.Command, args ...string) {
	t.Helper()
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute %s %v: %v", cmd.Use, args, err)
	}
}

func lastRequest(t *testing.T, requests *[]url.Values) url.Values {
	t.Helper()
	if len(*requests) == 0 {
		t.Fatal("command did not call the API")
	}
	return (*requests)[len(*requests)-1]
}

func assertRequest(t *testing.T, got url.Values, want map[string]string) {
	t.Helper()
	for key, expected := range want {
		if actual := got.Get(key); actual != expected {
			t.Errorf("request %s = %q, want %q", key, actual, expected)
		}
	}
}

func TestClusterCommandsDispatch(t *testing.T) {
	tests := []struct {
		name   string
		build  func(*cli.Context) *cobra.Command
		args   []string
		action string
		want   map[string]string
	}{
		{
			name: "delete", build: newDelete,
			args:   []string{"--cluster-id", "uk8s-a/name,uk8s-b/name", "--release-udisk", "--release-eip", "--yes"},
			action: "DelUK8SCluster", want: map[string]string{"ClusterId": "uk8s-b", "ReleaseUDisk": "true", "ReleaseEIP": "true"},
		},
		{
			name: "list", build: newList,
			args:   []string{"--cluster-id", "uk8s-a/name", "--limit", "25", "--offset", "5"},
			action: "ListUK8SClusterV2", want: map[string]string{"ClusterId": "uk8s-a", "Limit": "25", "Offset": "5"},
		},
		{
			name: "describe", build: newDescribe,
			args:   []string{"--cluster-id", "uk8s-a/name"},
			action: "DescribeUK8SCluster", want: map[string]string{"ClusterId": "uk8s-a"},
		},
		{
			name: "get-config", build: newGetConfig,
			args:   []string{"--cluster-id", "uk8s-a/name", "--external"},
			action: "GetClusterConfig", want: map[string]string{"ClusterId": "uk8s-a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, requests := setupCommandGateway(t)
			runUK8SCommand(t, tt.build(ctx), tt.args...)
			if tt.name == "delete" && len(*requests) != 2 {
				t.Fatalf("delete sent %d requests, want 2", len(*requests))
			}
			got := lastRequest(t, requests)
			tt.want["Action"] = tt.action
			assertRequest(t, got, tt.want)
		})
	}
}

func TestDescribeJSONUsesStructuredResponse(t *testing.T) {
	ctx, _ := setupCommandGateway(t)
	ctx.SetFormat(cli.OutputJSON)
	runUK8SCommand(t, newDescribe(ctx), "--cluster-id", "uk8s-a")

	var got map[string]json.RawMessage
	if err := json.Unmarshal(ctx.Out().(*bytes.Buffer).Bytes(), &got); err != nil {
		t.Fatalf("decode JSON output: %v", err)
	}
	if _, ok := got["Attribute"]; ok {
		t.Fatal("describe JSON must be a structured response object, not describe rows")
	}
	if string(got["ClusterId"]) != `"uk8s-a"` {
		t.Fatalf("ClusterId = %s, want uk8s-a", got["ClusterId"])
	}
	var kubeProxy struct{ Mode string }
	if err := json.Unmarshal(got["KubeProxy"], &kubeProxy); err != nil || kubeProxy.Mode != "iptables" {
		t.Fatalf("KubeProxy = %s, want structured object with iptables mode", got["KubeProxy"])
	}
	var masterList []struct{ NodeId string }
	if err := json.Unmarshal(got["MasterList"], &masterList); err != nil || len(masterList) != 1 || masterList[0].NodeId != "uhost-master" {
		t.Fatalf("MasterList = %s, want structured node array", got["MasterList"])
	}
}

func TestNodeGroupCommandsDispatch(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		action string
		want   map[string]string
	}{
		{
			name: "add",
			args: []string{
				"add", "--cluster-id", "uk8s-a/name", "--name", "workers", "--machine-type", "G",
				"--cpu", "4", "--memory-mb", "8192", "--gpu", "1", "--gpu-type", "V100",
				"--image-id", "uimage-a/name", "--subnet-id", "subnet-a/name", "--boot-disk-type", "CLOUD_RSSD", "--boot-disk-size-gb", "40",
				"--charge-type", "Month", "--cpu-platform", "Intel/Cascadelake",
			},
			action: "AddUK8SNodeGroup", want: map[string]string{
				"ClusterId": "uk8s-a", "NodeGroupName": "workers", "MachineType": "G", "CPU": "4", "Mem": "8192",
				"GPU": "1", "GpuType": "V100", "ImageId": "uimage-a", "SubnetId": "subnet-a",
				"ChargeType": "Month", "MinimalCpuPlatform": "Intel/Cascadelake",
			},
		},
		{
			name:   "delete",
			args:   []string{"delete", "--cluster-id", "uk8s-a/name", "--nodegroup-id", "uk8sng-a/name", "--yes"},
			action: "RemoveUK8SNodeGroup", want: map[string]string{"ClusterId": "uk8s-a", "NodeGroupId": "uk8sng-a"},
		},
		{
			name:   "list",
			args:   []string{"list", "--cluster-id", "uk8s-a/name"},
			action: "ListUK8SNodeGroup", want: map[string]string{"ClusterId": "uk8s-a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, requests := setupCommandGateway(t)
			runUK8SCommand(t, newNodeGroup(ctx), tt.args...)
			got := lastRequest(t, requests)
			tt.want["Action"] = tt.action
			assertRequest(t, got, tt.want)
		})
	}
}

func TestNodeGroupAddOmitsUnspecifiedResourceDefaults(t *testing.T) {
	ctx, requests := setupCommandGateway(t)
	runUK8SCommand(t, newNodeGroup(ctx), "add",
		"--cluster-id", "uk8s-a/name", "--name", "workers",
		"--machine-type", "N", "--cpu", "2", "--memory-mb", "4096",
		"--image-id", "uimage-a/name", "--subnet-id", "subnet-a/name", "--boot-disk-type", "CLOUD_RSSD", "--boot-disk-size-gb", "40",
		"--charge-type", "Month", "--cpu-platform", "Intel/Auto")
	got := lastRequest(t, requests)
	assertRequest(t, got, map[string]string{
		"Action": "AddUK8SNodeGroup", "ClusterId": "uk8s-a", "NodeGroupName": "workers",
		"MachineType": "N", "CPU": "2", "Mem": "4096", "ImageId": "uimage-a", "SubnetId": "subnet-a",
		"BootDiskType": "CLOUD_RSSD", "BootDiskSize": "40", "ChargeType": "Month",
		"MinimalCpuPlatform": "Intel/Auto",
	})
	for _, key := range []string{"DataDiskType", "DataDiskSize", "GPU", "GpuType"} {
		if _, ok := got[key]; ok {
			t.Errorf("unspecified node-group field %s must be omitted", key)
		}
	}
}

func TestNodeCommandsDispatch(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		action string
		want   map[string]string
	}{
		{
			name: "add",
			args: []string{
				"add", "--cluster-id", "uk8s-a/name", "--cpu", "2", "--memory-mb", "4096", "--count", "2",
				"--charge-type", "Dynamic", "--password", "Password1", "--image-id", "uimage-a/name",
				"--isolation-group-id", "ig-a/name", "--group", "Default", "--user-data", "cloud-init", "--init-script", "echo ready",
			},
			action: "AddUK8SUHostNode", want: map[string]string{
				"ClusterId": "uk8s-a", "CPU": "2", "Mem": "4096", "Count": "2", "ChargeType": "Dynamic",
				"Password": base64.StdEncoding.EncodeToString([]byte("Password1")), "ImageId": "uimage-a",
				"IsolationGroup": "ig-a", "Tag": "Default",
				"UserData":   base64.StdEncoding.EncodeToString([]byte("cloud-init")),
				"InitScript": base64.StdEncoding.EncodeToString([]byte("echo ready")),
			},
		},
		{
			name:   "delete",
			args:   []string{"delete", "--cluster-id", "uk8s-a/name", "--node-id", "node-a/name,node-b/name", "--release-data-udisk=false", "--yes"},
			action: "DelUK8SClusterNodeV2", want: map[string]string{"ClusterId": "uk8s-a", "NodeId": "node-b", "ReleaseDataUDisk": "false"},
		},
		{
			name:   "list",
			args:   []string{"list", "--cluster-id", "uk8s-a/name"},
			action: "ListUK8SClusterNodeV2", want: map[string]string{"ClusterId": "uk8s-a"},
		},
		{
			name:   "describe",
			args:   []string{"describe", "--cluster-id", "uk8s-a/name", "--node-id", "10.0.0.8"},
			action: "DescribeUK8SNode", want: map[string]string{"ClusterId": "uk8s-a", "Name": "10.0.0.8"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, requests := setupCommandGateway(t)
			runUK8SCommand(t, newNode(ctx), tt.args...)
			if tt.name == "delete" && len(*requests) != 2 {
				t.Fatalf("delete sent %d requests, want 2", len(*requests))
			}
			got := lastRequest(t, requests)
			tt.want["Action"] = tt.action
			assertRequest(t, got, tt.want)
		})
	}
}

func TestImageListDispatch(t *testing.T) {
	ctx, requests := setupCommandGateway(t)
	runUK8SCommand(t, newImage(ctx), "list", "--zone", "cn-sh2-01")
	assertRequest(t, lastRequest(t, requests), map[string]string{
		"Action": "DescribeUK8SImage", "Region": "cn-sh2", "ProjectId": "org-test", "Zone": "cn-sh2-01",
	})
}

func TestCompatibilityResponsesKeepUpdatedUK8SFields(t *testing.T) {
	tests := []struct {
		name  string
		build func(*cli.Context) *cobra.Command
		args  []string
		field string
	}{
		{
			name: "cluster describe", build: newDescribe,
			args: []string{"--cluster-id", "uk8s-a"}, field: "ClusterId",
		},
		{
			name: "image list", build: func(ctx *cli.Context) *cobra.Command { return newImage(ctx) },
			args: []string{"list"}, field: "CustomImageSet",
		},
		{
			name: "node list", build: func(ctx *cli.Context) *cobra.Command { return newNode(ctx) },
			args: []string{"list", "--cluster-id", "uk8s-a"}, field: "NodeSet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _ := setupCommandGateway(t)
			ctx.SetFormat(cli.OutputJSON)
			runUK8SCommand(t, tt.build(ctx), tt.args...)

			var output map[string]json.RawMessage
			if err := json.Unmarshal(ctx.Out().(*bytes.Buffer).Bytes(), &output); err != nil {
				t.Fatalf("decode JSON output: %v", err)
			}
			if _, ok := output[tt.field]; !ok {
				t.Fatalf("output does not contain %q: %s", tt.field, ctx.Out().(*bytes.Buffer).String())
			}
		})
	}
}

func TestVersionListDispatch(t *testing.T) {
	ctx, requests := setupCommandGateway(t)
	runUK8SCommand(t, newVersion(ctx), "list")
	assertRequest(t, lastRequest(t, requests), map[string]string{
		"Action": "GetUK8SVersions", "Region": "cn-sh2", "ProjectId": "org-test", "Kind": defaultUK8SKind,
	})
}

func TestMutationValidationStopsBeforeAPI(t *testing.T) {
	tests := []struct {
		name string
		cmd  func(*cli.Context) *cobra.Command
		args []string
		want string
	}{
		{
			name: "node count",
			cmd:  func(ctx *cli.Context) *cobra.Command { return newNode(ctx) },
			args: []string{"add", "--cluster-id", "uk8s-a", "--cpu", "2", "--memory-mb", "4096", "--count", "0", "--charge-type", "Dynamic", "--password", "Password1"},
			want: "--count must be between 1 and 50",
		},
		{
			name: "nodegroup gpu",
			cmd:  func(ctx *cli.Context) *cobra.Command { return newNodeGroup(ctx) },
			args: []string{"add", "--cluster-id", "uk8s-a", "--name", "gpu", "--machine-type", "G", "--cpu", "2", "--memory-mb", "4096", "--image-id", "uimage-a", "--subnet-id", "subnet-a", "--boot-disk-type", "CLOUD_RSSD", "--boot-disk-size-gb", "40"},
			want: "--gpu and --gpu-type are required",
		},
		{
			name: "nodegroup boot disk type",
			cmd:  func(ctx *cli.Context) *cobra.Command { return newNodeGroup(ctx) },
			args: []string{"add", "--cluster-id", "uk8s-a", "--name", "workers", "--machine-type", "N", "--cpu", "2", "--memory-mb", "4096", "--image-id", "uimage-a", "--subnet-id", "subnet-a", "--boot-disk-size-gb", "40"},
			want: "--boot-disk-type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, requests := setupCommandGateway(t)
			cmd := tt.cmd(ctx)
			cmd.SetArgs(tt.args)
			err := cmd.Execute()
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want containing %q", err, tt.want)
			}
			if len(*requests) != 0 {
				t.Fatal("invalid mutation must not reach the API")
			}
		})
	}
}
