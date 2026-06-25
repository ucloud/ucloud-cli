// encoding.go —— encoding 纯工具（stdlib-only）。
package common

import "encoding/base64"

// IsBase64Encoded 判断字节是否为合法的标准 base64 编码
func IsBase64Encoded(data []byte) bool {
	_, err := base64.StdEncoding.DecodeString(string(data))
	return err == nil
}
