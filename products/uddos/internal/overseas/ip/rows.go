// Package ip ...
//
// @Brief  海外高防IP列表行结构体定义
//
// @File   rows.go
//
// @Author leas.li(cc)
//
// @Email  leas.li@ucloud.cn
//
// @Date   2026/07/11
//
// @CopyRights(C) UCloud All rights reserved.
package ip

// IPRow 高防IP列表行（透传模式）
type IPRow struct {
	EIPIP     string
	EIPID     string
	Status    string
	EIPRegion string
	Tag       string
	Remark    string
}
