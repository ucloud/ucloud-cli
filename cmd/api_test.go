package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ucloud/ucloud-cli/cmd/internal/platform"
)

func testIntPtr(i int) *int { return &i }

func withTestRuntime(t *testing.T, baseURL string) {
	t.Helper()
	t.Setenv("COMP_LINE", "1")

	oldRuntime, oldAutoStub := activeRuntime, runtimeAutoStub
	oldConfig, oldClientConfig, oldCredential := platform.ConfigIns, platform.ClientConfig, platform.AuthCredential
	t.Cleanup(func() {
		activeRuntime, runtimeAutoStub = oldRuntime, oldAutoStub
		platform.ConfigIns, platform.ClientConfig, platform.AuthCredential = oldConfig, oldClientConfig, oldCredential
	})

	ac := &platform.AggConfig{
		Profile:       "test",
		Active:        true,
		ProjectID:     "org-test",
		Region:        "cn-bj2",
		Zone:          "cn-bj2-03",
		BaseURL:       baseURL,
		Timeout:       platform.DefaultTimeoutSec,
		MaxRetryTimes: testIntPtr(0),
		PublicKey:     "pub",
		PrivateKey:    "pri",
	}
	sdkConfig, credConfig, err := platform.BuildClientRuntime(ac)
	if err != nil {
		t.Fatalf("BuildClientRuntime returned error: %v", err)
	}
	activeRuntime = &runtimeState{
		Config:     ac,
		SDKConfig:  sdkConfig,
		Credential: credConfig,
	}
	runtimeAutoStub = false
}

func TestGenericInvokeRepeatWrapperReturnsErrorOnCreateFailure(t *testing.T) {
	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"Action":"CreateULHostInstanceResponse","RetCode":8010,"Message":"boom"}`)
	}))
	t.Cleanup(gateway.Close)
	withTestRuntime(t, gateway.URL)

	var out bytes.Buffer
	err := genericInvokeRepeatWrapper(&RepeatsConfig{IDInResp: "ULHostId"}, map[string]interface{}{
		ActionField: "CreateULHostInstance",
		"Region":    "cn-bj2",
	}, "CreateULHostInstance", 1, 1, &out)
	if err == nil {
		t.Fatalf("expected repeat wrapper to return an error on create failure, output: %s", out.String())
	}
	if !strings.Contains(out.String(), "boom") {
		t.Fatalf("expected output to include API error detail, got: %s", out.String())
	}
	if !strings.Contains(out.String(), "finally, total:1, success:0, fail:1") {
		t.Fatalf("expected final summary to report one failure, got: %s", out.String())
	}
}

func TestNewCmdAPIReturnsErrorOnRepeatFailure(t *testing.T) {
	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"Action":"CreateULHostInstanceResponse","RetCode":8010,"Message":"boom"}`)
	}))
	t.Cleanup(gateway.Close)
	withTestRuntime(t, gateway.URL)

	var out bytes.Buffer
	cmd := NewCmdAPI(&out)
	if cmd.RunE == nil {
		t.Fatal("api command must expose RunE so direct-run callers can preserve a non-zero exit code")
	}
	err := cmd.RunE(cmd, []string{
		"--Action", "CreateULHostInstance",
		"--Region", "cn-bj2",
		"--repeats", "1",
		"--concurrent", "1",
	})
	if err == nil {
		t.Fatalf("expected api RunE to return repeat failure, output: %s", out.String())
	}
	if !strings.Contains(out.String(), "boom") {
		t.Fatalf("expected output to include API error detail, got: %s", out.String())
	}
}
