package uhost

/*
CreateUHostInstanceParamNetworkInterfaceEIPGlobalSSH is request schema for complex param
*/
type CreateUHostInstanceParamNetworkInterfaceEIPGlobalSSH struct {

	// 填写支持SSH访问IP的地区名称，如“洛杉矶”，“新加坡”，“香港”，“东京”，“华盛顿”，“法兰克福”。Area和AreaCode两者必填其中之一。
	Area *string `required:"false"`

	// GlobalSSH的地区编码，格式为区域航空港国际通用代码。Area和AreaCode两者必填其中之一。
	AreaCode *string `required:"false"`

	// SSH端口，1-65535且不能使用80，443端口
	Port *int `required:"false"`
}
