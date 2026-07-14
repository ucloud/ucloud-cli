package tidb

import (
	"strings"
	"testing"

	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

func TestEnrichAPIError_202555(t *testing.T) {
	err := uerr.NewServerCodeError(202555, "")
	got := enrichAPIError(err)
	uErr, ok := got.(uerr.Error)
	if !ok {
		t.Fatalf("want uerr.Error, got %T", got)
	}
	if uErr.Code() != 202555 {
		t.Fatalf("code = %d, want 202555", uErr.Code())
	}
	if !strings.Contains(uErr.Message(), "backup databases is empty") {
		t.Fatalf("message = %q, want empty-db hint", uErr.Message())
	}
	if !strings.Contains(uErr.Message(), "备份数据库为空库") {
		t.Fatalf("message = %q, want Chinese hint", uErr.Message())
	}
}

func TestEnrichAPIError_unknownCode(t *testing.T) {
	err := uerr.NewServerCodeError(999999, "original")
	got := enrichAPIError(err)
	if got != err {
		t.Fatal("unknown code should be unchanged")
	}
}
