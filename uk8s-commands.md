# `ucloud uk8s` 命令参考

所有命令的 flag/返回列以 `products/uk8s/internal/uk8s/` 下的源码为准。除显式标注的"全局"项外，每个 verb 都有 `region` / `project-id`（来自 `~/.ucloud/config.json` 默认值或显式传入），所有 ID 形如 `<id>/<name>` 时 `PickResourceID` 自动剥成 `<id>`。

---

## 集群操作

### `ucloud uk8s create`

创建 UK8S 集群。同步模式会等集群 `RUNNING`，`--async` 立即返回。

| 必填 flag | 说明 |
|---|---|
| `--name` | 集群名 |
| `--password` | 节点密码。8–30 字符，2/4 字符类（upper/lower/digit/special），CLI 自动 base64 |
| `--vpc-id` | VPC ID（`uvnet-xxx/name`） |
| `--subnet-id` | 子网 ID（`subnet-xxx/name`） |
| `--service-cidr` | Service CIDR（如 `192.168.0.0/16`） |
| `--master-cpu` | Master vCPU，2–64 |
| `--master-memory-mb` | Master 内存 MiB，4096–262144，1024 倍数 |
| `--master-machine-type` | `N` / `C` / `O` / `OS` |
| `--master-zone` | 1 个或 3 个 zone（逗号分隔） |
| `--node-cpu` | Node vCPU，2–64 |
| `--node-count` | 同规格 Node 数，1–10（带 `--node-isolation-group-id` 时 ≤ 8） |
| `--node-memory-mb` | Node 内存 MiB，4096–262144，1024 倍数 |
| `--node-machine-type` | `N` / `C` / `G` / `O` / `OS` |
| `--node-zone` | Node 所在 zone |
| `--image-id` | UK8S UHost 镜像 ID（`uimage-xxx/name`） |
| `--k8s-version` | k8s 版本（`ucloud uk8s version list` 查白名单） |

| 可选 flag | 说明 |
|---|---|
| `--external-api-server` | `Yes` / `No`，是否暴露 API server 公网 |
| `--cluster-domain` | 自定义 cluster domain |
| `--async` | 不等 `RUNNING` 直接返回 |
| `--charge-type` | `Dynamic` / `Month` / `Year` |
| `--quantity` | 购买时长 |
| `--kube-proxy-mode` | `iptables` / `ipvs` |
| `--master-boot-disk-type` / `--master-boot-disk-size-gb` | Master 系统盘 |
| `--master-data-disk-type` / `--master-data-disk-size-gb` | Master 数据盘 |
| `--master-cpu-platform` | Master CPU 平台 |
| `--node-boot-disk-type` / `--node-boot-disk-size-gb` | Node 系统盘 |
| `--node-data-disk-type` / `--node-data-disk-size-gb` | Node 数据盘 |
| `--node-cpu-platform` | Node CPU 平台 |
| `--node-max-pods` | 每节点最大 Pod 数 |
| `--node-isolation-group-id` | 隔离组 ID |
| `--node-labels` | `key=value` 逗号分隔，最多 5 个 |
| `--node-taints` | `key=value:effect` 逗号分隔，最多 5 个 |
| `--node-gpu` / `--node-gpu-type` | 仅 `--node-machine-type G`；`K80` / `P40` / `V100` |
| `--group` | 业务组 |
| `--user-data` / `--user-data-base64` | cloud-init（≤16 KiB 编码后），互斥 |
| `--init-script` / `--init-script-base64` | 安装后脚本，互斥 |

**返回**（创建后，json/yaml 时结构化，table 模式只打状态文本）：

| 字段 | 类型 | 说明 |
|---|---|---|
| `ClusterId` | string | 新集群 ID |

---

### `ucloud uk8s delete`

删除一个或多个集群。`--yes` 跳过确认。

| 必填 flag | 说明 |
|---|---|
| `--cluster-id` | 一个或多个集群 ID（`uk8s-a,uk8s-b` 逗号分隔） |

| 可选 flag | 说明 |
|---|---|
| `--release-udisk` | bool，释放挂载的数据盘 |
| `--yes` / `-y` | 跳过确认 |

**返回**：每条删除的 cluster 一行 `OpResultRow`：

| 字段 | 说明 |
|---|---|
| `ResourceID` | 集群 ID |
| `Action` | `delete` |
| `Status` | `Deleting`（请求已提交） |

---

### `ucloud uk8s list`

列出当前 region/project 下的集群。

| 必填 flag | 说明 |
|---|---|
| _无_ | |

