package request

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"

	"github.com/ucloud/ucloud-sdk-go/private/protocol/http"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
	"github.com/ucloud/ucloud-sdk-go/ucloud/config"
)

type FormEncoder struct {
	cfg  *config.Config
	cred *auth.Credential
}

// Encode will return a map for url form encoded
func (e *FormEncoder) Encode(req Common) (*http.HttpRequest, error) {
	if req == nil {
		return nil, fmt.Errorf("invalid request, got nil")
	}
	httpReq := http.NewHttpRequest()
	_ = httpReq.SetURL(e.cfg.BaseUrl)
	_ = httpReq.SetTimeout(req.GetTimeout())
	_ = httpReq.SetMethod("POST")
	_ = httpReq.SetQuery("Action", req.GetAction()) // workaround for http log handler
	_ = httpReq.SetHeader(http.HeaderNameContentType, http.MimeFormURLEncoded)

	// encode struct to map
	form, err := EncodeForm(req)
	if err != nil {
		return nil, err
	}
	payload := make(map[string]interface{})
	for k, v := range form {
		payload[k] = v
	}
	payload = e.cred.Apply(payload)

	// marshal payload as request body
	values := url.Values{}
	for k, v := range payload {
		values.Set(k, v.(string))
	}
	bs := values.Encode()
	_ = httpReq.SetRequestBody([]byte(bs))
	return httpReq, nil
}

func NewFormEncoder(cfg *config.Config, cred *auth.Credential) Encoder {
	return &FormEncoder{cfg: cfg, cred: cred}
}

func EncodeForm(req Common) (map[string]string, error) {
	m, err := structToMap(req)
	if err != nil {
		return nil, err
	}

	rv := reflect.ValueOf(m)
	query, err := encodeMapToForm(&rv, "")
	if err != nil {
		return nil, err
	}
	return query, nil
}

// encodeMapToForm will expand array and map as `.N` style
func encodeMapToForm(rv *reflect.Value, prefix string) (map[string]string, error) {
	result := make(map[string]string)

	for _, mapKey := range rv.MapKeys() {
		f := rv.MapIndex(mapKey)
		for f.Kind() == reflect.Ptr || f.Kind() == reflect.Interface {
			if f.IsNil() {
				break
			}
			f = f.Elem()
		}

		// check if nil-pointer
		if f.Kind() == reflect.Ptr && f.IsNil() {
			continue
		}

		name := mapKey.String()
		if prefix != "" {
			name = fmt.Sprintf("%s.%s", prefix, name)
		}

		switch f.Kind() {
		case reflect.Slice, reflect.Array:
			for n := 0; n < f.Len(); n++ {
				item := f.Index(n)
				for item.Kind() == reflect.Ptr || item.Kind() == reflect.Interface {
					if f.IsNil() {
						break
					}
					item = item.Elem()
				}

				if item.Kind() == reflect.Ptr && item.IsNil() {
					continue
				}

				keyPrefix := fmt.Sprintf("%s.%v", name, n)
				switch item.Kind() {
				case reflect.Map:
					kv, err := encodeMapToForm(&item, keyPrefix)
					if err != nil {
						return result, err
					}

					for k, v := range kv {
						if v != "" {
							result[k] = v
						}
					}
				default:
					s, err := encodeOne(&item)
					if err != nil {
						return result, err
					}

					if s != "" {
						result[keyPrefix] = s
					}
				}
			}
		case reflect.Map:
			kv, err := encodeMapToForm(&f, name)
			if err != nil {
				return result, err
			}

			for k, v := range kv {
				if v != "" {
					result[k] = v
				}
			}
		default:
			s, err := encodeOne(&f)
			if err != nil {
				return result, err
			}

			// set field value into result
			if s != "" {
				result[name] = s
			}
		}
	}

	return result, nil
}

// encodeOne will encode any value as string
func encodeOne(v *reflect.Value) (string, error) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64), nil
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil
	case reflect.String:
		return v.String(), nil
	case reflect.Ptr, reflect.Interface:
		ptrValue := v.Elem()
		return encodeOne(&ptrValue)
	default:
		message := fmt.Sprintf(
			"Invalid variable type, type must be one of int-, uint-,"+
				" float-, bool, string and ptr, got %s",
			v.Kind().String(),
		)
		return "", errors.New(message)
	}
}
