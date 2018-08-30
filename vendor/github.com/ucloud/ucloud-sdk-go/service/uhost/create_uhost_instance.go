package uhost

import (
	"encoding/base64"

	"github.com/ucloud/ucloud-sdk-go/sdk/request"
	"github.com/ucloud/ucloud-sdk-go/sdk/response"
)

type CreateUHostInstanceRequest struct {
	request.CommonBase

	// 地域。 参见 [地域和可用区列表](../summary/regionlist.html)
	Region string

	// 可用区。参见 [可用区列表](../summary/regionlist.html)
	Zone string

	// 项目ID。不填写为默认项目，子帐号必须填写。 请参考[GetProjectList接口](../summary/get_project_list.html)
	ProjectId string

	// 镜像ID。 请通过 [DescribeImage](describe_image.html)获取
	ImageId string

	// UHost密码，LoginMode为Password时此项必须（密码需使用base64进行编码）
	Password string

	// UHost实例名称。默认：UHost
	Name string

	// 业务组。默认：Default（Default即为未分组）
	Tag string

	// 计费模式。枚举值为： Year，按年付费； Month，按月付费； Dynamic，按小时付费（需开启权限）。默认为月付
	ChargeType string

	// 购买时长。默认: 1。按小时购买(Dynamic)时无需此参数。 月付时，此参数传0，代表了购买至月末。
	Quantity int

	// 云主机机型。枚举值：N1：系列1标准型；N2：系列2标准型；I1: 系列1高IO型；I2，系列2高IO型； D1: 系列1大数据机型；G1: 系列1GPU型；G2：系列2GPU型；北京A、北京C、上海二A、香港A可用区默认N1，其他机房默认N2。不同机房的主机类型支持情况不同。详情请参考控制台。
	UHostType string

	// 虚拟CPU核数。 单位：个。可选参数：{1,2,4,8,12,16,24,32}。默认值: 4
	CPU int

	// 内存大小。单位：MB。范围 ：[1024, 131072]， 取值为2的幂次方。默认值：8192。
	Memory int

	// GPU卡核心数。仅GPU机型支持此字段；系列1可选1,2；系列2可选1,2,3,4。GPU可选数量与CPU有关联，详情请参考控制台。
	GPU int

	// 主机登陆模式。密码（默认选项）: Password，key: KeyPair（此项暂不支持）
	LoginMode string

	// 磁盘类型，同时设定系统盘和数据盘的磁盘类型。枚举值为：LocalDisk，本地磁盘; UDisk，云硬盘；默认为LocalDisk。仅部分可用区支持云硬盘方式的主机存储方式，具体请查询控制台。
	StorageType string

	// 系统盘大小。 单位：GB， 范围[20,100]， 步长：10
	BootDiskSpace int

	// 数据盘大小。 单位：GB， 范围[0,8000]， 步长：10， 默认值：20，云盘支持0-8000；本地普通盘支持0-2000；本地SSD盘（包括所有GPU机型）支持100-1000
	DiskSpace int

	// 网络增强。目前仅Normal（不开启） 和Super（开启）可用。默认Normal。 不同机房的网络增强支持情况不同。详情请参考控制台。
	NetCapability string

	// 是否开启方舟特性。Yes为开启方舟，No为关闭方舟。Basic为免费基础快照模式（暂不支持）。
	TimemachineFeature string

	// 是否开启热升级特性。True为开启，False为未开启，默认False。仅系列1云主机需要使用此字段，系列2云主机根据镜像是否支持云主机。
	HotplugFeature bool

	// 加密盘的密码。若输入此字段，自动选择加密盘。加密盘需要权限位。
	DiskPassword string

	// 网络ID（VPC2.0情况下无需填写）。VPC1.0情况下，若不填写，代表选择基础网络； 若填写，代表选择子网。参见DescribeSubnet。
	NetworkId string

	// VPC ID。VPC2.0下需要填写此字段。
	VPCId string

	// 子网ID。VPC2.0下需要填写此字段。
	SubnetId string

	// 【数组】创建云主机时指定内网IP。当前只支持一个内网IP。调用方式举例：PrivateIp.0=x.x.x.x。
	PrivateIp []string

	// 防火墙Id，默认：Web推荐防火墙。如何查询SecurityGroupId请参见 [DescribeSecurityGroup](../unet-api/describe_security_group.html)
	SecurityGroupId string

	// 代金券ID。请通过DescribeCoupon接口查询，或登录用户中心查看
	CouponId string
}

// NewCreateUHostInstanceRequest will create request of CreateUHostInstance action.
func (c *UHostClient) NewCreateUHostInstanceRequest() *CreateUHostInstanceRequest {
	cfg := c.client.GetConfig()

	return &CreateUHostInstanceRequest{
		CommonBase: request.CommonBase{
			Region:    cfg.Region,
			ProjectId: cfg.ProjectId,
		},
	}
}

type CreateUHostInstanceResponse struct {
	response.CommonBase

	// Message, if an error was occupy, return the error message
	Message string

	// 操作返回码
	RetCode int

	// 操作名称
	Action string

	// UHost实例Id集合
	UHostIds []string

	// IP信息
	IPs []string
}

// CreateUHostInstance - 指定数据中心，根据资源使用量创建指定数量的UHost实例。
func (c *UHostClient) CreateUHostInstance(req *CreateUHostInstanceRequest) (*CreateUHostInstanceResponse, error) {
	var err error
	var res CreateUHostInstanceResponse
	req.Password = base64.StdEncoding.EncodeToString([]byte(req.Password))

	err = c.client.InvokeAction("CreateUHostInstance", req, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