| 可选 flag | 说明 |
|---|---|
| `--cluster-id` | 只列指定 ID（与 `--limit`/`--offset` 配合可分页） |
| `--limit` | 返回数量上限，默认 100 |
| `--offset` | 偏移量，默认 0 |
| `--region` / `--project-id` | 同上 |

**返回**（每行一个集群）：

| 列 | 来源 |
|---|---|
| `ResourceID` | `ClusterId` |
| `Name` | `ClusterName` |
| `K8sVersion` | 集群 k8s 版本 |
| `VPCID` | VPC ID |
| `SubnetID` | 子网 ID |
| `MasterCnt` | Master 节点数 |
| `NodeCnt` | Node 节点数 |
| `Status` | `RUNNING` / `CREATEFAILED` / `ERROR` / `ABNORMAL` 等 |
| `Created` | `CreateTime` 转换的 RFC3339 时间 |

---

### `ucloud uk8s describe`

查看集群详情。

| 必填 flag | 说明 |
|---|---|
| `--cluster-id` | 目标集群 ID |

| 可选 flag | 说明 |
|---|---|
| _无_ | |

**返回**（key/value 描述视图）：

| Attribute | Content |
|---|---|
| `ResourceID` | `ClusterId` |
| `Name` | `ClusterName` |
| `Version` | k8s 版本 |
| `Status` | `RUNNING` / `CREATEFAILED` / `ERROR` / `ABNORMAL` |
| `VPCID` | VPC ID |
| `SubnetID` | 子网 ID |
| `ServiceCIDR` | Service CIDR |
| `PodCIDR` | Pod CIDR |
| `ClusterDomain` | 集群域名 |
| `MasterCount` | Master 节点数 |
| `NodeCount` | Node 节点数 |
| `APIServer` | 内网 API server |
| `ExternalAPIServer` | 外网 API server（启用 `--external-api-server Yes` 时） |
| `KubeProxyMode` | `iptables` / `ipvs` |
| `Created` | `CreateTime` 转 RFC3339 |

---

### `ucloud uk8s get-config`

打印集群 kubeconfig（kubectl 可直接消费的 YAML）。

| 必填 flag | 说明 |
|---|---|
| `--cluster-id` | 目标集群 ID（仅限 `RUNNING` 集群） |

| 可选 flag | 说明 |
|---|---|
| `--external` | 打印外网 kubeconfig（默认内网） |

**返回**：原始 kubeconfig 文本到 stdout，无表格包装。

---

## 节点池操作

### `ucloud uk8s nodegroup add`

向集群新增一个 NodeGroup。返回的 NodeGroupId 是异步创建的，后续可用 `nodegroup list` 查状态。

| 必填 flag | 说明 |
|---|---|
| `--cluster-id` | 目标集群 ID |
| `--name` | NodeGroup 名 |

| 可选 flag | 说明 |
|---|---|
| `--machine-type` | `N` / `C` / `G` / `O` / `OS` |
| `--cpu` | vCPU，2–64 |
| `--memory-mb` | 内存 MiB，4096–262144，1024 倍数 |
| `--image-id` | UK8S 镜像 ID |
| `--subnet-id` | 子网 ID（与 cluster 同一 VPC） |
| `--zone` | 可用区 |
| `--boot-disk-type` / `--boot-disk-size-gb` | 系统盘（40–500 GB） |
| `--data-disk-type` / `--data-disk-size-gb` | 数据盘（20–1000 GB） |
| `--cpu-platform` | 最低 CPU 平台 |
| `--charge-type` | `Dynamic` / `Month` / `Year` |
| `--group` | 业务组 |
| `--gpu` / `--gpu-type` | 仅 `--machine-type G`；`K80` / `P40` / `V100` |

**返回**：

| 字段 | 说明 |
|---|---|
| `ResourceID` | `NodeGroupId` |
| `Action` | `add` |
| `Status` | `Created` |

---

### `ucloud uk8s nodegroup delete`

删除一个 NodeGroup。

| 必填 flag | 说明 |
|---|---|
| `--cluster-id` | 目标集群 ID |
| `--nodegroup-id` | NodeGroup ID（`uk8sng-xxx/name`） |

| 可选 flag | 说明 |
|---|---|
| `--yes` / `-y` | 跳过确认 |

**返回**：

| 字段 | 说明 |
|---|---|
| `ResourceID` | `NodeGroupId` |
| `Action` | `delete` |
| `Status` | `Deleting` |

---

### `ucloud uk8s nodegroup list`

列出指定集群下的所有 NodeGroup。

| 必填 flag | 说明 |
|---|---|
| `--cluster-id` | 目标集群 ID |

