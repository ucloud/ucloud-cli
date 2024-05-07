// Code is generated by ucloud-model, DO NOT EDIT IT.

package udb

/*
UDBBackupSet - DescribeUDBBackup
*/
type UDBBackupSet struct {

	// 备份完成时间(Unix时间戳)
	BackupEndTime int

	// 备份id
	BackupId int

	// 备份名称
	BackupName string

	// 备份文件大小(字节)
	BackupSize int

	// 备份时间(Unix时间戳)
	BackupTime int

	// 备份类型,取值为0或1,0表示自动，1表示手动
	BackupType int

	// 跨机房高可用备库所在可用区
	BackupZone string

	// dbid
	DBId string

	// 对应的db名称
	DBName string

	// 备份错误信息
	ErrorInfo string

	// 备份文件的MD5值，备份完成后显示，备份中或备份失败时为空,目前只支持Mysql NVMe机型与Mongo
	MD5 string

	// 备份状态 Backuping // 备份中 Success // 备份成功 Failed // 备份失败 Expired // 备份过期
	State string

	// 备份所在可用区
	Zone string
}

/*
UDBSlaveInstanceSet - DescribeUDBSlaveInstance
*/
type UDBSlaveInstanceSet struct {

	// 管理员帐户名，默认root
	AdminUser string

	// 备份策略，不可修改，开始时间，单位小时计，默认3点
	BackupBeginTime int

	// 备份策略，备份黑名单，mongodb则不适用
	BackupBlacklist string

	// 备份策略，不可修改，备份文件保留的数量，默认7次
	BackupCount int

	// 备份日期标记位。共7位,每一位为一周中一天的备份情况 0表示关闭当天备份,1表示打开当天备份。最右边的一位 为星期天的备份开关，其余从右到左依次为星期一到星期 六的备份配置开关，每周必须至少设置两天备份。 例如：1100000 表示打开星期六和星期五的自动备份功能
	BackupDate string

	// 备份策略，一天内备份时间间隔，单位小时，默认24小时
	BackupDuration int

	// 0 区分大小写, 1不区分, 只针对mysql8.0
	CaseSensitivityParam int

	// Year， Month， Dynamic，Trial，默认: Dynamic
	ChargeType string

	// 当DB类型为mongodb时，返回该实例所在集群中的角色，包括：mongos、configsrv_sccc、configsrv_csrs、shardsrv_datanode、shardsrv_arbiter，其中congfigsrv分为sccc和csrs两种模式，shardsrv分为datanode和arbiter两种模式
	ClusterRole string

	// DB实例创建时间，采用UTC计时时间戳
	CreateTime int

	// DB实例id
	DBId string

	// DB类型id，mysql/mongodb按版本细分各有一个id 目前id的取值范围为[1,7],数值对应的版本如下： 1：mysql-5.5，2：mysql-5.1，3：percona-5.5 4：mongodb-2.4，5：mongodb-2.6，6：mysql-5.6， 7：percona-5.6
	DBTypeId string

	// DB实例数据文件大小，单位GB
	DataFileSize float64

	// 磁盘空间(GB), 默认根据配置机型
	DiskSpace int

	// DB实例磁盘已使用空间，单位GB
	DiskUsedSize float64

	// DB实例过期时间，采用UTC计时时间戳
	ExpiredTime int

	// UDB实例模式类型, 可选值如下: "Normal": 普通版UDB实例;"HA": 高可用版UDB实例
	InstanceMode string

	// UDB数据库机型
	InstanceType string

	// UDB数据库机型ID
	InstanceTypeId int

	// DB实例日志文件大小，单位GB
	LogFileSize float64

	// 规格类型ID,当SpecificationType为1时有效
	MachineType string

	// 内存限制(MB)，默认根据配置机型
	MemoryLimit int

	// DB实例修改时间，采用UTC计时时间戳
	ModifyTime int

	// 实例名称，至少6位
	Name string

	// DB实例使用的配置参数组id
	ParamGroupId int

	// 端口号，mysql默认3306，mongodb默认27017
	Port int

	// 延时从库时长
	ReplicationDelaySeconds int

	// DB实例角色，mysql区分master/slave，mongodb多种角色
	Role string

	// SSD类型，SATA/PCI-E
	SSDType string

	// 实例计算规格类型，0或不传代表使用内存方式购买，1代表使用内存-cpu可选配比方式购买，需要填写MachineType
	SpecificationType string

	// 对mysql的slave而言是master的DBId，对master则为空， 对mongodb则是副本集id
	SrcDBId string

	// DB状态标记 Init：初始化中，Fail：安装失败，Starting：启动中，Running：运行，Shutdown：关闭中，Shutoff：已关闭，Delete：已删除，Upgrading：升级中，Promoting：提升为独库进行中，Recovering：恢复中，Recover fail：恢复失败,Remakeing:重做中,RemakeFail:重做失败, MajorVersionUpgrading:小版本升级中，MajorVersionUpgradeWaitForSwitch:高可用等待切换，MajorVersionUpgradeFail
	State string

	// 子网ID
	SubnetId string

	// DB实例系统文件大小，单位GB
	SystemFileSize float64

	// 获取资源其他信息
	Tag string

	// 是否使用SSD
	UseSSD bool

	// VPC的ID
	VPCId string

	// DB实例虚ip
	VirtualIP string

	// DB实例虚ip的mac地址
	VirtualIPMac string

	// 可用区
	Zone string
}

