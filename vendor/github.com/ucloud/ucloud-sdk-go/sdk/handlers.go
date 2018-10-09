package sdk

import (
	"math/rand"
	"time"

	uerr "github.com/ucloud/ucloud-sdk-go/sdk/error"
	"github.com/ucloud/ucloud-sdk-go/sdk/log"
	"github.com/ucloud/ucloud-sdk-go/sdk/protocol/http"
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

// ReponseHandler receive response and write data into this response memory area
type ReponseHandler func(c *Client, req request.Common, resp response.Common, err error) (response.Common, error)

// HttpReponseHandler receive http response and return a new http response
type HttpReponseHandler func(c *Client, req *http.HttpRequest, resp *http.HttpResponse, err error) (*http.HttpResponse, error)

var defaultResponseHandlers = []ReponseHandler{errorHandler, logHandler, retryHandler}
var defaultHttpResponseHandlers = []HttpReponseHandler{errorHTTPHandler, logDebugHTTPHandler}

func retryHandler(c *Client, req request.Common, resp response.Common, err error) (response.Common, error) {
	for retryCount := 1; err != nil && err.(uerr.Error).Retryable(); retryCount++ {
		// if max retries number is reached, stop and raise last error
		if retryCount > req.GetMaxretries() {
			break
		}

		// use exponential backoff constant as retry delay
		delay := getExpBackoffDelay(retryCount)
		time.Sleep(delay)

		req.SetRetryCount(retryCount)

		// the resp will be changed after invoke
		err = c.InvokeAction(req.GetAction(), req, resp)
	}

	return resp, err
}

func getExpBackoffDelay(retryCount int) time.Duration {
	minTime := 100
	if retryCount > 7 {
		retryCount = 7
	}

	delay := (1 << (uint(retryCount) * 2)) * (rand.Intn(minTime) + minTime)
	return time.Duration(delay) * time.Millisecond
}

// errorHandler will normalize error to several specific error
func errorHandler(c *Client, req request.Common, resp response.Common, err error) (response.Common, error) {
	if err != nil {
		if _, ok := err.(uerr.Error); ok {
			return resp, err
		}
		if uerr.IsNetworkError(err) {
			return resp, uerr.NewClientError(uerr.ErrNetwork, err)
		}
		return resp, uerr.NewClientError(uerr.ErrSendRequest, err)
	}

	if resp.GetRetCode() != 0 {
		return resp, uerr.NewServerCodeError(resp.GetRetCode(), resp.GetMessage())
	}

	return resp, err
}

func errorHTTPHandler(c *Client, req *http.HttpRequest, resp *http.HttpResponse, err error) (*http.HttpResponse, error) {
	if statusErr, ok := err.(http.StatusError); ok {
		return resp, uerr.NewServerStatusError(statusErr.StatusCode, statusErr.Message)
	}
	return resp, err
}

func logHandler(c *Client, req request.Common, resp response.Common, err error) (response.Common, error) {
	action := req.GetAction()
	if err != nil {
		log.Errorf("do %s failed, %s", action, err)
	} else {
		log.Infof("do %s successful!", action)
	}
	return resp, err
}

func logDebugHTTPHandler(c *Client, req *http.HttpRequest, resp *http.HttpResponse, err error) (*http.HttpResponse, error) {
	// TODO: move request debugging log to request handler
	log.Debugf("%s", req)

	if err != nil {
		log.Errorf("%s", err)
	} else if resp.GetStatusCode() > 400 {
		log.Warnf("%s", resp.GetStatusCode())
	} else {
		log.Debugf("%s - %v", resp.GetBody(), resp.GetStatusCode())
	}

	return resp, err
}
