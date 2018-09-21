package sdk

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"github.com/ucloud/ucloud-sdk-go/sdk/version"

	"github.com/ucloud/ucloud-sdk-go/sdk/response"

	"github.com/Sirupsen/logrus"
	"github.com/parnurzeal/gorequest"

	"github.com/ucloud/ucloud-sdk-go/sdk/auth"
	uerr "github.com/ucloud/ucloud-sdk-go/sdk/error"
	"github.com/ucloud/ucloud-sdk-go/sdk/log"
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	utrace "github.com/ucloud/ucloud-sdk-go/sdk/trace"
	"github.com/ucloud/ucloud-sdk-go/sdk/utils"
)

//Client 客户端
type Client struct {
	credential *auth.Credential
	config     *ClientConfig
	Tracer     *utrace.AggDasTracer
}

//WithTracer 给Client添加Tracer
func (c *Client) WithTracer(t *utrace.AggDasTracer) *Client {
	c.Tracer = t
	return c
}

// NewClient will create an client of ucloud sdk
func NewClient(config *ClientConfig, credential *auth.Credential) *Client {
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	if config.Logger == nil {
		log.Init(config.LogLevel)
		// config.Logger = logrus.WithField("client", "sdk")
		config.Logger = logrus.New()
		config.Logger.SetLevel(config.LogLevel)
	}

	// if config.Tracer == nil {
	// 	tracer := utrace.NewDasTracer()
	// 	config.Tracer = &tracer
	// }

	// if config.TracerData == nil {
	// 	config.TracerData = make(map[string]interface{})
	// }

	// if config.HTTPHeaders == nil {
	// 	config.HTTPHeaders = make(map[string]string)
	// }

	tracer := utrace.NewAggDasTracer()

	tracer.HTTPHeaders["User-Agent"] = fmt.Sprintf("GO/%s GO-SDK/%s %s", runtime.Version(), version.Version, config.UserAgent)

	// tracer.HTTPHeaders = config.HTTPHeaders

	return &Client{
		credential: credential,
		config:     config,
		Tracer:     tracer,
	}
}

// GetCredential will return the creadential config of client.
func (c *Client) GetCredential() *auth.Credential {
	return c.credential
}

// GetConfig will return the config of client.
func (c *Client) GetConfig() *ClientConfig {
	return c.config
}

// DoRequest will send a real http request to api endpoint with retry.
func (c *Client) DoRequest(req *request.HttpRequest, resp response.Common) error {
	// config := c.GetConfig()
	r, err := c.buildSuperAgent(req)
	if err != nil {
		return err
	}

	tracer := c.Tracer

	tracer.SetSDKRequest(req.Query)

	// temporary method, should use new version sdk
	sendWithTracer := func(sendType string) (*http.Response, error) {
		startTime := time.Now()
		innerHttpResp, body, err := c.send(r, sendType)
		endTime := time.Now()

		// Unmarshal response
		err = json.Unmarshal(body, &resp)
		if err != nil {
			return nil, err
		}

		tracer.SetSDKResponse(resp)
		tracer.SetExtraData("startTime", startTime.UnixNano()/1e6)
		tracer.SetExtraData("endTime", endTime.UnixNano()/1e6)
		tracer.SetExtraData("durationTime", endTime.Sub(startTime).Nanoseconds()/1e6)

		// for k, v := range config.TracerData {
		// 	tracer.SetExtraData(k, v)
		// }

		if tracer != nil {
			err = tracer.Send()
			if err != nil {
				return nil, err
			}
		}

		return innerHttpResp, nil
	}

	httpResp, err := sendWithTracer("Send")

	// exponential backoff delay, maximum 8 minute about.
	cfg := c.GetConfig()
	for retryCount := 1; utils.IsRetryableError(err) || (httpResp != nil && utils.IsRetryableHTTPStatusCode(httpResp.StatusCode)); retryCount++ {
		if retryCount > cfg.MaxRetries {
			break
		}

		delay := getExpBackoffDelay(retryCount)
		time.Sleep(delay)

		httpResp, err = sendWithTracer("Retry")
	}

	if err != nil {
		if utils.IsTimeoutError(err) {
			return uerr.TimeoutError
		} else {
			return uerr.SendRequestError
		}
	}

	if utils.IsErrorHTTPStatusCode(httpResp.StatusCode) {
		return uerr.SendRequestError
	}

	return nil
}

