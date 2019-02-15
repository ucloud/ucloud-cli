package udb

/*
UDBSlaveInstanceSet - DescribeUDBSlaveInstance

this model is auto created by ucloud code generater for open api,
you can also see https://docs.ucloud.cn for detail.
*/
type UDBSlaveInstanceSet struct {

	// DB实例id
	DBId string

	// 实例名称，至少6位
	Name string

	// DB类型id，mysql/mongodb按版本细分各有一个id 目前id的取值范围为[1,7],数值对应的版本如下： 1：mysql-5.5，2：mysql-5.1，3：percona-5.5 4：mongodb-2.4，5：mongodb-2.6，6：mysql-5.6， 7：percona-5.6
	DBTypeId string

	// DB实例使用的配置参数组id
	ParamGroupId int

	// 管理员帐户名，默认root
	AdminUser string

	// DB实例虚ip
	VirtualIP string

	// DB实例虚ip的mac地址
	VirtualIPMac string

	// 端口号，mysql默认3306，mongodb默认27017
	Port int

	// 对mysql的slave而言是master的DBId，对master则为空， 对mongodb则是副本集id
	SrcDBId string

	// 备份策略，不可修改，备份文件保留的数量，默认7次
	BackupCount int

	// 备份策略，不可修改，开始时间，单位小时计，默认3点
	BackupBeginTime int

	// 备份策略，一天内备份时间间隔，单位小时，默认24小时
	BackupDuration int

	// 备份策略，备份黑名单，mongodb则不适用
	BackupBlacklist string

	// DB状态标记 Init：初始化中，Fail：安装失败，Starting：启动中，Running：运行，Shutdown：关闭中，Shutoff：已关闭，Delete：已删除，Upgrading：升级中，Promoting：提升为独库进行中，Recovering：恢复中，Recover fail：恢复失败
	State string

	// DB实例创建时间，采用UTC计时时间戳
	CreateTime int

	// DB实例修改时间，采用UTC计时时间戳
	ModifyTime int

	// DB实例过期时间，采用UTC计时时间戳
	ExpiredTime int

	// Year， Month， Dynamic，Trial，默认: Dynamic
	ChargeType string

	// 内存限制(MB)，默认根据配置机型
	MemoryLimit int

	// 磁盘空间(GB), 默认根据配置机型
	DiskSpace int

	// 是否使用SSD
	UseSSD bool

	// SSD类型，SATA/PCI-E
	SSDType string

	// DB实例角色，mysql区分master/slave，mongodb多种角色
	Role string

	// DB实例磁盘已使用空间，单位GB
	DiskUsedSize int

	// DB实例数据文件大小，单位GB
	DataFileSize int

	// DB实例系统文件大小，单位GB
	SystemFileSize int

	// DB实例日志文件大小，单位GB
	LogFileSize int

	// 备份日期标记位。共7位,每一位为一周中一天的备份情况 0表示关闭当天备份,1表示打开当天备份。最右边的一位 为星期天的备份开关，其余从右到左依次为星期一到星期 六的备份配置开关，每周必须至少设置两天备份。 例如：1100000 表示打开星期六和星期五的自动备份功能
	BackupDate string

	// UDB实例模式类型, 可选值如下: "Normal": 普通版UDB实例;"HA": 高可用版UDB实例
	InstanceMode string
}
