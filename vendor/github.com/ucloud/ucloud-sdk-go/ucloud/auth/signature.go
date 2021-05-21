package auth

import (
	"crypto/sha1"
	"encoding/hex"
	"reflect"
	"sort"
	"strconv"
)

// extractKeys extract all Keys from map[string]interface{}
func extractKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys
}

// sign generate signature
func sign(params map[string]interface{}, privateKey string) string {
	str := map2String(params) + privateKey
	hashed := sha1.Sum([]byte(str))
	return hex.EncodeToString(hashed[:])
}

// simple2String convert map type to string
func map2String(params map[string]interface{}) (str string) {
	for _, k := range extractKeys(params) {
		str += k + any2String(params[k])
	}
	return
}

// any2String convert any type to string
func any2String(v interface{}) string {
	switch v := v.(type) {
	case string, bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64,
		*string, *bool, *int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64, *float32, *float64:
		return simple2String(v)
	case []interface{}:
		return slice2String(v)
	case map[string]interface{}:
		return map2String(v)
	default:
		return reflect2String(reflect.ValueOf(v))
	}
}

// simple2String convert slice type to string
func slice2String(arr []interface{}) (str string) {
	for _, v := range arr {
		switch v := v.(type) {
		case string, bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64,
			*string, *bool, *int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64, *float32, *float64:
			str += simple2String(v)
		case []interface{}:
			str += slice2String(v)
		case map[string]interface{}:
			str += map2String(v)
		default:
			str += reflect2String(reflect.ValueOf(v))
		}
	}
	return
}

// simple2String convert simple type to string
func simple2String(v interface{}) string {
	switch v := v.(type) {
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	case int:
		return strconv.FormatInt(int64(v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatInt(int64(v), 10)
	case uint8:
		return strconv.FormatInt(int64(v), 10)
	case uint16:
		return strconv.FormatInt(int64(v), 10)
	case uint32:
		return strconv.FormatInt(int64(v), 10)
	case uint64:
		return strconv.FormatInt(int64(v), 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 64)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case *string:
		return *v
	case *bool:
		return strconv.FormatBool(*v)
	case *int:
		return strconv.FormatInt(int64(*v), 10)
	case *int8:
		return strconv.FormatInt(int64(*v), 10)
	case *int16:
		return strconv.FormatInt(int64(*v), 10)
	case *int32:
		return strconv.FormatInt(int64(*v), 10)
	case *int64:
		return strconv.FormatInt(*v, 10)
	case *uint:
		return strconv.FormatInt(int64(*v), 10)
	case *uint8:
		return strconv.FormatInt(int64(*v), 10)
	case *uint16:
		return strconv.FormatInt(int64(*v), 10)
	case *uint32:
		return strconv.FormatInt(int64(*v), 10)
	case *uint64:
		return strconv.FormatInt(int64(*v), 10)
	case *float32:
		return strconv.FormatFloat(float64(*v), 'f', -1, 64)
	case *float64:
		return strconv.FormatFloat(*v, 'f', -1, 64)
	}
	return ""
}

// reflect2String convert array and slice to string in reflect way
func reflect2String(rv reflect.Value) (str string) {
	if rv.Kind() != reflect.Array && rv.Kind() != reflect.Slice {
		return
	}

	for i := 0; i < rv.Len(); i++ {
		str += any2String(rv.Index(i).Interface())
	}
	return
}
