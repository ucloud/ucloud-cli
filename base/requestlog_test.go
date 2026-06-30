package base

import (
	"strings"
	"testing"

	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
)

func TestRequestLogLine(t *testing.T) {
	req := &request.CommonBase{}
	req.SetAction("DescribeUHostInstance")
	req.SetRegion("cn-bj2")

	line := requestLogLine(req)

	if !strings.HasPrefix(line, "api: DescribeUHostInstance") {
		t.Fatalf("want 'api: DescribeUHostInstance...' prefix, got %q", line)
	}
	if !strings.Contains(line, "request:") {
		t.Fatalf("want 'request:' marker in %q", line)
	}
	if !strings.Contains(line, "cn-bj2") {
		t.Fatalf("want region in serialized request, got %q", line)
	}
}
