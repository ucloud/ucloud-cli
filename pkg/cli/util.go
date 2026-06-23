package cli

import (
	"fmt"
	"io/ioutil" //nolint:staticcheck // keep ioutil for zero-behavior-change verbatim copy
	"strings"

	uerr "github.com/ucloud/ucloud-sdk-go/ucloud/error"
)

// PickResourceID extracts the resource ID from a "resourceID/name" string.
// Example: "uhost-xxx/uhost-name" => "uhost-xxx"
func PickResourceID(str string) string {
	if strings.Index(str, "/") > -1 {
		return strings.SplitN(str, "/", 2)[0]
	}
	return str
}

// ParseError converts an error to a human-readable string.
func ParseError(err error) string {
	if uErr, ok := err.(uerr.Error); ok && uErr.Code() != 0 {
		format := "Something wrong. RetCode:%d. Message:%s"
		message := uErr.Message()
		if uErr.Code() == -1 || uErr.Code() == -2 {
			message = "request timeout, retry later please"
		}
		return fmt.Sprintf(format, uErr.Code(), message)
	}
	return fmt.Sprintf("Error:%v", err)
}

// ReadFile reads the contents of the named file and returns them as a string.
// Relocated verbatim from cmd/ulb.go readFile with exported name.
func ReadFile(file string) (string, error) {
	byts, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(byts), nil
}