/*
UFileDataSet - 增加ufile的描述
*/
type UFileDataSet struct {

	// bucket名称
	Bucket string

	// Ufile的令牌tokenid
	TokenID string
}

/*
UDBInstanceSet - DescribeUDBInstance
*/
type UDBInstanceSet struct {

	// 管理员帐户名，默认root
	AdminUser string

	// 备份策略，不可修改，开始时间，单位小时计，默认3点
	BackupBeginTime int

	// 备份策略，备份黑名单，mongodb则不适用
	BackupBlacklist string

	// 备份策略，不可修改，备份文件保留的数量，默认7次
	BackupCount int

	// 备份日期标记位。共7位,每一位为一周中一天的备份情况 0表示关闭当天备份,1表示打开当天备份。最右边的一位 为星期天的备份开关，其余从右到左依次为星期一到星期 六的备份配置开关，每周必须至少设置两天备份。 例如：1100000 表示打开星期六和星期五的自动备份功能
	BackupDate string

	// 备份策略，一天内备份时间间隔，单位小时，默认24小时
	BackupDuration int

	// 默认的备份方式，nobackup表示不备份， snapshot 表示使用快照备份，logic 表示使用逻辑备份，xtrabackup表示使用物理备份。
	BackupMethod string

	// 跨可用区高可用备库所在可用区
	BackupZone string

	// CPU核数
	CPU int

	// 0区分大小写, 1不分区
	CaseSensitivityParam int

	// Year， Month， Dynamic，Trial，默认: Dynamic
	ChargeType string

	//
	CluserRole string `deprecated:"true"`

	// 当DB类型为mongodb时，返回该实例所在集群中的角色，包括：mongos、configsrv_sccc、configsrv_csrs、shardsrv_datanode、shardsrv_arbiter，其中congfigsrv分为sccc和csrs两种模式，shardsrv分为datanode和arbiter两种模式
	ClusterRole string

	// DB实例创建时间，采用UTC计时时间戳
	CreateTime int

	// DB实例id
	DBId string

	// mysql实例提供具体小版本信息
	DBSubVersion string

	// DB类型id，mysql/mongodb按版本细分各有一个id 目前id的取值范围为[1,7],数值对应的版本如下： 1：mysql-5.5，2：mysql-5.1，3：percona-5.5 4：mongodb-2.4，5：mongodb-2.6，6：mysql-5.6， 7：percona-5.6
	DBTypeId string

	// DB实例数据文件大小，单位GB
	DataFileSize float64

	// 如果在需要返回从库的场景下，返回该DB实例的所有从库DB实例信息列表。列表中每一个元素的内容同UDBSlaveInstanceSet 。如果这个DB实例没有从库的情况下，此时返回一个空的列表
	DataSet []UDBSlaveInstanceSet

	// 磁盘空间(GB), 默认根据配置机型
	DiskSpace int

	// DB实例磁盘已使用空间，单位GB
	DiskUsedSize float64

	// mysql是否开启了SSL；1->未开启  2->开启
	EnableSSL int

	// DB实例过期时间，采用UTC计时时间戳
	ExpiredTime int

	// 该实例的ipv6地址
	IPv6Address string

	// UDB实例模式类型, 可选值如下: “Normal”： 普通版UDB实例 “HA”: 高可用版UDB实例
	InstanceMode string

	// UDB数据库机型
	InstanceType string

	// UDB数据库机型ID (已弃用)
	InstanceTypeId int

	// DB实例日志文件大小，单位GB
	LogFileSize float64

	// 数据库机型规格
	MachineType string

	// 内存限制(MB)，默认根据配置机型
	MemoryLimit int

	// DB实例修改时间，采用UTC计时时间戳
	ModifyTime int

	// 实例名称，至少6位
	Name string

	// DB实例使用的配置参数组id
	ParamGroupId int

	// 端口号，mysql默认3306，mongodb默认27017
	Port int

	// DB实例角色，mysql区分master/slave，mongodb多种角色
	Role string

	// SSD类型，SATA/PCI-E/NVMe
	SSDType string

	// SSL到期时间
	SSLExpirationTime int

	// 是否使用可选cpu类型规格
	SpecificationType int

	// 对mysql的slave而言是master的DBId，对master则为空， 对mongodb则是副本集id
	SrcDBId string

	// DB状态标记 Init：初始化中，Fail：安装失败，Starting：启动中，Running：运行，Shutdown：关闭中，Shutoff：已关闭，Delete：已删除，Upgrading：升级中，Promoting：提升为独库进行中，Recovering：恢复中，Recover fail：恢复失败, Remakeing:重做中,RemakeFail:重做失败，VersionUpgrading:小版本升级中，VersionUpgradeWaitForSwitch:高可用等待切换，VersionUpgradeFail：小版本升级失败，UpdatingSSL：修改SSL中，UpdateSSLFail：修改SSL失败,MajorVersionUpgrading:小版本升级中，MajorVersionUpgradeWaitForSwitch:高可用等待切换，MajorVersionUpgradeFail
	State string

	// 子网ID
	SubnetId string

	// DB实例系统文件大小，单位GB
	SystemFileSize float64

	// 获取资源其他信息
	Tag string

	// 是否使用SSD
	UseSSD bool

	// 用户转存备份到自己的UFILE配置, 结构参考UFileDataSet
	UserUFileData UFileDataSet

	// VPC的ID
	VPCId string

	// DB实例虚ip
	VirtualIP string

	// DB实例虚ip的mac地址
	VirtualIPMac string

	// DB实例所在可用区
	Zone string
}

