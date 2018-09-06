package model

import (
	"os"
	"testing"

	"github.com/ucloud/ucloud-sdk-go/sdk"
)

var context_test = Context{
	os.Stdout,
	&sdk.ClientConfig{},
}

func TestPrintln(t *testing.T) {
	context_test.Println("test print")
}
