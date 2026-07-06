// Package common holds non-product, dependency-free utilities shared across
// the platform and the product modules.
//
// Unlike base/ (which is platform-internal and forbidden to products by §6.1),
// anything here is importable by products under products/<name>/. Keep it pure:
// no platform singletons, no SDK clients, no authentication/config state, no
// I/O beyond the standard library. See docs §2 "目录归属判据" and §4.6.
package common

import "time"

// DateTimeLayout is the canonical timestamp layout used across the CLI.
// Verbatim from base.DateTimeLayout.
const DateTimeLayout = "2006-01-02/15:04:05"

// FormatDateTime formats a unix-second timestamp as DateTimeLayout.
// Verbatim from base.FormatDateTime.
func FormatDateTime(seconds int) string {
	return time.Unix(int64(seconds), 0).Format("2006-01-02/15:04:05")
}

// FormatDate 格式化时间，把以秒为单位的时间戳格式化为年月日
func FormatDate(seconds int) string {
	return time.Unix(int64(seconds), 0).Format("2006-01-02")
}
