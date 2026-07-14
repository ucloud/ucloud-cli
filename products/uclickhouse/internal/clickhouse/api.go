package clickhouse

import (
	"fmt"
	"strings"

	uclickhousesdk "github.com/ucloud/ucloud-sdk-go/services/uclickhouse"
	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"
)

type opResponse struct {
	response.CommonBase
}

type createUClickhouseClusterResponse struct {
	response.CommonBase
	Data createUClickhouseClusterResponseData
}

type createUClickhouseClusterResponseData struct {
	ClusterId string
}

func invokeUClickhouseAction(client *uclickhousesdk.UClickhouseClient, action string, req request.Common, resp response.Common) error {
	return enrichUClickhouseError(action, client.Client.InvokeAction(action, req, resp))
}

func enrichUClickhouseError(action string, err error) error {
	if err == nil {
		return nil
	}
	uErr, ok := err.(uerr.Error)
	if !ok || uErr.Code() == 0 {
		return err
	}
	message := strings.TrimSpace(uErr.Message())
	if message == "" {
		message = "<empty from service>"
	}
	detail := fmt.Sprintf("UClickhouse API %s failed. RetCode:%d. Message:%s", action, uErr.Code(), message)
	if strings.TrimSpace(uErr.Message()) == "" {
		detail += "\nThe service did not return an error message. Check region/project and create-option compatibility, for example: ucloud uclickhouse create-option --region <region>. Rerun with --debug if request details are needed."
	}
	return fmt.Errorf("%s", detail)
}
