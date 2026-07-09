package uhost

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/spf13/cobra"

	uhostsdk "github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/services/unet"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/internal/common"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

// _MaxBoundSecGroupCount caps the --security-group-id count. Verbatim from
// cmd/uhost.go.
const _MaxBoundSecGroupCount = 5

// newCreate ucloud uhost create
func newCreate(ctx *cli.Context) *cobra.Command {
	var bindEipIDs []string
	var hotPlug string
	var async bool
	var count int
	var concurrent int
	var hotPlugImageFlag bool
	var userData string
	var userDataImageFlag bool
	var userDataBase64 string
	var firewallId string
	var secGroupIds []string
	var keyPairId string
	var password string

	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	unetClient := cli.NewServiceClient(ctx, unet.NewClient)
	req := client.NewCreateUHostInstanceRequest()
	eipReq := uhostsdk.CreateUHostInstanceParamNetworkInterfaceEIP{}
	updateEIPReq := unetClient.NewUpdateEIPAttributeRequest()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create UHost instance",
		Long:  "Create UHost instance",
		// SilenceUsage: runtime failures (RunE returning an error below) must not
		// dump the full flag usage — aws/gcloud print the error only. Flag/arg
		// mistakes still print their own message via cobra.
		SilenceUsage: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(userData) > 0 && len(userDataBase64) > 0 {
				return fmt.Errorf("%q conflicts with %q, can only set one of both", "user-data", "user-data-base64")
			}

			if len(userDataBase64) > 0 {
				if !common.IsBase64Encoded([]byte(userDataBase64)) {
					return fmt.Errorf("%q must be base64-encoded", "user-data-base64")
				}
			}

			if concurrent > 50 {
				return fmt.Errorf("%q should not be more than 50, current value is %v", "concurrent", concurrent)
			}

			// GPU and GpuType must be specified together: if one is set, the other must also be set.
			gpuVal, _ := cmd.Flags().GetInt("gpu")
			gpuTypeVal, _ := cmd.Flags().GetString("gpu-type")
			if gpuVal > 0 && gpuTypeVal == "" {
				return fmt.Errorf("--gpu requires --gpu-type, e.g. --gpu-type V100")
			}
			if gpuTypeVal != "" && gpuVal <= 0 {
				return fmt.Errorf("--gpu-type requires --gpu, e.g. --gpu 1")
			}

			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			*req.Memory *= 1024
			// If --gpu and --gpu-type are specified, auto-set MachineType to "G"
			// unless the user explicitly set --machine-type.
			if *req.GPU > 0 && *req.GpuType != "" {
				if !cmd.Flags().Changed("machine-type") {
					req.MachineType = sdk.String("G")
				}
			}
			if len(password) > 0 {
				req.LoginMode = sdk.String("Password")
				req.KeyPairId = nil
				req.Password = sdk.String(password)
			} else if len(keyPairId) > 0 {
				req.LoginMode = sdk.String("KeyPair")
				req.KeyPairId = sdk.String(keyPairId)
				req.Password = nil
			} else {
				return fmt.Errorf("password or key-pair-id is required")
			}
			if len(firewallId) > 0 {
				req.SecurityGroupId = sdk.String(firewallId)
			} else if len(secGroupIds) > 0 {
				if len(secGroupIds) > _MaxBoundSecGroupCount {
					return fmt.Errorf("security group count should not be more than 5")
				}
				secGroupList := make([]uhostsdk.CreateUHostInstanceParamSecGroupId, 0)
				for idx, secGroupId := range secGroupIds {
					secGroupList = append(secGroupList, uhostsdk.CreateUHostInstanceParamSecGroupId{Id: sdk.String(secGroupId), Priority: sdk.Int(1 + idx)})
				}
				req.SecGroupId = secGroupList
				req.SecurityMode = sdk.String("SecGroup")
			}
			req.ImageId = sdk.String(ctx.PickResourceID(*req.ImageId))
			req.VPCId = sdk.String(ctx.PickResourceID(*req.VPCId))
			req.SubnetId = sdk.String(ctx.PickResourceID(*req.SubnetId))
			req.IsolationGroup = sdk.String(ctx.PickResourceID(*req.IsolationGroup))
			if *req.Disks[1].Type == "NONE" || *req.Disks[1].Type == "" {
				req.Disks = req.Disks[:1]
			}
			if hotPlug == "true" || len(userData) > 0 || len(userDataBase64) > 0 {
				any, err := describeImageByID(ctx, *req.ProjectId, *req.Region, *req.Zone)(ctx.PickResourceID(*req.ImageId), nil)
				if err != nil {
					return fmt.Errorf("check image support feaures failed: %v", err)
				} else {
					image, ok := any.(*uhostsdk.UHostImageSet)
					if !ok {
						return fmt.Errorf("check image support feaures failed, image %s may not exist", *req.ImageId)
					}
					for _, feature := range image.Features {
						if feature == "HotPlug" {
							hotPlugImageFlag = true
						}
						if feature == "CloudInit" {
							userDataImageFlag = true
						}
					}
				}
				if !hotPlugImageFlag && hotPlug == "true" {
					ctx.LogWarn(fmt.Sprintf("warning. image %s does not support hot-plug", *req.ImageId))
					req.HotplugFeature = sdk.Bool(false)
				}

				if !userDataImageFlag && (len(userData) > 0 || len(userDataBase64) > 0) {
					return fmt.Errorf("image %s does not support user-data feature", *req.ImageId)
				}

				if hotPlug == "true" {
					req.HotplugFeature = sdk.Bool(true)
				}

				if len(userData) > 0 {
					req.UserData = sdk.String(base64.StdEncoding.EncodeToString([]byte(userData)))
				}

				if len(userDataBase64) > 0 {
					req.UserData = sdk.String(userDataBase64)
				}
			}
			if *eipReq.Bandwidth != 0 || *eipReq.PayMode == "ShareBandwidth" {
				if *eipReq.OperatorName == "" {
					*eipReq.OperatorName = getEIPLine(*req.Region)
				}
				req.NetworkInterface = []uhostsdk.CreateUHostInstanceParamNetworkInterface{{EIP: &eipReq}}
			}

			prog := ctx.NewProgress()
			wg := &sync.WaitGroup{}
			tokens := make(chan struct{}, concurrent)
			fc := &failCounter{}
			rc := &resultCollector{}
			wg.Add(count)
			batchRename, err := regexp.Match(`\[\d+,\d+\]`, []byte(*req.Name))
			if err != nil || !batchRename {
				batchRename = false
			}
			if batchRename {
				var actualRequest uhostsdk.CreateUHostInstanceRequest
				actualRequest = *req
				if len(bindEipIDs) > 0 {
					if len(bindEipIDs) != count {
						return fmt.Errorf("bind-eip count should be equal to uhost count")
					}
					actualRequest.NetworkInterface = nil
				}
				wg.Add(1 - count)
				createMultipleUhostWrapper(ctx, prog, client, unetClient, &actualRequest, count, updateEIPReq, bindEipIDs, async, make(chan bool, 1), wg, tokens, fc, rc)

			} else if count <= 5 {
				for i := 0; i < count; i++ {
					bindEipID := ""
					if len(bindEipIDs) > i {
						bindEipID = bindEipIDs[i]
					}
					var actualRequest uhostsdk.CreateUHostInstanceRequest
					actualRequest = *req
					if bindEipID != "" {
						actualRequest.NetworkInterface = nil
					}
					createUhostWrapper(ctx, prog, client, unetClient, &actualRequest, updateEIPReq, bindEipID, async, make(chan bool, count), wg, tokens, i, fc, rc)
				}
			} else {
				retCh := make(chan bool, count)
				prog.Disable()

				go func(req uhostsdk.CreateUHostInstanceRequest) {
					for i := 0; i < count; i++ {
						actualRequest := req
						bindEipID := ""
						if len(bindEipIDs) > i {
							bindEipID = bindEipIDs[i]
							actualRequest.NetworkInterface = nil
						}
						go createUhostWrapper(ctx, prog, client, unetClient, &actualRequest, updateEIPReq, bindEipID, async, retCh, wg, tokens, i, fc, rc)
					}
				}(*req)

				go func() {
					var success, fail int
					prog.Refresh(fmt.Sprintf("uhost creating, total:%d, success:%d, fail:%d", count, success, fail))
					for ret := range retCh {
						if ret {
							success++
						} else {
							fail++
						}
						prog.Refresh(fmt.Sprintf("uhost creating, total:%d, success:%d, fail:%d", count, success, fail))
						if count == success+fail && fail > 0 {
							fmt.Fprintf(ctx.ProgressWriter(), "Check logs in %s\n", ctx.LogFilePath())
						}
					}
				}()
			}
			wg.Wait()
			ctx.EmitResult(rc.all()...)
			if n := fc.count(); n > 0 {
				return fmt.Errorf("%d of %d uhost create operation(s) failed; see the error(s) above or logs in %s", n, count, ctx.LogFilePath())
			}
			return nil
		},
	}

	req.Disks = make([]uhostsdk.UHostDisk, 2)
	req.Disks[0].IsBoot = sdk.String("True")
	req.Disks[1].IsBoot = sdk.String("False")

	flags := cmd.Flags()
	flags.SortFlags = false
	req.CPU = flags.Int("cpu", 4, "Required. The count of CPU cores. Optional parameters: {1, 2, 4, 8, 12, 16, 24, 32, 64}")
	req.Memory = flags.Int("memory-gb", 8, "Required. Memory size. Unit: GB. Range: [1, 512], multiple of 2")
	flags.StringVar(&password, "password", "", "Optional. Password of the uhost user(root/ubuntu)")
	flags.StringVar(&keyPairId, "key-pair-id", "", "Optional. Resource ID of ssh key pair. See 'ucloud api --Action DescribeUHostKeyPairs' Where both password and key-pair-id are set, the key-pair-id is ignored")
	req.ImageId = flags.String("image-id", "", "Required. The ID of image. see 'ucloud image list'")
	flags.BoolVar(&async, "async", false, "Optional. Do not wait for the long-running operation to finish.")
	flags.IntVar(&count, "count", 1, "Optional. Number of uhost to create.")
	flags.IntVar(&concurrent, "concurrent", 20, "Optional. The count of concurrent uhost creation requests.")
	req.VPCId = flags.String("vpc-id", "", "Optional. VPC ID. This field is required under VPC2.0. See 'ucloud vpc list'")
	req.SubnetId = flags.String("subnet-id", "", "Optional. Subnet ID. This field is required under VPC2.0. See 'ucloud subnet list'")
	req.Name = flags.String("name", "UHost", "Optional. UHost instance name")
	flags.StringSliceVar(&bindEipIDs, "bind-eip", nil, "Optional. Resource ID or IP Address of eip that will be bound to the new created uhost")
	eipReq.OperatorName = flags.String("create-eip-line", "", "Optional. BGP for regions in the chinese mainland and International for overseas regions")
	eipReq.Bandwidth = flags.Int("create-eip-bandwidth-mb", 0, "Optional. Required if you want to create new EIP. Bandwidth(Unit:Mbps).The range of value related to network charge mode. By traffic [1, 300]; by bandwidth [1,800] (Unit: Mbps); it could be 0 if the eip belong to the shared bandwidth")
	eipReq.PayMode = flags.String("create-eip-traffic-mode", "Bandwidth", "Optional. 'Traffic','Bandwidth' or 'ShareBandwidth'")
	eipReq.ShareBandwidthId = flags.String("shared-bw-id", "", "Optional. Resource ID of shared bandwidth. It takes effect when create-eip-traffic-mode is ShareBandwidth ")
	updateEIPReq.Name = flags.String("create-eip-name", "", "Optional. Name of created eip to bind with the uhost")
	updateEIPReq.Remark = flags.String("create-eip-remark", "", "Optional.Remark of your EIP.")

	req.ChargeType = flags.String("charge-type", "Month", "Optional.'Year',pay yearly;'Month',pay monthly;'Dynamic', pay hourly")
	req.Quantity = flags.Int("quantity", 1, "Optional. The duration of the instance. N years/months.")
	// bindProjectID/bindRegion/bindZone (cmd/uhost.go) → ctx.Bind*: these register
	// the dynamic project/region/zone completion the golden requires (raw flags
	// would drop it) and share the value with req via SetRef.
	ctx.BindProjectID(cmd, req)
	ctx.BindRegion(cmd, req)
	ctx.BindZone(cmd, req)

	req.MachineType = flags.String("machine-type", "O", "Optional. Accept values: N, C, G, O, OS. Forward to https://docs.ucloud.cn/api/uhost-api/uhost_type for details")
	req.MinimalCpuPlatform = flags.String("minimal-cpu-platform", "", "Optional. Accept values: Intel/Auto, Intel/IvyBridge, Intel/Haswell, Intel/Broadwell, Intel/Skylake, Intel/Cascadelake")
	req.UHostType = flags.String("type", "", "Optional. Accept values: N1, N2, N3, G1, G2, G3, I1, I2, C1. Forward to https://docs.ucloud.cn/api/uhost-api/uhost_type for details")
	req.GPU = flags.Int("gpu", 0, "Optional. The count of GPU cores.")
	req.NetCapability = flags.String("net-capability", "Normal", "Optional. Accept values: Normal, Super and Ultra. 'Normal' will disable network enhancement. 'Super' will enable network enhancement 1.0. 'Ultra' will enable network enhancement 2.0")
	flags.StringVar(&hotPlug, "hot-plug", "true", "Optional. Enable hot plug feature or not. Accept values: true or false")
	req.Disks[0].Type = flags.String("os-disk-type", "CLOUD_SSD", "Optional. Enumeration value. 'LOCAL_NORMAL', Ordinary local disk; 'CLOUD_NORMAL', Ordinary cloud disk; 'LOCAL_SSD',local ssd disk; 'CLOUD_SSD',cloud ssd disk; 'EXCLUSIVE_LOCAL_DISK',big data. The disk only supports a limited combination.")
	req.Disks[0].Size = flags.Int("os-disk-size-gb", 20, "Optional. Default 20G. Windows should be bigger than 40G Unit GB")
	req.Disks[0].BackupType = flags.String("os-disk-backup-type", "NONE", "Optional. Enumeration value, 'NONE' or 'DATAARK'. DataArk supports real-time backup, which can restore the disk back to any moment within the last 12 hours. (Normal Local Disk and Normal Cloud Disk Only)")
	req.Disks[1].Type = flags.String("data-disk-type", "CLOUD_SSD", "Optional. Accept values: 'LOCAL_NORMAL','LOCAL_SSD','CLOUD_NORMAL',CLOUD_SSD','CLOUD_RSSD','EXCLUSIVE_LOCAL_DISK' and 'NONE'. 'LOCAL_NORMAL', Ordinary local disk; 'CLOUD_NORMAL', Ordinary cloud disk; 'LOCAL_SSD',local ssd disk; 'CLOUD_SSD',cloud ssd disk; 'CLOUD_RSSD', coud rssd disk; 'EXCLUSIVE_LOCAL_DISK',big data. The disk only supports a limited combination. 'NONE', create uhost without data disk. More details https://docs.ucloud.cn/api/uhost-api/disk_type")
	req.Disks[1].Size = flags.Int("data-disk-size-gb", 20, "Optional. Disk size. Unit GB")
	req.Disks[1].BackupType = flags.String("data-disk-backup-type", "NONE", "Optional. Enumeration value, 'NONE' or 'DATAARK'. DataArk supports real-time backup, which can restore the disk back to any moment within the last 12 hours. (Normal Local Disk and Normal Cloud Disk Only)")
	flags.StringVar(&firewallId, "firewall-id", "", "Optional. Firewall Id, default: Web recommended firewall. see 'ucloud firewall list'.")
	flags.StringSliceVar(&secGroupIds, "security-group-id", nil, "Optional. Security Group Id. Before using security group function, please confirm the account has such permission. When both firewall-id and security-group-id are set, the security-group-id will be ignored")
	req.Tag = flags.String("group", "Default", "Optional. Business group")
	req.IsolationGroup = flags.String("isolation-group", "", "Optional. Resource ID of isolation group. see 'ucloud uhost isolation-group list")
	req.GpuType = flags.String("gpu-type", "", "Optional. The type of GPU instance. Required if defined the `machine-type` as 'G'. Accept values: 'K80','P40','V100','T4','T4S','T4A','2080Ti','2080Ti-4C','1080Ti','V100S','MI100','2080','2080TiS','2080TiPro','3090','A100','A800','3080Ti','4090','4090Pro','4090_48G','4090LD','MR-V100','MetaX-C500','H800','H20','H100','H200','5090','5090D','5090Pro'. Forward to https://docs.ucloud.cn/api/uhost-api/uhost_type for details.")
	flags.StringVar(&userData, "user-data", "", "Optional. Conflicts with `user-data-base64`. ConCustomize the startup behaviors when launching the instance. Forward to https://docs.ucloud.cn/uhost/guide/metadata/userdata for details.")
	flags.StringVar(&userDataBase64, "user-data-base64", "", "Optional. Conflicts with `user-data`. Customize the startup behaviors when launching the instance. The value must be base64-encode. Forward to https://docs.ucloud.cn/uhost/guide/metadata/userdata for details.")

	flags.MarkDeprecated("type", "please use --machine-type instead")
	command.SetFlagValues(cmd, "charge-type", "Month", "Year", "Dynamic", "Trial")
	command.SetFlagValues(cmd, "hot-plug", "true", "false")
	command.SetFlagValues(cmd, "cpu", "1", "2", "4", "8", "12", "16", "24", "32", "64")
	command.SetFlagValues(cmd, "type", "N2", "N1", "N3", "I2", "I1", "C1", "G1", "G2", "G3")
	command.SetFlagValues(cmd, "machine-type", "N", "C", "G", "O", "OS")
	command.SetFlagValues(cmd, "minimal-cpu-platform", "Intel/Auto", "Intel/IvyBridge", "Intel/Haswell", "Intel/Broadwell", "Intel/Skylake", "Intel/Cascadelake")
	command.SetFlagValues(cmd, "net-capability", "Normal", "Super", "Ultra")
	command.SetFlagValues(cmd, "os-disk-type", "LOCAL_NORMAL", "CLOUD_NORMAL", "LOCAL_SSD", "CLOUD_SSD", "CLOUD_RSSD", "EXCLUSIVE_LOCAL_DISK")
	command.SetFlagValues(cmd, "os-disk-backup-type", "NONE", "DATAARK")
	command.SetFlagValues(cmd, "data-disk-type", "LOCAL_NORMAL", "CLOUD_NORMAL", "LOCAL_SSD", "CLOUD_SSD", "CLOUD_RSSD", "EXCLUSIVE_LOCAL_DISK", "NONE")
	command.SetFlagValues(cmd, "data-disk-backup-type", "NONE", "DATAARK")
	command.SetFlagValues(cmd, "create-eip-line", "BGP", "International")
	command.SetFlagValues(cmd, "create-eip-traffic-mode", "Bandwidth", "Traffic", "ShareBandwidth")
	command.SetFlagValues(cmd, "gpu-type", "K80", "P40", "V100", "T4", "T4S", "T4A", "2080Ti", "2080Ti-4C", "1080Ti", "V100S", "MI100", "2080", "2080TiS", "2080TiPro", "3090", "A100", "A800", "3080Ti", "4090", "4090Pro", "4090_48G", "4090LD", "MR-V100", "MetaX-C500", "H800", "H20", "H100", "H200", "5090", "5090D", "5090Pro")

	command.SetCompletion(cmd, "image-id", func() []string {
		return getImageList(ctx, []string{IMAGE_AVAILABLE}, IMAGE_BASE, *req.ProjectId, *req.Region, *req.Zone)
	})
	command.SetCompletion(cmd, "vpc-id", func() []string {
		return getAllVPCIdNames(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "bind-eip", func() []string {
		return getAllEip(ctx, *req.ProjectId, *req.Region, []string{EIP_FREE}, nil)
	})
	command.SetCompletion(cmd, "firewall-id", func() []string {
		return getFirewallIDNames(ctx, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "subnet-id", func() []string {
		return getAllSubnetIDNames(ctx, *req.VPCId, *req.ProjectId, *req.Region)
	})
	command.SetCompletion(cmd, "isolation-group", func() []string {
		return getIsolationGroupList(ctx, *req.ProjectId, *req.Region)
	})

	cmd.MarkFlagRequired("cpu")
	cmd.MarkFlagRequired("memory-gb")
	cmd.MarkFlagRequired("image-id")

	return cmd
}

// createMultipleUhostWrapper handles UI + concurrency control for the batch-rename
// path. Mirrors cmd/uhost.go createMultipleUhostWrapper.
func createMultipleUhostWrapper(ctx *cli.Context, prog *cli.Progress, client *uhostsdk.UHostClient, unetClient *unet.UNetClient, req *uhostsdk.CreateUHostInstanceRequest, count int, updateEIPReq *unet.UpdateEIPAttributeRequest, bindEipIDs []string, async bool, retCh chan<- bool, wg *sync.WaitGroup, tokens chan struct{}, fc *failCounter, rc *resultCollector) {
	//控制并发数量
	tokens <- struct{}{}
	defer func() {
		<-tokens
		//设置延时，使报错能渲染出来
		time.Sleep(time.Second / 5)
		wg.Done()
	}()

	success, logs := createMultipleUhost(ctx, prog, client, unetClient, req, count, updateEIPReq, bindEipIDs, async, rc)
	if !success {
		fc.inc()
	}
	retCh <- success
	logs = append(logs, fmt.Sprintf("result:%t", success))
	ctx.LogInfo(logs...)
}

// createUhostWrapper handles UI + concurrency control for one uhost. Mirrors
// cmd/uhost.go createUhostWrapper.
func createUhostWrapper(ctx *cli.Context, prog *cli.Progress, client *uhostsdk.UHostClient, unetClient *unet.UNetClient, req *uhostsdk.CreateUHostInstanceRequest, updateEIPReq *unet.UpdateEIPAttributeRequest, bindEipID string, async bool, retCh chan<- bool, wg *sync.WaitGroup, tokens chan struct{}, idx int, fc *failCounter, rc *resultCollector) {
	//控制并发数量
	tokens <- struct{}{}
	defer func() {
		<-tokens
		//设置延时，使报错能渲染出来
		time.Sleep(time.Second / 5)
		wg.Done()
	}()

	success, logs := createUhost(ctx, prog, client, unetClient, req, updateEIPReq, bindEipID, async, rc)
	if !success {
		fc.inc()
	}
	retCh <- success
	logs = append(logs, fmt.Sprintf("index:%d, result:%t", idx, success))
	ctx.LogInfo(logs...)
}

func createMultipleUhost(ctx *cli.Context, prog *cli.Progress, client *uhostsdk.UHostClient, unetClient *unet.UNetClient, req *uhostsdk.CreateUHostInstanceRequest, count int, updateEIPReq *unet.UpdateEIPAttributeRequest, bindEipIDs []string, async bool, rc *resultCollector) (bool, []string) {
	if req.MaxCount == nil {
		req.MaxCount = sdk.Int(1)
	}
	req.MaxCount = sdk.Int(count)

	resp, err := client.CreateUHostInstance(req)
	block := prog.NewBlock()
	logs := []string{"=================================================="}
	if err != nil {
		logs = append(logs, fmt.Sprintf("err:%v", err))
		reportFail(ctx, prog, block, cli.ParseError(err))
		return false, logs
	}
	if len(bindEipIDs) > 0 && len(bindEipIDs) != count {
		reportFail(ctx, prog, block, fmt.Sprintf("expect eip count %d, accept %d", count, len(bindEipIDs)))
		return false, logs
	}

	logs = append(logs, fmt.Sprintf("resp:%#v", resp))

	if len(resp.UHostIds) != *req.MaxCount {
		reportFail(ctx, prog, block, fmt.Sprintf("expect uhost count %d, accept %d", count, len(resp.UHostIds)))
		return false, logs
	}
	for _, uhostID := range resp.UHostIds {
		rc.add(cli.OpResultRow{ResourceID: uhostID, Action: "create", Status: "Initializing"})
	}
	for i, uhostID := range resp.UHostIds {
		block = prog.NewBlock()

		text := fmt.Sprintf("the uhost[%s]", uhostID)
		if len(req.Disks) > 1 {
			text = fmt.Sprintf("%s which attached a data disk", text)
			if len(req.NetworkInterface) > 0 {
				text = fmt.Sprintf("%s and binded an eip", text)
			}
		} else if len(req.NetworkInterface) > 0 {
			text = fmt.Sprintf("%s which binded an eip", text)
		}
		text = fmt.Sprintf("%s is initializing", text)

		if async {
			block.Append(text)
		} else {
			prog.Sspoll(sdescribeUHostByID(ctx), uhostID, text, []string{HOST_RUNNING, HOST_FAIL}, block, &req.CommonBase)
		}
		bindEipID := ""
		if len(bindEipIDs) > i {
			bindEipID = bindEipIDs[i]
		}

		if bindEipID != "" {
			eip := ctx.PickResourceID(bindEipID)
			logs = append(logs, fmt.Sprintf("bind eip: %s", eip))
			eipLogs, err := sbindEIP(ctx, sdk.String(uhostID), sdk.String("uhost"), &eip, req.ProjectId, req.Region)
			logs = append(logs, eipLogs...)
			if err != nil {
				reportFail(ctx, prog, block, fmt.Sprintf("bind eip[%s] with uhost[%s] failed: %v", eip, uhostID, err))
				return false, logs
			}
			block.Append(fmt.Sprintf("bind eip[%s] with uhost[%s] successfully", eip, uhostID))
		} else if len(req.NetworkInterface) > 0 {
			ipSet, err := getEIPByUHostId(ctx, uhostID)
			if err != nil {
				reportFail(ctx, prog, block, err.Error())
				return false, logs
			}
			block.Append(fmt.Sprintf("IP:%s  Line:%s", ipSet.IP, ipSet.Type))
			if *updateEIPReq.Name != "" || *updateEIPReq.Remark != "" {
				var message string
				if *updateEIPReq.Name != "" && *updateEIPReq.Remark != "" {
					message = "name and remark"
				} else if *updateEIPReq.Name != "" {
					message = "name"
				} else {
					message = "remark"
				}

				logs = append(logs, fmt.Sprintf("update attribute %s of eip[%s] binded uhost[%s]", message, ipSet.IPId, uhostID))
				updateEIPReq.EIPId = sdk.String(ipSet.IPId)
				_, err = unetClient.UpdateEIPAttribute(updateEIPReq)
				if err != nil {
					reportFail(ctx, prog, block, fmt.Sprintf("update attribute %s of eip[%s] binded uhost[%s] got err, %s", message, ipSet.IPId, uhostID, err))
					return false, logs
				}
				block.Append(fmt.Sprintf("update attribute %s of eip[%s] binded uhost[%s] successfully", message, ipSet.IPId, uhostID))
			}
		}
	}
	return true, logs
}

func createUhost(ctx *cli.Context, prog *cli.Progress, client *uhostsdk.UHostClient, unetClient *unet.UNetClient, req *uhostsdk.CreateUHostInstanceRequest, updateEIPReq *unet.UpdateEIPAttributeRequest, bindEipID string, async bool, rc *resultCollector) (bool, []string) {
	resp, err := client.CreateUHostInstance(req)
	block := prog.NewBlock()
	logs := []string{"=================================================="}
	if err != nil {
		logs = append(logs, fmt.Sprintf("err:%v", err))
		reportFail(ctx, prog, block, cli.ParseError(err))
		return false, logs
	}

	logs = append(logs, fmt.Sprintf("resp:%#v", resp))
	if len(resp.UHostIds) != 1 {
		reportFail(ctx, prog, block, fmt.Sprintf("expect uhost count 1 , accept %d", len(resp.UHostIds)))
		return false, logs
	}
	rc.add(cli.OpResultRow{ResourceID: resp.UHostIds[0], Action: "create", Status: "Initializing"})
	text := fmt.Sprintf("the uhost[%s]", resp.UHostIds[0])
	if len(req.Disks) > 1 {
		text = fmt.Sprintf("%s which attached a data disk", text)
		if len(req.NetworkInterface) > 0 {
			text = fmt.Sprintf("%s and binded an eip", text)
		}
	} else if len(req.NetworkInterface) > 0 {
		text = fmt.Sprintf("%s which binded an eip", text)
	}
	text = fmt.Sprintf("%s is initializing", text)

	if async {
		block.Append(text)
	} else {
		prog.Sspoll(sdescribeUHostByID(ctx), resp.UHostIds[0], text, []string{HOST_RUNNING, HOST_FAIL}, block, &req.CommonBase)
	}

	if bindEipID != "" {
		eip := ctx.PickResourceID(bindEipID)
		logs = append(logs, fmt.Sprintf("bind eip: %s", eip))
		eipLogs, err := sbindEIP(ctx, sdk.String(resp.UHostIds[0]), sdk.String("uhost"), &eip, req.ProjectId, req.Region)
		logs = append(logs, eipLogs...)
		if err != nil {
			reportFail(ctx, prog, block, fmt.Sprintf("bind eip[%s] with uhost[%s] failed: %v", eip, resp.UHostIds[0], err))
			return false, logs
		}
		block.Append(fmt.Sprintf("bind eip[%s] with uhost[%s] successfully", eip, resp.UHostIds[0]))
	} else if len(req.NetworkInterface) > 0 {
		ipSet, err := getEIPByUHostId(ctx, resp.UHostIds[0])
		if err != nil {
			reportFail(ctx, prog, block, err.Error())
			return false, logs
		}
		block.Append(fmt.Sprintf("IP:%s  Line:%s", ipSet.IP, ipSet.Type))
		if *updateEIPReq.Name != "" || *updateEIPReq.Remark != "" {
			var message string
			if *updateEIPReq.Name != "" && *updateEIPReq.Remark != "" {
				message = "name and remark"
			} else if *updateEIPReq.Name != "" {
				message = "name"
			} else {
				message = "remark"
			}

			logs = append(logs, fmt.Sprintf("update attribute %s of eip[%s] binded uhost[%s]", message, ipSet.IPId, resp.UHostIds[0]))
			updateEIPReq.EIPId = sdk.String(ipSet.IPId)
			_, err = unetClient.UpdateEIPAttribute(updateEIPReq)
			if err != nil {
				reportFail(ctx, prog, block, fmt.Sprintf("update attribute %s of eip[%s] binded uhost[%s] got err, %s", message, ipSet.IPId, resp.UHostIds[0], err))
				return false, logs
			}
			block.Append(fmt.Sprintf("update attribute %s of eip[%s] binded uhost[%s] successfully", message, ipSet.IPId, resp.UHostIds[0]))
		}
	}
	return true, logs
}

// getEIPByUHostId polls (up to 6 times) for a non-private EIP bound to the uhost.
// Ported verbatim from cmd/uhost.go getEIPByUHostId (base.BizClient →
// cli.NewServiceClient).
func getEIPByUHostId(ctx *cli.Context, uhostId string) (*uhostsdk.UHostIPSet, error) {
	if uhostId == "" {
		return nil, fmt.Errorf("the uhost[%s] is not found", uhostId)
	}
	client := cli.NewServiceClient(ctx, uhostsdk.NewClient)
	for i := 0; i <= 5; i++ {
		req := client.NewDescribeUHostInstanceRequest()
		req.UHostIds = []string{uhostId}

		resp, err := client.DescribeUHostInstance(req)
		if err != nil {
			return nil, err
		}
		if len(resp.UHostSet) < 1 {
			return nil, fmt.Errorf("the uhost[%s] is not found", uhostId)
		}

		if len(resp.UHostSet[0].IPSet) > 0 {
			for _, v := range resp.UHostSet[0].IPSet {
				if v.Type != "Private" && v.IPId != "" {
					return &v, nil
				}
			}
		}

		time.Sleep(1 * time.Second)
	}

	return nil, fmt.Errorf("can not get eip by uhost[%s]", uhostId)
}