| 可选 flag | 说明 |
|---|---|
| _无_ | |

**返回**（每行一个 NodeGroup）：

| 列 | 来源 |
|---|---|
| `ResourceID` | `NodeGroupId` |
| `Name` | `NodeGroupName` |
| `MachineType` | `N` / `C` / `G` / `O` / `OS` |
| `CPU` | vCPU |
| `MemoryMB` | 内存 MiB |
| `NodeCount` | 当前组内节点数（来自 `len(NodeList)`） |
| `ChargeType` | 计费方式 |
| `ImageID` | 镜像 ID |

---

## 节点操作

### `ucloud uk8s node add`

往现有 NodeGroup（隐式选择或通过 `--nodegroup-id`）添加若干同规格 UHost 节点。

| 必填 flag | 说明 |
|---|---|
| `--cluster-id` | 目标集群 ID（必须 `RUNNING`） |
| `--cpu` | vCPU，2–64 |
| `--memory-mb` | 内存 MiB，4096–262144，1024 倍数 |
| `--count` | 节点数，1–50（带 `--isolation-group-id` 时 ≤ 8） |
| `--charge-type` | `Dynamic` / `Month` / `Year` / `Postpay` |
| `--password` | 节点密码（与 create 同规则：8–30 字符，2/4 字符类） |

| 可选 flag | 说明 |
|---|---|
| `--machine-type` | `N` / `C` / `G` / `O` / `OS` |
| `--nodegroup-id` | 目标 NodeGroup（不指定则用 default） |
| `--subnet-id` | 子网 ID |
| `--image-id` | 镜像 ID |
| `--zone` | 可用区（与 subnet 同） |
| `--boot-disk-type` / `--boot-disk-size-gb` | 系统盘 |
| `--data-disk-type` / `--data-disk-size-gb` | 数据盘 |
| `--cpu-platform` | 最低 CPU 平台 |
| `--max-pods` | 每节点最大 Pod 数 |
| `--quantity` | 购买时长（`Dynamic` 时不可设） |
| `--gpu` / `--gpu-type` | 仅 `--machine-type G`；`K80` / `P40` / `V100` |
| `--isolation-group-id` | 隔离组 |
| `--labels` | 节点标签 |
| `--taints` | 节点污点 |
| `--disable-schedule` | 创建后立即 cordon |
| `--user-data` / `--user-data-base64` | cloud-init，互斥 |
| `--init-script` / `--init-script-base64` | 安装后脚本，互斥 |
| `--group` | 业务组 |

**返回**：

| 字段 | 说明 |
|---|---|
| `ResourceID` | 每个新建的 `NodeId`（一行一个） |
| `Action` | `add` |
| `Status` | `Adding` |

---

### `ucloud uk8s node delete`

从集群移除一个或多个节点。

| 必填 flag | 说明 |
|---|---|
| `--cluster-id` | 目标集群 ID |
| `--node-id` | 一个或多个 Node ID（`node-a,node-b` 逗号分隔） |

| 可选 flag | 说明 |
|---|---|
| `--release-data-udisk` | bool，默认 `true`，释放数据盘 |
| `--yes` / `-y` | 跳过确认 |

**返回**：

| 字段 | 说明 |
|---|---|
| `ResourceID` | 每个被删的 `NodeId` |
| `Action` | `delete` |
| `Status` | `Deleting` |

---

### `ucloud uk8s node list`

列出集群所有节点（含 master 和 worker）。

| 必填 flag | 说明 |
|---|---|
| `--cluster-id` | 目标集群 ID |

| 可选 flag | 说明 |
|---|---|
| _无_ | |

**返回**（每行一个节点）：

| 列 | 来源 |
|---|---|
| `ResourceID` | `NodeId` |
| `InstanceID` | UHost 实例 ID |
| `Name` | 实例名 |
| `Role` | `master` / `worker` |
| `Zone` | 可用区 |
| `MachineType` | `N` / `C` / `G` / `O` / `OS` |
| `CPU` | vCPU |
| `MemoryMB` | 内存 MiB |
| `Status` | 节点状态 |
| `OS` | 操作系统 |

---

### `ucloud uk8s node describe`

查看单个节点详情。

| 必填 flag | 说明 |
|---|---|
| `--cluster-id` | 目标集群 ID |
| `--node-id` | 节点 ID 或 IP（`PickResourceID` 接受 `<id>/<name>` 形式；IP 时直接走 `Name` 字段） |

| 可选 flag | 说明 |
|---|---|
| _无_ | |

**返回**（key/value 描述视图）：

