## UHadoop CLI 测试清单

### 准备工作

```bash
# 1. 确认凭据已配置
ucloud config list

# 2. 查看可用地域
ucloud region

# 3. 设置 GOROOT 使用正确的 Go 工具链（g use 1.25.12 后需要）
export GOROOT=/Users/user/.g/go GOTOOLCHAIN=local

# 4. 记录测试用的 region 和 zone，后续命令中替换
# 示例: --region cn-bj2 --zone cn-bj2-02
```

> **编译/运行命令前缀**: 后续所有 `go run/build/test` 命令需加 `GOROOT=/Users/user/.g/go GOTOOLCHAIN=local`

### 测试用参数

| 参数 | 示例值 | 说明 |
|------|--------|------|
| `<region>` | `cn-bj2` | 地域 |
| `<zone>` | `cn-bj2-02` | 可用区 |
| `<instance-id>` | `uhadoop-xxxxx` | 已有集群 ID |
| `<vpc-id>` | `uvpc-xxxxx` | VPC ID（可选，不传则自动发现） |
| `<subnet-id>` | `subnet-xxxxx` | 子网 ID（可选，不传则自动发现） |

---

### 一、只读命令测试

#### 1.1 list-framework-app — 查询框架和应用

```bash
# help
go run . uhadoop list-framework-app --help

# table 输出
go run . uhadoop list-framework-app --region <region> --zone <zone>

# JSON 输出
go run . uhadoop list-framework-app --region <region> --zone <zone> --output json
```

**验证点**：返回表格含 Framework、FrameworkVersion、ReleaseVersion、UseCase、Apps、Versions、MustHas 列。

---

#### 1.2 list-node-type — 查询可用机型

```bash
# help
go run . uhadoop list-node-type --help

# 全部机型
go run . uhadoop list-node-type --region <region> --zone <zone>

# 按框架过滤
go run . uhadoop list-node-type \
  --region <region> --zone <zone> \
  --framework Hadoop --framework-version 3.3.4-udh3.2

# 按角色过滤
go run . uhadoop list-node-type \
  --region <region> --zone <zone> \
  --node-role master

# JSON 输出
go run . uhadoop list-node-type --region <region> --zone <zone> --output json
```

**验证点**：返回表格含 NodeType、HostType、CPU、Memory、SuitableRole、IsUsable、GpuType、GpuCount、DiskMinSize、DiskMaxSize 列。

---

#### 1.3 list — 列出集群

```bash
# help
go run . uhadoop list --help

# 基本查询
go run . uhadoop list --region <region>

# 按 zone 过滤
go run . uhadoop list --region <region> --zone <zone>

# 分页
go run . uhadoop list --region <region> --limit 5 --offset 0

# 只输出 ID
go run . uhadoop list --region <region> --id-only

# JSON 输出
go run . uhadoop list --region <region> --output json

# 跨地域
go run . uhadoop list --all-region --output json
```

**验证点**：表格含 InstanceId、InstanceName、Framework、ReleaseVersion、HadoopVersion、State、Zone、CreateTime、ExpireTime 列。

---

#### 1.4 describe — 查询集群详情

```bash
# help
go run . uhadoop describe --help

# 查询（替换 <instance-id> 为上一步 list 返回的 ID）
go run . uhadoop describe <instance-id> --region <region> --zone <zone>

# JSON 输出
go run . uhadoop describe <instance-id> --region <region> --zone <zone> --output json
```

**验证点**：返回 ClusterSet 数组中包含 InstanceId、NodeSet、AppConfigSet 等详情。

---

### 二、修改类命令测试（谨慎执行）

> ⚠️ 以下命令会对真实资源产生修改，请在明确后果后执行。

#### 2.1 restart-service — 启停服务

