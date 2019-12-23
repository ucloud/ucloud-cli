package pathx

/*
GlobalSSHInfo - GlobalSSH实例信息

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type GlobalSSHInfo struct {

	// 实例ID，资源唯一标识
	InstanceId string

	// 加速域名
	AcceleratingDomain string

	// 被SSH访问的IP所在地区
	Area string

	// 被SSH访问的EIP
	TargetIP string

	// 备注信息
	Remark string

	// SSH登陆端口
	Port int

	// 支付周期，如Month,Year等
	ChargeType string

	// 资源创建时间戳
	CreateTime int

	// 资源过期时间戳
	ExpireTime int
}
