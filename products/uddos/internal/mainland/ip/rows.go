// Package ip ...
//
// @Brief  国内高防IP列表行结构体定义
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

// IPRow 高防IP列表行
type IPRow struct {
	DefenceIP           string
	UserIP              string
	LineType            string
	Status              string
	Cname               string
	RuleCnt             int
	DefenceDDosBaseFlow int
	DefenceDDosMaxFlow  int
	Remark              string
}
