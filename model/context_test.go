package model

import (
	"fmt"
	"os"
	"testing"
)

var ctx = Context{
	writer: os.Stdout,
}

func TestPrintln(t *testing.T) {
	ctx.Println("test print")
}

func TestPrintErr(t *testing.T) {
	err := fmt.Errorf("error test")
	ctx.PrintErr(err)
}
