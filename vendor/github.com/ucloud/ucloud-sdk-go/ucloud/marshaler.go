package ucloud

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"

	"github.com/ucloud/ucloud-sdk-go/private/protocol/http"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/response"
	"github.com/ucloud/ucloud-sdk-go/ucloud/version"
)

// SetupRequest will init request by client configuration
func (c *Client) SetupRequest(req request.Common) request.Common {
	req.SetRetryable(true)
	cfg := c.GetConfig()
	if cfg == nil {
		return req
	}

	// set optional client level variables
	if len(req.GetRegion()) == 0 && len(cfg.Region) > 0 {
		_ = req.SetRegion(cfg.Region)
	}

	if len(req.GetZone()) == 0 && len(cfg.Zone) > 0 {
		_ = req.SetZone(cfg.Zone)
	}

	if len(req.GetProjectId()) == 0 && len(cfg.ProjectId) > 0 {
		_ = req.SetProjectId(cfg.ProjectId)
	}

	if req.GetTimeout() == 0 && cfg.Timeout != 0 {
		req.WithTimeout(cfg.Timeout)
	}

	if req.GetMaxretries() == 0 && cfg.MaxRetries != 0 {
		req.WithRetry(cfg.MaxRetries)
	}

	if req.GetEncoder() == nil {
		req.SetEncoder(request.NewFormEncoder(cfg, c.GetCredential()))
	}
	return req
}

func (c *Client) buildHTTPRequest(req request.Common) (*http.HttpRequest, error) {
	encoder := req.GetEncoder()
	if encoder == nil {
		encoder = request.NewFormEncoder(c.GetConfig(), c.GetCredential())
	}

	httpReq, err := encoder.Encode(req)
	if err != nil {
		return nil, err
	}

	product := c.GetMeta().Product
	if _, ok := req.(request.GenericRequest); ok {
		product = "@generic"
	}
	if product == "" {
		product = "@none"
	}

	ua := fmt.Sprintf(
		"GO/%s GO-SDK/%s Product/%s %s",
		runtime.Version(), version.Version, product, c.GetConfig().UserAgent,
	)
	_ = httpReq.SetHeader(http.HeaderNameUserAgent, strings.TrimSpace(ua))
	_ = httpReq.SetHeader(http.HeaderUTimestampMs, strconv.FormatInt(req.GetRequestTime().UnixNano()/1000000, 10))
	return httpReq, nil
}

// unmarshalHTTPResponse will get body from http response and unmarshal it's data into response struct
func (c *Client) unmarshalHTTPResponse(body []byte, resp response.Common) error {
	if len(body) == 0 {
		return uerr.NewEmptyResponseBodyError()
	}

	if r, ok := resp.(response.GenericResponse); ok {
		m := make(map[string]interface{})
		if err := json.Unmarshal(body, &m); err != nil {
			return uerr.NewResponseBodyError(err, string(body))
		}
		if err := r.SetPayload(m); err != nil {
			return uerr.NewResponseBodyError(err, string(body))
		}
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return uerr.NewResponseBodyError(err, string(body))
	}
	return nil
}