```bash
# help
go run . uhadoop restart-service --help

# 重启服务
go run . uhadoop restart-service \
  --region <region> --zone <zone> \
  --instance-id <instance-id> \
  --service-name Hive

# 只停止
go run . uhadoop restart-service \
  --region <region> --zone <zone> \
  --instance-id <instance-id> \
  --service-name Hive --only-stop

# 只启动
go run . uhadoop restart-service \
  --region <region> --zone <zone> \
  --instance-id <instance-id> \
  --service-name Hive --only-start

# 指定节点重启
go run . uhadoop restart-service \
  --region <region> --zone <zone> \
  --instance-id <instance-id> \
  --service-name Hive \
  --node-role core --node-role task
```

**验证点**：返回 `{"RetCode": 0, "State": "running"}`。

---

#### 2.2 add-node — 扩容节点

```bash
# help
go run . uhadoop add-node --help

# Core 节点扩容（磁盘参数有默认值，可省略）
go run . uhadoop add-node \
  --region <region> --zone <zone> \
  --instance-id <instance-id> \
  --node-role core \
  --node-type o.hadoop2m.xlarge \
  --node-count 1

# Task 节点扩容
go run . uhadoop add-node \
  --region <region> --zone <zone> \
  --instance-id <instance-id> \
  --node-role task \
  --node-type o.hadoop2m.xlarge \
  --node-count 1

# Client 节点扩容（需要密码，明文即可，自动 base64 编码）
go run . uhadoop add-node \
  --region <region> --zone <zone> \
  --instance-id <instance-id> \
  --node-role client \
  --node-type o.hadoop2m.medium \
  --node-count 1 \
  --password 'YourPass!'
```

**验证点**：返回 `{"RetCode": 0}`。

---

#### 2.3 upgrade-node — 升级节点机型

```bash
# help
go run . uhadoop upgrade-node --help

# Master 升级（不需要 --node-name）
go run . uhadoop upgrade-node \
  --region <region> --zone <zone> \
  --instance-id <instance-id> \
  --node-role master \
  --node-type o.hadoop2m.xlarge

# Core/Task 升级（需要 --node-name，从 describe 返回的 NodeSet 中获取）
go run . uhadoop upgrade-node \
  --region <region> --zone <zone> \
  --instance-id <instance-id> \
  --node-role core \
  --node-type o.hadoop2m.2xlarge \
  --node-name <instance-id>-core1
```

**验证点**：返回 `{"RetCode": 0}`。

---

#### 2.4 upgrade-disk — 扩容磁盘

```bash
# help
go run . uhadoop upgrade-disk --help

# Core 节点扩容数据盘到 500G
go run . uhadoop upgrade-disk \
  --region <region> --zone <zone> \
  --instance-id <instance-id> \
  --node-role core \
  --data-disk-size 500 \
  --node-name <instance-id>-core1

# Master 节点扩容（不需要 --node-name）
go run . uhadoop upgrade-disk \
  --region <region> --zone <zone> \
  --instance-id <instance-id> \
  --node-role master \
  --data-disk-size 500 \
  --boot-disk-size 100
```

**验证点**：返回 `{"RetCode": 0}`。

---

#### 2.5 create — 创建集群

