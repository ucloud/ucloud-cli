package http

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/ucloud/ucloud-sdk-go/private/utils"
)

var availableHTTPMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTION", "HEAD", "PATCH"}

// HttpRequest is the internal http request of sdk, don't use it at your code
type HttpRequest struct {
	url         string
	method      string
	queryMap    map[string]string
	queryString string
	headers     map[string]string
	requestBody []byte
	timeout     time.Duration
}

// NewHttpRequest will create a http request
func NewHttpRequest() *HttpRequest {
	r := &HttpRequest{
		queryMap: make(map[string]string),
		headers:  make(map[string]string),
		timeout:  DefaultTimeout,
	}

	for k, v := range DefaultHeaders {
		r.headers[k] = v
	}
	return r
}

// SetURL will set url into request
func (h *HttpRequest) SetURL(val string) error {
	// check url is valid
	uri, err := url.ParseRequestURI(val)
	if err != nil {
		return errors.Errorf("url is invalid, got %s", val)
	}

	err = h.SetQueryString(uri.RawQuery)
	if err != nil {
		return err
	}

	h.url = fmt.Sprintf("%s://%s%s", uri.Scheme, uri.Host, uri.Path)
	return nil
}

// GetURL will get request url value
func (h *HttpRequest) GetURL() string {
	return h.url
}

// SetMethod will set method of current request
func (h *HttpRequest) SetMethod(val string) error {
	err := utils.CheckStringIn(val, availableHTTPMethods)
	if err != nil {
		return errors.Errorf("method is invalid, %s", err)
	}

	h.method = strings.ToUpper(val)
	return nil
}

// GetMethod will get request url value
func (h *HttpRequest) GetMethod() string {
	return h.method
}

// SetQueryString will set query map by query string,
// it also save as query string attribute to keep query ordered.
func (h *HttpRequest) SetQueryString(val string) error {
	// check url query is valid
	values, err := url.ParseQuery(val)
	if err != nil {
		return errors.Errorf("url query is invalid, got %s", val)
	}

	// copy url query into request query map, it will overwrite current query
	for k, v := range values {
		if len(v) > 0 {
			_ = h.SetQuery(k, v[0])
		}
	}

	h.queryString = val
	return nil
}

// BuildQueryString will return the query string of this request,
// it will also append key-value of query map after existed query string
func (h *HttpRequest) BuildQueryString() (string, error) {
	values := url.Values{}
	for k, v := range h.queryMap {
		values.Add(k, v)
	}

	// if query string is not set by user,
	// otherwise needn't keep them ordered, encode immediately.
	if h.queryString == "" {
		return values.Encode(), nil
	}

	// exclude query that existed in query string pass by user,
	// to keep ordered from user definition
	existsValues, _ := url.ParseQuery(h.queryString)
	for k := range existsValues {
		values.Del(k)
	}

	// append query map after existed query string
	qs := h.queryString
	if len(values) > 0 {
		qs += "&" + values.Encode()
	}

	return qs, nil
}

// SetQuery will store key-value data into query map
func (h *HttpRequest) SetQuery(k, v string) error {
	if h.queryMap == nil {
		h.queryMap = make(map[string]string)
	}
	h.queryMap[k] = v
	return nil
}

// GetQuery will get value by key from map
func (h *HttpRequest) GetQuery(k string) string {
	if v, ok := h.queryMap[k]; ok {
		return v
	}
	return ""
}

// GetQueryMap will get all of query as a map
func (h *HttpRequest) GetQueryMap() map[string]string {
	return h.queryMap
}

// SetTimeout will set timeout of current request
func (h *HttpRequest) SetTimeout(val time.Duration) error {
	h.timeout = val
	return nil
}

// GetTimeout will get timeout of current request
func (h *HttpRequest) GetTimeout() time.Duration {
	return h.timeout
}

// SetHeader will set http header of current request
func (h *HttpRequest) SetHeader(k, v string) error {
	if h.headers == nil {
		h.headers = make(map[string]string)
	}
	h.headers[k] = v
	return nil
}

// GetHeaderMap wiil get all of header as a map
func (h *HttpRequest) GetHeaderMap() map[string]string {
	return h.headers
}

// SetRequestBody will set http body of current request
func (h *HttpRequest) SetRequestBody(val []byte) error {
	h.requestBody = val
	return nil
}

// GetRequestBody will get origin http request ("net/http")
func (h *HttpRequest) GetRequestBody() []byte {
	return h.requestBody
}

func (h *HttpRequest) String() string {
	s := h.GetURL()

	// resolve query
	qs, err := h.BuildQueryString()
	if err != nil {
		return s
	}
	if len(qs) > 0 {
		s = fmt.Sprintf("%s?%s", s, qs)
	}

	// resolve body
	bs := h.GetRequestBody()
	if len(bs) > 0 {
		s = fmt.Sprintf("%s %s", s, string(bs))
	}
	return s
}

func (h *HttpRequest) buildHTTPRequest() (*http.Request, error) {
	qs, err := h.BuildQueryString()
	if err != nil {
		return nil, err
	}

	uri := h.GetURL()
	if qs != "" {
		uri = fmt.Sprintf("%s?%s", uri, qs)
	}

	httpReq, err := http.NewRequest(h.GetMethod(), uri, bytes.NewReader(h.GetRequestBody()))
	if err != nil {
		return nil, errors.Errorf("cannot build request, %s", err)
	}

	for k, v := range utils.MergeMap(DefaultHeaders, h.GetHeaderMap()) {
		httpReq.Header.Set(k, v)
	}
	return httpReq, nil
}
