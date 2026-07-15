// Package service ...
//
// @Brief  国内高防服务生命周期状态定义与轮询辅助
//
// @File   status.go
//
// @Author leas.li(cc)
//
// @Email  leas.li@ucloud.cn
//
// @Date   2026/07/14
//
// @CopyRights(C) UCloud All rights reserved.
package service

// 服务生命周期状态：API 响应字段 DefenceStatus 为字符串，由 nap-api 的
// NapServiceStatus2Str 映射（NAP_SERVICE_STATUS_IS_NORMAL=1 -> "Started" 等）。
const (
	napServiceStatusStarted = "Started" // NAP_SERVICE_STATUS_IS_NORMAL(1)：创建完成、可用
	napServiceStatusStopped = "Stopped" // NAP_SERVICE_STATUS_IS_STOPPED(2)：已停用
	napServiceStatusExpired = "Expired" // NAP_SERVICE_STATUS_IS_EXPIRED(3)：已过期
)

// serviceStatusRow 是 poller 反射读取的最小结构体：它读取 Status 字段（字符串）
// 与 targetStates 比较（见 pkg/cli/poller.go state()，仅识别 State/Status 字段）。
type serviceStatusRow struct {
	Status string
}
