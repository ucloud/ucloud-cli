package pathx

/*
GlobalSSHArea - GlobalSSH覆盖地区,包括关联的UCloud机房信息

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type GlobalSSHArea struct {

	// GlobalSSH覆盖的地区,如香港、东京、洛杉矶等
	Area string

	// 地区代号,以地区AirPort Code
	AreaCode string

	// ucloud机房代号构成的数组，如["hk","us-ca"]
	RegionSet []string
}
