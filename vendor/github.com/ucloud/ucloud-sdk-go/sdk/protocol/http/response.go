package http

import (
	"net/http"
)

// HttpResponse is a simple wrapper of "net/http" response
type HttpResponse struct {
	body               []byte
	originHttpResponse *http.Response // origin "net/http" response
}

// NewHttpResponse will create a new response of http request
func NewHttpResponse() *HttpResponse {
	return &HttpResponse{}
}

// GetBody will get body from from sdk http request
func (h *HttpResponse) GetBody() []byte {
	return h.body
}

// GetStatusCode will return status code of origin http response
func (h *HttpResponse) GetStatusCode() int {
	return h.originHttpResponse.StatusCode
}

// setBody will set body into http response
// it usually used for restore the body already read from an stream
// it will also cause extra memory usage
func (h *HttpResponse) setBody(body []byte) error {
	h.body = body
	return nil
}

func (h *HttpResponse) setHttpReponse(resp *http.Response) {
	h.originHttpResponse = resp
}
