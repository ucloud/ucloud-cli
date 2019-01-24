package ulb

/*
ULBSSLSet - DescribeSSL

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type ULBSSLSet struct {

	// SSL证书的Id
	SSLId string

	// SSL证书的名字
	SSLName string

	// SSL证书类型，暂时只有 Pem 一种类型
	SSLType string

	// SSL证书的内容
	SSLContent string

	// SSL证书的创建时间
	CreateTime int

	// SSL证书绑定到的对象
	BindedTargetSet []SSLBindedTargetSet

	// 证书的 Hash 值
	HashValue string
}
