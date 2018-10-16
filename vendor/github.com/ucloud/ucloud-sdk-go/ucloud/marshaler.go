package ucloud

import (
	"encoding/json"
	"fmt"
	"regexp"
	"runtime"

	"github.com/pkg/errors"

	"github.com/ucloud/ucloud-sdk-go/private/protocol/http"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"
	"github.com/ucloud/ucloud-sdk-go/ucloud/version"
)

// SetupRequest will init request by client configuration
func (c *Client) SetupRequest(req request.Common) request.Common {
	cfg := c.GetConfig()

	req.SetRetryable(true)

	// set optional client level variables
	if len(req.GetRegion()) == 0 && len(cfg.Region) > 0 {
		req.SetRegion(cfg.Region)
	}

	if len(req.GetProjectId()) == 0 && len(cfg.ProjectId) > 0 {
		req.SetProjectId(cfg.ProjectId)
	}

	if req.GetTimeout() == 0 && cfg.Timeout != 0 {
		req.WithTimeout(cfg.Timeout)
	}

	if req.GetMaxretries() == 0 && cfg.MaxRetries != 0 {
		req.WithRetry(cfg.MaxRetries)
	}

	return req
}

func (c *Client) buildHTTPRequest(req request.Common) (*http.HttpRequest, error) {
	// convert request struct to query map
	query, err := request.ToQueryMap(req)
	if err != nil {
		return nil, errors.Errorf("convert request to map failed, %s", err)
	}

	// check credential information is avaliable
	credential := c.GetCredential()
	if credential == nil {
		return nil, errors.Errorf("invalid credential infomation, please set it before request.")
	}

	config := c.GetConfig()
	httpReq := http.NewHttpRequest()
	httpReq.SetURL(config.BaseUrl)
	httpReq.SetMethod("GET")

	// set timeout with client configuration
	httpReq.SetTimeout(config.Timeout)

	// keep query stirng is ordered and append credential signiture as the last query param
	httpReq.SetQueryString(credential.BuildCredentialedQuery(query))

	ua := fmt.Sprintf("GO/%s GO-SDK/%s %s", runtime.Version(), version.Version, config.UserAgent)
	httpReq.SetHeader("User-Agent", ua)

	return &httpReq, nil
}

// unmarshalHTTPReponse will get body from http response and unmarshal it's data into response struct
func (c *Client) unmarshalHTTPReponse(httpResp *http.HttpResponse, resp response.Common) error {
	body := httpResp.GetBody()
	if len(body) < 0 {
		return nil
	}

	body = patchForRetCodeString(body)
	return json.Unmarshal([]byte(body), &resp)
}

var patchForCodePattern = regexp.MustCompile(`"RetCode":\s*"(\d+)"`)

func patchForRetCodeString(body []byte) []byte {
	return patchForCodePattern.ReplaceAll(body, []byte(`"RetCode": $1`))
}
