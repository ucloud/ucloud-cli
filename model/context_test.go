package model

import (
	"fmt"
	"os"
	"testing"

	"github.com/ucloud/ucloud-sdk-go/sdk/trace"
)

var dasTraceInfo = trace.NewDasTraceInfo()

var ctx = Context{
	os.Stdout,
	&dasTraceInfo,
}

func TestPrintln(t *testing.T) {
	ctx.Println("test print")
}

func TestAppendError(t *testing.T) {
	err := fmt.Errorf("error test")
	ctx.AppendError(err)
	ctx.AppendError(err)
	errStr := dasTraceInfo.GetExtra()["error"].(string)
	expectErrStr := "error test->error test"
	if errStr != expectErrStr {
		t.Errorf("model.Context.AppendError(%v),Expect: %s, Got: %s", err, expectErrStr, errStr)
	}
}
