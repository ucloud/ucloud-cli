// Package rule ...
//
// @Brief  国内高防转发规则列表行结构体定义
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
package rule

// RuleRow 转发规则列表行
type RuleRow struct {
	RuleIndex    string
	RuleID       string
	BgpIP        string
	SourceIP     string
	FwdType      string
	BgpIPPort    string
	LoadBalance  string
	SourceDetect string
	Remark       string
}
