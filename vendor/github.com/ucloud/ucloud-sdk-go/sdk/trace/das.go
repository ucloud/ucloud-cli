package trace

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

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

//AggDasTracer 聚合了DasTracer和DasTraceInfo, 把发送和数据整合到一起，方便使用
type AggDasTracer struct {
	DasTracer
	DasTraceInfo
	HTTPHeaders map[string]string
	AllowTrace  bool
	Output      io.Writer
}

//NewAggDasTracer 创建新的AggDasTracer
func NewAggDasTracer() *AggDasTracer {
	tracer := &AggDasTracer{
		DasTracer:    NewDasTracer(),
		DasTraceInfo: NewDasTraceInfo(),
		HTTPHeaders:  make(map[string]string),
		Output:       os.Stdout,
	}
	//上报服务对Origin请求头有限制，必须以'.ucloud.cn'结尾，因此这里伪造了一个sdk.ucloud.cn,跟其他上报区分
	tracer.HTTPHeaders["Origin"] = "https://sdk.ucloud.cn"
	return tracer
}

//Send 发送上报数据
func (t *AggDasTracer) Send() error {
	if t.AllowTrace == false {
		return nil
	}
	body, err := marshalTraceInfo(&t.DasTraceInfo)
	if err != nil {
		return err
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", t.URI, bytes.NewReader(body))
	for key, value := range t.HTTPHeaders {
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
	t.Clear()
	return nil
}

// Println 在当前执行环境打印一行
func (t *AggDasTracer) Println(a ...interface{}) (n int, err error) {
	text := fmt.Sprintln(a...)
	return fmt.Fprint(t.Output, text)
}

// Printf 在当前执行环境打印一串
func (t *AggDasTracer) Printf(str string, a ...interface{}) (n int, err error) {
	text := fmt.Sprintf(str, a...)
	return fmt.Fprint(t.Output, text)
}

// Print 在当前执行环境打印一串
func (t *AggDasTracer) Print(a ...interface{}) (n int, err error) {
	text := fmt.Sprint(a...)
	return fmt.Fprint(t.Output, text)
}

// AppendError 添加上报的错误
func (t *AggDasTracer) AppendError(err error) {
	t.AppendInfo("error", err.Error())
}

// AppendInfo 添加上报的内容（key,value)
func (t *AggDasTracer) AppendInfo(key, value string) {
	tracerData := t.DasTraceInfo.GetExtra()
	oldVal, ok := tracerData[key].(string)
	if ok {
		tracerData[key] = oldVal + "->" + value
	} else {
		tracerData[key] = value
	}
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
