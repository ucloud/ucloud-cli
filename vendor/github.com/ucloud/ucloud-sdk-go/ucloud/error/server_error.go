package uerr

import (
	"fmt"

	"github.com/pkg/errors"
)

var (
	// ErrHTTPStatus is error type of http status
	ErrHTTPStatus = "server.HTTPStatusError"
	// ErrRetCode is error type of server return code is larger than 0
	ErrRetCode = "server.RetCodeError"
	// ErrResponseBodyError is error type of server response body
	ErrResponseBodyError = "server.ResponseBodyError"
	// ErrEmptyResponseBodyError is empty of server response body
	ErrEmptyResponseBodyError = "server.EmptyResponseBodyError"
)

// ServerError is the ucloud common error for server response
type ServerError struct {
	err        error
	name       string
	statusCode int
	retCode    int
	message    string
	retryable  bool
}

func (e ServerError) Error() string {
	if e.retCode > 0 {
		return fmt.Sprintf("api:\n[%s] %v %s", e.name, e.retCode, e.message)
	}
	return fmt.Sprintf("api:\n[%s] %s", e.name, e.message)
}

// NewServerStatusError will return a new instance of NewServerStatusError
func NewServerStatusError(statusCode int, message string) ServerError {
	return ServerError{
		retCode:    -1,
		statusCode: statusCode,
		message:    message,
		name:       ErrHTTPStatus,
		err:        errors.Errorf("%s", message),
		retryable:  false,
	}
}

// NewServerCodeError will return a new instance of NewServerStatusError
func NewServerCodeError(retCode int, message string) ServerError {
	return ServerError{
		retCode:    retCode,
		statusCode: 200,
		message:    message,
		name:       ErrRetCode,
		err:        errors.Errorf("%s", message),
		retryable:  retCode >= 2000,
	}
}

// NewResponseBodyError will create a new response body error
func NewResponseBodyError(err error, body string) ServerError {
	message := fmt.Sprintf("response body\n[%v] got error, %s", body, err)
	return ServerError{
		name:       ErrResponseBodyError,
		err:        fmt.Errorf("%s", message),
		message:    message,
		statusCode: 200,
		retryable:  false,
	}
}

// NewEmptyResponseBodyError will create a new response body error
func NewEmptyResponseBodyError() ServerError {
	message := "response body got empty"
	return ServerError{
		name:       ErrEmptyResponseBodyError,
		err:        fmt.Errorf("%s", message),
		message:    message,
		statusCode: 200,
		retryable:  false,
	}
}

// Name will return error name
func (e ServerError) Name() string {
	return e.name
}

// Code will return server code
func (e ServerError) Code() int {
	return e.retCode
}

// StatusCode will return http status code
func (e ServerError) StatusCode() int {
	return e.statusCode
}

// Message will return message
func (e ServerError) Message() string {
	return e.message
}

// OriginError will return the origin error that caused by
func (e ServerError) OriginError() error {
	return e.err
}

// Retryable will return if the error is retryable
func (e ServerError) Retryable() bool {
	return isIn(e.statusCode, []int{429, 502, 503, 504}) || e.retryable
}

func isIn(i int, availables []int) bool {
	for _, v := range availables {
		if i == v {
			return true
		}
	}
	return false
}

// IsCodeError will check if the error is the retuen code error
func IsCodeError(err error) bool {
	if e, ok := err.(Error); ok && e.Name() == ErrRetCode {
		return true
	}
	return false
}