/*
UDBInstanceBinlogSet - DescribeUDBInstanceBinlog
*/
type UDBInstanceBinlogSet struct {

	// Binlog文件生成时间(时间戳)
	BeginTime int

	// Binlog文件结束时间(时间戳)
	EndTime int

	// Binlog文件名
	Name string

	// Binlog文件大小
	Size int
}

/*
UDBInstancePriceSet - DescribeUDBInstancePrice
*/
type UDBInstancePriceSet struct {

	// Year， Month， Dynamic，Trial
	ChargeType string

	// 价格，单位为分
	Price int
}

/*
LogPackageDataSet - DescribeUDBLogPackage
*/
type LogPackageDataSet struct {

	// 备份id
	BackupId int

	// 备份名称
	BackupName string

	// 备份文件大小
	BackupSize int

	// 备份时间
	BackupTime int

	// 备份类型，包括2-binlog备份，3-slowlog备份
	BackupType int

	// 跨可用区高可用备库所在可用区
	BackupZone string

	// binlog备份类型 Manual //手动备份 Auto //自动备份
	BinlogType string

	// dbid
	DBId string

	// 对应的db名称
	DBName string

	// 备份状态 Backuping // 备份中 Success // 备份成功 Failed // 备份失败 Expired // 备份过期
	State string

	// 所在可用区
	Zone string
}

/*
UDBParamMemberSet - DescribeUDBParamGroup
*/
type UDBParamMemberSet struct {

	// 允许的值(根据参数类型，用分隔符表示)
	AllowedVal string

	// 参数值应用类型,取值范围为{0,10,20}，各值代表 意义为0-unknown、10-static、20-dynamic
	ApplyType int

	// 允许值的格式类型，取值范围为{0,10,20}，意义分 别为PVFT_UNKOWN=0,PVFT_RANGE=10, PVFT_ENUM=20
	FormatType int

	// 参数名称
	Key string

	// 是否可更改，默认为false
	Modifiable bool

	// 参数值
	Value string

	// 参数值应用类型，取值范围为{0,10,20,30},各值 代表意义为 0-unknown、10-int、20-string、 30-bool
	ValueType int
}

/*
UDBParamGroupSet - DescribeUDBParamGroup
*/
type UDBParamGroupSet struct {

	// DB类型id，mysql/mongodb按版本细分各有一个id 目前id的取值范围为[1,7],数值对应的版本如下 1：mysql-5.5，2：mysql-5.1，3：percona-5.5 4：mongodb-2.4，5：mongodb-2.6，6：mysql-5.6 7：percona-5.6
	DBTypeId string

	// 参数组描述
	Description string

	// 参数组id
	GroupId int

	// 参数组名称
	GroupName string

	// 参数组是否可修改
	Modifiable bool

	// 参数的键值对表 UDBParamMemberSet
	ParamMember []UDBParamMemberSet

	//
	RegionFlag bool

	//
	Zone string
}

/*
UDBRWSplittingSet - 读写分离
*/
type UDBRWSplittingSet struct {

	// DB实例ID
	DBId string

	// 读写分离比重
	ReadWeight int

	// 主库/从库
	Role string

	// DB状态
	State string

	// DBIP
	VirtualIP string
}

/*
UDBTypeSet - DescribeUDBType
*/
type UDBTypeSet struct {

	// mysql子版本，如mysql-8.0.25,mysql-8.0.16
	DBSubVersion string

	// DB类型id，mysql/mongodb按版本细分各有一个id, 目前id的取值范围为[1,7],数值对应的版本如下： 1：mysql-5.5，2：mysql-5.1，3：percona-5.5 4：mongodb-2.4，5：mongodb-2.6，6：mysql-5.6， 7：percona-5.6
	DBTypeId string
}

/*
ConnNumMap - db实例ip和连接数信息
*/
type ConnNumMap struct {

	// 客户端IP
	Ip string

	// 该Ip连接数
	Num int
}

/*
TableData - 用户表详情
*/
type TableData struct {

	// 表所属的库名称
	DBName string

	// 表的引擎（innodb, myisam）
	Engine string

	// 表名称
	TableName string
}

/*
UDBDatabaseData - 某个库的详细信息
*/
type UDBDatabaseData struct {

	// 数据库名称
	DBName string

	// 该库所有的表集合
	TableDataSet []TableData
}
