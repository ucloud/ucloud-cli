package trace

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"

	"github.com/satori/go.uuid"

	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

// DefaultDasURI is the default das api endpoint at internet.
const DefaultDasURI = "https://das-rpt.ucloud.cn/log"

// DasTracer is a reporter to send traceback to remote trace system (named dasman)
type DasTracer struct {
	URI string
}

// NewDasTracer will create a new das tracer instance to send tracing report
func NewDasTracer() DasTracer {
	return DasTracer{
		URI: DefaultDasURI,
	}
}

// Send will send trace information via a http(s)/tcp connection
func (d *DasTracer) Send(t TraceInfo, header map[string]string) error {
	body, err := marshalTraceInfo(t)
	if err != nil {
		return err
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", d.URI, bytes.NewReader(body))
	for key, value := range header {
		req.Header.Add(key, value)
	}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("invalid http status with tracer, %s", resp.Status)
	}
	return nil
}

func marshalTraceInfo(t TraceInfo) ([]byte, error) {
	// TODO: shouldn't use map, use new version sdk
	query := t.GetSDKRequest().(map[string]string)
	resp := t.GetSDKResponse().(response.Common)
	extra := t.GetExtra()

	dataSet := make([]map[string]interface{}, 0)
	dataItem := map[string]interface{}{
		"level":   "info", //todo
		"topic":   "api",
		"action":  resp.GetAction(),
		"command": extra["command"],
		"error":   extra["error"],
		"req":     query,
		"res": map[string]interface{}{
			"Action":  resp.GetAction(),
			"RetCode": resp.GetRetCode(),
			"Message": resp.GetMessage(),
		},
		"st": extra["startTime"],
		"rt": extra["endTime"],
		"dt": extra["durationTime"],
	}

	dataSet = append(dataSet, dataItem)
	reqUUID := uuid.NewV4()
	sessionID := uuid.NewV4()
	payload := map[string]interface{}{
		"aid":  "iywtleaa",
		"uuid": reqUUID,
		"sid":  sessionID,
		"cs": map[string]interface{}{
			"uname": extra["userName"],
			// "cname": extra["companyName"],
		},
		"ds": dataSet,
	}
	if action, ok := query["action"]; ok {
		payload["action"] = action
	}

	logrus.Infof("payload: %#v", payload)
	marshaled, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.Errorf("cannot to marshal traceinfo, %s", err)
	}
	for i := 0; i < len(marshaled); i++ {
		marshaled[i] = ^marshaled[i]
	}
	return marshaled, nil
}
