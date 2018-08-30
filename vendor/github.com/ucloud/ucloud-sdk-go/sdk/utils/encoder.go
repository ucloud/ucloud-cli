package utils

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

const NIL = "nil"

// StructEncoder convert struct to map[string][string]
// Only allowed 1-layer struct, no recursive
type StructEncoder struct {
	IsZeroValueOmitted bool
}

func NewStructEncoder(isZeroValueOmitted bool) *StructEncoder {
	return &StructEncoder{
		IsZeroValueOmitted: isZeroValueOmitted,
	}
}

func (e *StructEncoder) encodeInt(v *reflect.Value) (string, error) {
	realV := v.Int()
	if e.IsZeroValueOmitted && IsZeroValue(realV) {
		return NIL, nil
	}
	return strconv.FormatInt(realV, 10), nil
}

func (e *StructEncoder) encodeUint(v *reflect.Value) (string, error) {
	realV := v.Uint()
	if e.IsZeroValueOmitted && IsZeroValue(realV) {
		return NIL, nil
	}
	return strconv.FormatUint(realV, 10), nil
}

func (e *StructEncoder) encodeBool(v *reflect.Value) (string, error) {
	realV := v.Bool()
	if e.IsZeroValueOmitted && IsZeroValue(realV) {
		return NIL, nil
	}
	return strconv.FormatBool(realV), nil
}

func (e *StructEncoder) encodeString(v *reflect.Value) (string, error) {
	realV := v.String()
	if e.IsZeroValueOmitted && IsZeroValue(realV) {
		return NIL, nil
	}
	return realV, nil
}

func (e *StructEncoder) encodeFloat(v *reflect.Value) (string, error) {
	realV := v.Float()
	if e.IsZeroValueOmitted && IsZeroValue(realV) {
		return NIL, nil
	}
	return strconv.FormatFloat(realV, 'E', -1, 64), nil
}

func (e *StructEncoder) encodeInterface(v *reflect.Value) (string, error) {
	// TODO: ...
	return NIL, nil
}

func (e *StructEncoder) encodeArray(v *reflect.Value) ([]string, error) {
	result := make([]string, v.Len())
	for i := 0; i < v.Len(); i++ {
		item := v.Index(i)

		encoded, err := e.encodeBuiltin(&item)
		if err != nil {
			return make([]string, 0), err
		}

		result[i] = encoded
	}
	return result, nil
}

func (e *StructEncoder) encodeBuiltin(v *reflect.Value) (string, error) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return e.encodeInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return e.encodeUint(v)
	case reflect.Float32, reflect.Float64:
		return e.encodeFloat(v)
	case reflect.Bool:
		return e.encodeBool(v)
	case reflect.String:
		return e.encodeString(v)
	case reflect.Ptr:
		ptrValue := v.Elem()
		return e.encodeBuiltin(&ptrValue)
	default:
		return "", errors.New(fmt.Sprintf("Invalid variable type, type must be one of int-, uint-, float-, bool, string and ptr, got %s", v.Kind().String()))
	}
}

func (e *StructEncoder) encode(v *reflect.Value) (map[string]string, error) {
	result := make(map[string]string)

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		name := v.Type().Field(i).Name

		switch f.Kind() {
		case reflect.Slice, reflect.Array:
			encodedArray, err := e.encodeArray(&f)
			if err != nil {
				return result, err
			}

			for index, encoded := range encodedArray {
				if len(encoded) > 0 {
					result[fmt.Sprintf("%s.%v", name, index)] = encoded
				}
			}
		case reflect.Interface:
			// TODO: implement ISO8601/RFC3339 like ucloudgo
			continue
		case reflect.Struct:
			// resolve composite common struct
			for i := 0; i < f.NumField(); i++ {
				composited := f.Field(i)
				name := f.Type().Field(i).Name

				encoded, err := e.encodeBuiltin(&composited)
				if err != nil {
					return result, err
				}

				if encoded != NIL {
					result[name] = encoded
				}
			}
		default:
			encoded, err := e.encodeBuiltin(&f)
			if err != nil {
				return result, err
			}

			if encoded != NIL {
				result[name] = encoded
			}
		}
	}
	return result, nil
}