```bash
# help
go run . uhadoop create --help

# Hadoop 集群 — 使用 Spark 模板（推荐方式）
go run . uhadoop create \
  --region <region> --zone <zone> \
  --name test-uhadoop-cli \
  --framework Hadoop \
  --framework-version 3.3.4-udh3.2 \
  --password 'YourPass123!' \
  --cluster-case Spark \
  --master-node-type o.hadoop2m.xlarge \
  --core-node-type o.hadoop2m.xlarge

# Hadoop 集群 — Hbase 场景
go run . uhadoop create \
  --region <region> --zone <zone> \
  --name test-hbase-cli \
  --framework Hadoop \
  --framework-version 3.3.4-udh3.2 \
  --password 'YourPass123!' \
  --cluster-case Hbase \
  --master-node-type o.hadoop2m.xlarge \
  --core-node-type o.hadoop2m.xlarge

# Hadoop 集群 — Core Hadoop 场景
go run . uhadoop create \
  --region <region> --zone <zone> \
  --name test-core-cli \
  --framework Hadoop \
  --framework-version 3.3.4-udh3.2 \
  --password 'YourPass123!' \
  --cluster-case Core-Hadoop \
  --master-node-type o.hadoop2m.xlarge \
  --core-node-type o.hadoop2m.xlarge

# Hadoop 集群 — 手动指定 app-config（不使用模板）
go run . uhadoop create \
  --region <region> --zone <zone> \
  --name test-manual-cli \
  --framework Hadoop \
  --framework-version 3.3.4-udh3.2 \
  --password 'YourPass123!' \
  --master-node-type o.hadoop2m.xlarge \
  --core-node-type o.hadoop2m.xlarge \
  --app-config Spark#3.5.3 \
  --app-config Hive#3.1.3 \
  --app-config Zookeeper#3.8.4 \
  --app-config Hue#4.11.0 \
  --app-config Hdfs#3.3.4

# MR 集群（需要 --storage-cluster-id，含 Task 节点）
go run . uhadoop create \
  --region <region> --zone <zone> \
  --name test-mr-cli \
  --framework MR \
  --framework-version 3.3.4-udh3.2 \
  --password 'YourPass123!' \
  --storage-cluster-id <storage-cluster-id> \
  --cluster-case Spark \
  --master-node-type o.hadoop2m.xlarge \
  --core-node-type o.hadoop2m.xlarge \
  --task-node-type o.hadoop2m.xlarge --task-count 2

# StarRocks 集群（--cluster-case 不适用，手动指定 app-config）
go run . uhadoop create \
  --region <region> --zone <zone> \
  --name test-sr-cli \
  --framework StarRocks-Shared-Data \
  --framework-version 3.3.4-udh3.2 \
  --password 'YourPass123!' \
  --master-node-type o.hadoop2m.xlarge \
  --task-node-type o.hadoop2m.xlarge --task-count 1 \
  --app-config StarRocks#4.1.1
```

**验证点**：返回 `{"RetCode": 0, "InstanceId": "uhadoop-xxxxx"}`。

---

#### 2.6 delete — 删除集群

```bash
# help
go run . uhadoop delete --help

# 删除集群（不释放 EIP）
go run . uhadoop delete <instance-id> \
  --region <region> --zone <zone>

# 删除集群并释放 EIP
go run . uhadoop delete <instance-id> \
  --region <region> --zone <zone> \
  --release-eip
```

**验证点**：返回 `{"RetCode": 0}`。

---

### 三、自动化校验

```bash
# 编译
GOROOT=/Users/user/.g/go GOTOOLCHAIN=local go build ./...

# 运行 uhadoop 模块的 golden 测试
GOROOT=/Users/user/.g/go GOTOOLCHAIN=local go test ./hack/snapshot -v -run 'ProductGoldens/uhadoop'
GOROOT=/Users/user/.g/go GOTOOLCHAIN=local go test ./hack/snapshot -v -run 'ProductCompletionGoldens/uhadoop'

# 运行全部 snapshot 测试
GOROOT=/Users/user/.g/go GOTOOLCHAIN=local go test ./hack/snapshot -v
```

---

### 四、--cluster-case 模板参考

