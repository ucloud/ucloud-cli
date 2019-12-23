package pathx

/*
UGAATask - 用户在UGAA实例下配置的多端口任务

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type UGAATask struct {

	// 端口
	Port int

	// 接入协议,包括TCP|UDP
	Protocol string

	// 转发协议，包括TCP|UDP|HTTP|HTTPS
	TransferProtocol string
}
