package request

const (
	HTTP_REQUEST_TYPE_MULTIPART = "multipart"
	HTTP_REQUEST_TYPE_JSON      = "json"
	HTTP_REQUEST_TYPE_STRING    = "string"
)

type HttpRequest struct {
	Url     string
	Type    string
	Method  string
	Query   map[string]string
	Header  map[string]string
	Form    map[string]string
	Content []byte
}