| framework-version | Spark | Hbase | Core-Hadoop |
| --- | --- | --- | --- |
| `3.3.4-udh3.2` | Spark#3.5.3, Hive#3.1.3, Hue#4.11.0, Zookeeper#3.8.4, Mysql#8.0.32, Yarn#3.3.4 | Hbase#2.4.18, Hue#4.11.0, Zookeeper#3.8.4, Mysql#8.0.32, Yarn#3.3.4, Phoenix#5.2.1 | Hive#3.1.3, Hue#4.11.0, Zookeeper#3.8.4, Yarn#3.3.4 |
| `3.3.4-udh3.1` | Spark#3.5.3, Hive#3.1.3, Hue#4.11.0, Zookeeper#3.8.4, Mysql#8.0.32, Yarn#3.3.4 | Hbase#2.4.18, Hue#4.11.0, Zookeeper#3.8.4, Mysql#8.0.32, Yarn#3.3.4, Phoenix#5.2.1 | Hive#3.1.3, Hue#4.11.0, Zookeeper#3.8.4, Yarn#3.3.4 |
| `3.2.1-udh3.0` | Spark#3.3.0, Hive#3.1.3, Hue#4.7.1, Zookeeper#3.6.3, Mysql#5.6.47, Yarn#3.2.1 | Hbase#2.2.4, Hue#4.7.1, Zookeeper#3.6.3, Mysql#5.6.47, Yarn#3.2.1, Phoenix#5.1.2 | Hive#3.1.3, Hue#4.7.1, Zookeeper#3.6.3, Yarn#3.2.1 |
| `2.8.5-udh2.2` | Spark#2.4.6, Hive#2.3.6, Hue#4.7.1, Zookeeper#3.4.13, Mysql#5.6.47, Yarn#2.8.5 | Hbase#1.4.10, Hue#4.7.1, Zookeeper#3.4.13, Mysql#5.6.47, Yarn#2.8.5, Phoenix#4.14.3 | Hive#2.3.6, Hue#4.7.1, Zookeeper#3.4.13, Yarn#2.8.5 |
| `2.6.0-cdh5.13.3` | Spark#2.4.3, Hive#2.3.3, Hue#3.10.0, Zookeeper#3.4.5, Mysql#5.1.73, Yarn#2.6.0 | Hbase#1.2.0, Hue#3.10.0, Zookeeper#3.4.5, Mysql#5.1.73, Yarn#2.6.0, Phoenix#4.14.0 | Hive#2.3.3, Hue#3.10.0, Zookeeper#3.4.5, Yarn#2.6.0 |
| `2.6.0-cdh5.4.9` | Spark#2.0.1, Hive#1.2.1, Hue#3.10.0, Zookeeper#3.4.5, Mysql#5.1.73, Yarn#2.6.0 | Hbase#1.0.0, Hue#3.10.0, Zookeeper#3.4.5, Mysql#5.1.73, Yarn#2.6.0, Phoenix#4.6.0 | Hive#1.2.1, Hue#3.10.0, Zookeeper#3.4.5, Yarn#2.6.0 |

> Hadoop 框架自动追加 Hdfs 组件（版本号按上表推）。

---

### 测试结果记录

| # | 命令 | 状态 | 备注 |
|---|------|------|------|
| 1 | `list-framework-app --help` | ⬜ | |
| 2 | `list-framework-app <args>` | ⬜ | |
| 3 | `list-node-type --help` | ⬜ | |
| 4 | `list-node-type <args>` | ⬜ | |
| 5 | `list --help` | ⬜ | |
| 6 | `list <args>` | ⬜ | |
| 7 | `describe --help` | ⬜ | |
| 8 | `describe <args>` | ⬜ | |
| 9 | `restart-service --help` | ⬜ | |
| 10 | `restart-service <args>` | ⬜ | |
| 11 | `add-node --help` | ⬜ | |
| 12 | `add-node <args>` | ⬜ | |
| 13 | `upgrade-node --help` | ⬜ | |
| 14 | `upgrade-disk --help` | ⬜ | |
| 15 | `create --help` | ⬜ | |
| 16 | `create --cluster-case Spark` | ⬜ | |
| 17 | `create --cluster-case Hbase` | ⬜ | |
| 18 | `create --cluster-case Core-Hadoop` | ⬜ | |
| 19 | `delete --help` | ⬜ | |
| 20 | `go build ./...` | ⬜ | |
| 21 | `go test ./hack/snapshot -v` | ⬜ | |

> ⬜ = 待测试 &nbsp; ✅ = 通过 &nbsp; ❌ = 失败
