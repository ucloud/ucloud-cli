package sdk

import (
	"time"

	"github.com/ucloud/ucloud-sdk-go/sdk/auth"
	"github.com/ucloud/ucloud-sdk-go/sdk/log"
	"github.com/ucloud/ucloud-sdk-go/sdk/protocol/http"
	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

//Client 客户端
type Client struct {
	credential *auth.Credential
	config     *Config
	httpClient *http.HttpClient

	responseHandlers     []ReponseHandler
	httpResponseHandlers []HttpReponseHandler
}

// NewClient will create an client of ucloud sdk
func NewClient(config *Config, credential *auth.Credential) *Client {
	client := Client{
		credential: credential,
		config:     config,
	}

	client.responseHandlers = append(client.responseHandlers, defaultResponseHandlers...)
	client.httpResponseHandlers = append(client.httpResponseHandlers, defaultHttpResponseHandlers...)
	log.Init(config.LogLevel)
	return &client
}

// GetCredential will return the creadential config of client.
func (c *Client) GetCredential() *auth.Credential {
	return c.credential
}

// GetConfig will return the config of client.
func (c *Client) GetConfig() *Config {
	return c.config
}

// InvokeAction will do an action request from a request struct and set response value into res struct pointer
func (c *Client) InvokeAction(action string, req request.Common, resp response.Common) error {
	c.SetupRequest(req, action)
	req.SetRequestTime(time.Now())

	httpReq, err := c.buildHTTPRequest(req)
	if err != nil {
		return err
	}

	httpClient := http.NewHttpClient()
	httpResp, err := httpClient.Send(httpReq)
	if err != nil {
		return err
	}

	// use response middleware to handle http response
	// such as convert some http status to error
	for _, handler := range c.httpResponseHandlers {
		httpResp, err = handler(c, httpReq, httpResp, err)
	}

	err = c.UnmarshalHTTPReponse(httpResp, resp)
	if err != nil {
		return err
	}

	// use response middle to build and convert response when response has been created.
	// such as retry, report traceback, print log and etc.
	for _, handler := range c.responseHandlers {
		resp, err = handler(c, req, resp, err)
	}

	return err
}
