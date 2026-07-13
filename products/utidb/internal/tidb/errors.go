package tidb

import (
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// retCodeHints maps TiDB API RetCode to user-facing hints when Message is empty
// or too terse. Keep product-specific codes here rather than global CLI error handling.
var retCodeHints = map[int]string{
	202555: "backup databases is empty (备份数据库为空库)",
}

func retCodeHint(code int) string {
	return retCodeHints[code]
}

// enrichAPIError returns an error with a clearer Message for known TiDB RetCodes.
func enrichAPIError(err error) error {
	uErr, ok := err.(uerr.Error)
	if !ok || uErr.Code() == 0 {
		return err
	}
	hint := retCodeHint(uErr.Code())
	if hint == "" {
		return err
	}
	msg := uErr.Message()
	if msg == "" {
		msg = hint
	} else {
		msg = msg + "; " + hint
	}
	return uerr.NewServerCodeError(uErr.Code(), msg)
}

func handleAPIError(ctx *cli.Context, err error) {
	ctx.HandleError(enrichAPIError(err))
}
