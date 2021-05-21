package request

import (
	"fmt"
	"reflect"
)

func structToMap(req Common) (map[string]interface{}, error) {
	if r, ok := req.(GenericRequest); ok {
		return r.GetPayload(), nil
	}

	v := reflect.ValueOf(req)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("invalid request type, must mu a struct/pointer")
	}

	payload := map[string]interface{}{}
	if v := req.GetAction(); v != "" {
		payload["Action"] = v
	}
	if v := req.GetRegion(); v != "" {
		payload["Region"] = v
	}
	if v := req.GetZone(); v != "" {
		payload["Zone"] = v
	}
	if v := req.GetProjectId(); v != "" {
		payload["ProjectId"] = v
	}

	params, err := reflectToMap(v)
	if err != nil {
		return nil, err
	}
	for k, v := range params {
		payload[k] = v
	}
	return payload, nil
}

func reflectToAny(v reflect.Value) (interface{}, error) {
	// find the real value of pointer
	// such as **struct to struct
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			break
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		return reflectToArray(v)
	case reflect.Struct:
		return reflectToMap(v)
	default:
		return v.Interface(), nil
	}
}

func reflectToArray(v reflect.Value) ([]interface{}, error) {
	var values []interface{}
	for i := 0; i < v.Len(); i++ {
		item := v.Index(i)
		if item.Kind() == reflect.Ptr && item.IsNil() {
			continue
		}
		kv, err := reflectToAny(item)
		if err != nil {
			return nil, err
		}
		values = append(values, kv)
	}
	return values, nil
}

func reflectToMap(v reflect.Value) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		name := v.Type().Field(i).Name

		// skip common base
		if name == "CommonBase" {
			continue
		}

		// skip unexported field
		if !f.CanSet() {
			continue
		}

		if f.Kind() == reflect.Ptr && f.IsNil() {
			continue
		}

		v, err := reflectToAny(f)
		if err != nil {
			return nil, err
		}
		m[name] = v
	}
	return m, nil
}
