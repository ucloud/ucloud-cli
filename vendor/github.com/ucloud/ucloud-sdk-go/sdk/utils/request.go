package utils

import (
	"reflect"

	"github.com/Sirupsen/logrus"

	uerr "github.com/ucloud/ucloud-sdk-go/sdk/error"
)

// RequestToQuery used to convert an request struct to query
func RequestToQuery(req interface{}) (map[string]string, error) {
	vReq := reflect.ValueOf(req)
	if vReq.Kind() != reflect.Ptr {
		logrus.Errorf("Request has type %s, want struct pointer.", vReq.Kind().String())
		return make(map[string]string), uerr.InvalidRequestError
	}

	v := vReq.Elem()
	if v.Kind() != reflect.Struct {
		logrus.Errorf("Request has type %s, want struct pointer.", vReq.Kind().String())
		return make(map[string]string), uerr.InvalidRequestError
	}

	encoder := NewStructEncoder(true)
	return encoder.encode(&v)
}
