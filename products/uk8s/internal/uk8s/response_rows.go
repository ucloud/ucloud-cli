package uk8s

import (
	"fmt"
	"reflect"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

// responseRows expands an SDK response into attribute rows for table output.
// Keeping the expansion reflective makes table output follow SDK response
// additions without silently dropping newly generated fields. JSON/YAML output
// uses the original SDK response directly.
func responseRows(response interface{}) []cli.DescribeRow {
	rows := make([]cli.DescribeRow, 0)
	appendResponseRows(&rows, "", reflect.ValueOf(response))
	return rows
}

func appendResponseRows(rows *[]cli.DescribeRow, prefix string, value reflect.Value) {
	if !value.IsValid() {
		*rows = append(*rows, cli.DescribeRow{Attribute: prefix, Content: "<nil>"})
		return
	}

	for value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
		if value.IsNil() {
			*rows = append(*rows, cli.DescribeRow{Attribute: prefix, Content: "<nil>"})
			return
		}
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.Struct:
		typeOfValue := value.Type()
		for i := 0; i < value.NumField(); i++ {
			field := typeOfValue.Field(i)
			if field.PkgPath != "" {
				continue
			}
			fieldPrefix := field.Name
			if prefix != "" {
				fieldPrefix = prefix + "." + fieldPrefix
			}
			appendResponseRows(rows, fieldPrefix, value.Field(i))
		}
	case reflect.Slice, reflect.Array:
		if value.Len() == 0 {
			*rows = append(*rows, cli.DescribeRow{Attribute: prefix, Content: "[]"})
			return
		}
		for i := 0; i < value.Len(); i++ {
			appendResponseRows(rows, fmt.Sprintf("%s[%d]", prefix, i), value.Index(i))
		}
	case reflect.Map:
		if value.Len() == 0 {
			*rows = append(*rows, cli.DescribeRow{Attribute: prefix, Content: "{}"})
			return
		}
		iter := value.MapRange()
		for iter.Next() {
			appendResponseRows(rows, fmt.Sprintf("%s[%v]", prefix, iter.Key().Interface()), iter.Value())
		}
	default:
		*rows = append(*rows, cli.DescribeRow{Attribute: prefix, Content: fmt.Sprint(value.Interface())})
	}
}