| Attribute | Content |
|---|---|
| `Name` | 节点名（IP） |
| `Hostname` | 主机名 |
| `InternalIP` | 内网 IP |
| `ProviderID` | `UCloud://<region>/<node-id>` |
| `CPUCapacity` | vCPU |
| `MemoryCapacity` | 内存（含 `Ki` 后缀） |
| `PodCapacity` | 最大 Pod 数 |
| `AllocatedPods` | 当前已分配 Pod 数 |
| `Unschedulable` | 是否被 cordon |
| `KubeletVersion` | kubelet 版本 |
| `KubeProxyVersion` | kube-proxy 版本 |
| `ContainerRuntime` | containerd / docker |
| `OSImage` | OS 镜像描述 |
| `KernelVersion` | 内核版本 |
| `Labels` | k8s 节点标签（逗号分隔，**可能非常长，建议用 `--output json` 查全**） |
| `Taints` | 节点污点（逗号分隔） |
| `Created` | 创建时间 RFC3339 |

---

## 镜像操作

### `ucloud uk8s image list`

列出 UK8S 在当前 region/zone 兼容的 UHost/PHost 镜像。**`create` 必须从这里挑**——普通 UHost 镜像会被服务端 `Batch Create not support specify sys disk` 拒掉。

| 必填 flag | 说明 |
|---|---|
| _无_ | |

| 可选 flag | 说明 |
|---|---|
| `--zone` | 限定可用区（强烈建议指定，UK8S 镜像与 zone 强绑定） |
| `--region` / `--project-id` | 同上 |

**返回**（每个镜像一行）：

| 列 | 来源 |
|---|---|
| `ResourceID` | `ImageId` |
| `Name` | `ImageName` |
| `ZoneID` | 镜像所属 zone 编号 |
| `ProductType` | `UHost` 或 `PHost` |
| `NotSupportGPU` | bool，标记是否不兼容 GPU 节点 |

---

## 附：版本操作

### `ucloud uk8s version list`

列出 UK8S 当前支持的 k8s 版本（**`create` 的 `--k8s-version` 必须从这里挑**——mock 假数据如 `1.34.5` 跑真 API 会被 `RetCode 94003 invalid K8sVersion` 拒掉）。

| 必填 flag | 说明 |
|---|---|
| _无_ | |

| 可选 flag | 说明 |
|---|---|
| `--kind` | 集群类型，默认 `Dedicated` |

**返回**（每个版本一行）：

| 列 | 来源 |
|---|---|
| `K8sVersion` | k8s 版本号 |
| `ContainerdVersion` | 容器运行时版本 |

---

## 全局可选 flag（所有 verb 都可用）

| Flag | 说明 |
|---|---|
| `--region` | 覆盖 profile 默认 region |
| `--project-id` | 覆盖 profile 默认 project |
| `--public-key` / `--private-key` | 覆盖 profile 凭据 |
| `--base-url` | 覆盖 API endpoint（mock/灰度用） |
| `--timeout-sec` | 单请求超时 |
| `--max-retry-times` | 重试次数 |
| `--output` / `-o` | `table`（默认）/ `json` / `yaml` |
| `--debug` | 打印 HTTP 细节 |
| `--profile` | 切换 profile |
| `--help` / `-h` | 帮助 |

---

## 排错速查

| 错误 | 原因 | 修法 |
|---|---|---|
| `required flag(s) "X" not set` | cobra 必填 flag 缺 | 看 `Usage:` 里 `flags may be` 那一行的清单 |
| `invalid K8sVersion` / RetCode 94003 | `--k8s-version` 不在白名单 | `ucloud uk8s version list` 选一个真版本 |
| `Batch Create not support specify sys disk` | `--image-id` 不是 UK8S UHost | `ucloud uk8s image list` 重新选 |
| `--memory-mb must be ... multiple of 1024` | 内存不是 1024 倍数 | 改成 4096 / 8192 / 12288... |
| `--cpu must be between 2 and 64` | 超出范围 | 调整到 2–64 |
| `--master-zone requires exactly 1 or 3 entries` | 传了 2/4 个 | 改成 1 个（单 AZ）或 3 个（多 AZ HA） |
| `--gpu and --gpu-type are required when --machine-type is G` | 选了 G 但没 GPU | 加 `--gpu 1 --gpu-type V100` |
| `--password must contain at least 2 of: ...` | 弱密码（仅 1 类字符） | 大小写+数字+特殊中至少 2 类 |
| `--count cannot exceed 8 when --isolation-group-id is set` | 隔离组 + 节点数 > 8 | `--count ≤ 8` |
