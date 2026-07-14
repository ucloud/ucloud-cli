// Package service ...
//
// @Brief  国内高防服务列表行结构体定义
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
package service

// ServiceRow 国内高防服务列表行
type ServiceRow struct {
	ResourceID    string
	Name          string
	DefenceStatus string
	ExpireTime    string
}