// InvokeAction will do an action request from a request struct and set response value into res struct pointer
func (c *Client) InvokeAction(action string, req request.Common, resp response.Common) error {
	var err error
	cfg := c.GetConfig()
	defer logForAction(cfg.Logger, action, resp.(response.Common), &err)

	// Build query
	query, err := utils.RequestToQuery(req)
	if err != nil {
		return err
	}

	logrus.Infof("Request %#v", query)

	query["Action"] = action
	if region := req.GetRegion(); err != nil && len(region) > 0 {
		query["Region"] = region
	}
	if projectId := req.GetProjectId(); err != nil && len(projectId) > 0 {
		query["ProjectId"] = projectId
	}

	// Build request
	httpReq := &request.HttpRequest{
		Url:    cfg.BaseUrl,
		Method: "GET",
		Query:  query,
	}

	// Send request
	err = c.DoRequest(httpReq, resp)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) buildSuperAgent(req *request.HttpRequest) (*gorequest.SuperAgent, error) {
	config := c.GetConfig()
	r := gorequest.New()
	r.ClearSuperAgent()

	for k, v := range utils.MergeMap(DefaultHeaders, req.Header) {
		r.Set(k, v)
	}

	goVersion := runtime.Version()
	ua := fmt.Sprintf("GO/%s GO-SDK/%s", goVersion, version.Version)
	if config.UserAgent != "" {
		ua += " " + config.UserAgent
	}
	r.Set("User-Agent", ua)

	r.Method = req.Method
	if !utils.IsAvaliableMethod(r.Method) {
		logrus.Errorf("invalid method %s", r.Method)
		return nil, uerr.InvalidRequestError
	}

	r.Url = req.Url
	_, err := url.ParseRequestURI(req.Url)
	if err != nil {
		logrus.Error(err)
		return nil, uerr.InvalidRequestError
	}

	if len(req.Content) != 0 {
		r.Send(req.Content)
	}

	r.Errors = nil

	credential := c.GetCredential()
	if credential == nil {
		logrus.Errorln("invalid credential infomation, please set it before request.")
		return nil, uerr.InvalidRequestError
	}

	cfg := c.GetConfig()
	r.Timeout(cfg.Timeout)

	r.Query(credential.BuildCredentialedQuery(req.Query))
	return r, nil
}

func (c *Client) send(r *gorequest.SuperAgent, reqType string) (gorequest.Response, []byte, error) {
	var err error
	logger := c.GetConfig().Logger.WithField("type", reqType)

	logger.Debugf("%s %s?%s", r.Method, r.Url, r.QueryData.Encode())

	resp, body, errs := r.EndBytes()

	if len(errs) > 0 {
		err = errs[0]
		logger.Errorf("%T: %s", err, err)
	} else if resp != nil && resp.StatusCode > 400 {
		logger.Warnf("%s", resp.Status)
	} else {
		err = nil
		logger.Debugf("%s - %v", body, resp.StatusCode)
	}

	return resp, body, err
}

func logForAction(logger logrus.FieldLogger, action string, resp response.Common, err *error) {
	logger = logger.WithField("action", action)
	if err != nil && *err != nil {
		logger.Errorf("Do %s faild, %s", action, (*err).Error())
	} else if resp != nil && resp.GetRetCode() != 0 {
		logger.Errorf("Do %s faild, %s", action, resp.GetMessage())
	} else {
		logger.Infof("Do %s successful!", action)
	}
}

func getExpBackoffDelay(retryCount int) time.Duration {
	minTime := 100
	if retryCount > 7 {
		retryCount = 7
	}

	delay := (1 << (uint(retryCount) * 2)) * (rand.Intn(minTime) + minTime)
	return time.Duration(delay) * time.Millisecond
}
