## UHadoop CLI 测试清单

### 准备工作

```bash
# 1. 确认凭据已配置
ucloud config list

# 2. 查看可用地域
ucloud region

# 3. 切换 Go 版本并设置别名（一劳永逸）
g use 1.25.12
alias go=/Users/user/.g/go/bin/go

# 4. 记录测试用的 region 和 zone，后续命令中替换
# 示例: --region cn-bj2 --zone cn-bj2-02
```

---

### 一、只读命令测试

#### 1.1 list-framework-app — 查询框架和应用

| 参数 | 必填 | 说明 |
|------|:--:|------|
| `--region` | ✅ | 地域 |
| `--zone` | ✅ | 可用区 |
| `--project-id` | | 项目 ID |
| `--output` | | json / table / yaml |

```bash
# help
go run . uhadoop list-framework-app --help

# table 输出
go run . uhadoop list-framework-app --region <region> --zone <zone>

# JSON 输出
go run . uhadoop list-framework-app --region <region> --zone <zone> --output json

# 缺参立即报错
go run . uhadoop list-framework-app
# → --region is required
```

---

#### 1.2 list-node-type — 查询可用机型

| 参数 | 必填 | 说明 |
|------|:--:|------|
| `--region` | ✅ | 地域 |
| `--zone` | ✅ | 可用区 |
| `--framework` | | 按框架过滤：Hadoop/HDFS/MR/StarRocks-* |
| `--framework-version` | | 按框架版本过滤 |
| `--node-role` | | 按角色过滤：master/core/task/client |

```bash
# help
go run . uhadoop list-node-type --help

# 全部机型
go run . uhadoop list-node-type --region <region> --zone <zone>

# 按框架+角色过滤
go run . uhadoop list-node-type \
  --region <region> --zone <zone> \
  --framework Hadoop --framework-version 3.3.4-udh3.2 \
  --node-role master

# JSON 输出
go run . uhadoop list-node-type --region <region> --zone <zone> --output json
```

---

#### 1.3 list — 列出集群

| 参数 | 必填 | 说明 |
|------|:--:|------|
| `--region` | ✅ | 地域 |
| `--zone` | | 可用区 |
| `--limit` | | 每页数量，默认 60 |
| `--offset` | | 偏移量，默认 0 |
| `--all-region` | | 跨地域查询 |
| `--id-only` | | 只输出 ID 列表 |

```bash
# help
go run . uhadoop list --help

# 基本查询
go run . uhadoop list --region <region>

# 按 zone 过滤 + 分页
go run . uhadoop list --region <region> --zone <zone> --limit 5 --offset 0

# 只输出 ID
go run . uhadoop list --region <region> --id-only

# JSON 跨地域
go run . uhadoop list --all-region --output json
```

---

#### 1.4 describe — 查询集群详情

| 参数 | 必填 | 说明 |
|------|:--:|------|
| `<instance-id>` | ✅ | 位置参数，跟在 describe 后面 |
| `--region` | ✅ | 地域 |
| `--zone` | ✅ | 可用区 |

```bash
# help
go run . uhadoop describe --help

# 查询
go run . uhadoop describe <instance-id> --region <region> --zone <zone>

# JSON 输出
go run . uhadoop describe <instance-id> --region <region> --zone <zone> --output json
```

---

### 二、修改类命令测试

> ⚠️ 会对真实资源产生修改

#### 2.1 restart-service — 启停服务

| 参数 | 必填 | 说明 |
|------|:--:|------|
| `--instance-id` | ✅ | 集群 ID |
| `--service-name` | ✅ | 服务名：Hive/Spark/Yarn/Hdfs 等 |
| `--region` | ✅ | 地域 |
| `--zone` | ✅ | 可用区 |
| `--application-version` | | 应用版本，传则操作整个应用 |
| `--only-start` | | 只启动 |
| `--only-stop` | | 只停止 |
| `--node-id` | | 指定节点 ID，可多次传 |
| `--node-role` | | 指定节点角色，可多次传 |

```bash
# help
go run . uhadoop restart-service --help

# 重启服务
go run . uhadoop restart-service \
  --region <region> --zone <zone> \
  --instance-id <instance-id> \
  --service-name Hive

# 只停止/只启动
go run . uhadoop restart-service \
  --region <region> --zone <zone> \
  --instance-id <instance-id> \
  --service-name Hive --only-stop

# 指定节点
go run . uhadoop restart-service \
  --region <region> --zone <zone> \
  --instance-id <instance-id> \
  --service-name Hive \
  --node-role core --node-role task
```

---

#### 2.2 add-node — 扩容节点

| 参数 | 必填 | 说明 |
|------|:--:|------|
| `--instance-id` | ✅ | 集群 ID |
| `--node-role` | ✅ | core / task / client |
| `--node-type` | ✅ | 机型，默认 `o.hadoop2m.xlarge` |
| `--region` | ✅ | 地域 |
| `--zone` | ✅ | 可用区 |
| `--node-count` | | 节点数，默认 1 |
| `--password` | | Client 角色必填，明文（自动 base64） |
| `--boot-disk-size` | | 系统盘 GB，默认 50 |
| `--boot-disk-type` | | 系统盘类型，默认 `CLOUD_RSSD` |
| `--data-disk-size` | | 数据盘 GB，默认 200 |
| `--data-disk-num` | | 数据盘数量，默认 1 |
| `--data-disk-type` | | 数据盘类型，默认 `CLOUD_RSSD` |

```bash
# help
go run . uhadoop add-node --help

# Core 扩容（全默认磁盘参数）
go run . uhadoop add-node \
  --region <region> --zone <zone> \
  --instance-id <instance-id> \
  --node-role core \
  --node-type o.hadoop2m.xlarge

# Task 扩容
go run . uhadoop add-node \
  --region <region> --zone <zone> \
  --instance-id <instance-id> \
  --node-role task \
  --node-type o.hadoop2m.xlarge \
  --node-count 1

# Client 扩容（需要密码）
go run . uhadoop add-node \
  --region <region> --zone <zone> \
  --instance-id <instance-id> \
  --node-role client \
  --node-type o.hadoop2m.medium \
  --password 'YourPass!'

# 缺参立即报错
go run . uhadoop add-node --region cn-bj2 --zone cn-bj2-02
# → --instance-id is required
```

---

#### 2.3 upgrade-node — 升级节点机型

| 参数 | 必填 | 说明 |
|------|:--:|------|
| `--instance-id` | ✅ | 集群 ID |
| `--node-role` | ✅ | master / core / task / client |
| `--node-type` | ✅ | 新机型 |
| `--region` | ✅ | 地域 |
| `--zone` | ✅ | 可用区 |
| `--node-name` | | 节点名，非 master 角色必填 |

```bash
# help
go run . uhadoop upgrade-node --help

# Master 升级（不需要 --node-name）
go run . uhadoop upgrade-node \
  --region <region> --zone <zone> \
  --instance-id <instance-id> \
  --node-role master \
  --node-type o.hadoop2m.xlarge

# Core 升级（需要 --node-name，从 describe 的 NodeSet 获取）
go run . uhadoop upgrade-node \
  --region <region> --zone <zone> \
  --instance-id <instance-id> \
  --node-role core \
  --node-type o.hadoop2m.2xlarge \
  --node-name <instance-id>-core1
```

---

#### 2.4 upgrade-disk — 扩容磁盘

| 参数 | 必填 | 说明 |
|------|:--:|------|
| `--instance-id` | ✅ | 集群 ID |
| `--node-role` | ✅ | master / core / task / client |
| `--data-disk-size` | ✅ | 新数据盘大小 GB（必须 > 0） |
| `--region` | ✅ | 地域 |
| `--zone` | ✅ | 可用区 |
| `--boot-disk-size` | | 新系统盘大小 GB |
| `--node-name` | | 节点名，非 master 角色必填 |

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

# Master 扩容（不需要 --node-name）
go run . uhadoop upgrade-disk \
  --region <region> --zone <zone> \
  --instance-id <instance-id> \
  --node-role master \
  --data-disk-size 500 \
  --boot-disk-size 100
```

---

#### 2.5 create — 创建集群

| 参数 | 必填 | 说明 |
|------|:--:|------|
| `--name` | ✅ | 集群名称 |
| `--framework` | ✅ | Hadoop / HDFS / MR / StarRocks-Shared-Nothing / StarRocks-Shared-Data |
| `--framework-version` | ✅ | 版本号，如 `3.3.4-udh3.2` |
| `--password` | ✅ | 登录密码，明文（自动 base64） |
| `--region` | ✅ | 地域 |
| `--zone` | ✅ | 可用区 |
| `--cluster-case` | ① | Spark / Hbase / Core-Hadoop，自动填充 app-config |
| `--app-config` | ① | 手动指定组件，格式 `App#Version`，可多次传 |
| `--master-node-type` | | Master 机型，默认 `o.hadoop2m.xlarge` |
| `--core-node-type` | | Core 机型，默认 `o.hadoop2m.xlarge` |
| `--task-node-type` | | Task 机型（可选），默认 `o.hadoop2m.xlarge` |
| `--master-count` | | Master 数量，默认 2（StarRocks 默认 3）|
| `--core-count` | | Core 数量，默认 3 |
| `--task-count` | | Task 数量，默认 0 |
| `--vpc-id` | | VPC ID，不传自动发现 |
| `--subnet-id` | | Subnet ID，不传自动发现 |
| `--charge-type` | | 付费类型，默认 Month |
| `--storage-cluster-id` | | MR 框架必填 |

> ① `--cluster-case` 和 `--app-config` 二选一，至少填一个。

```bash
# help
go run . uhadoop create --help

# Spark 模板创建（推荐）
go run . uhadoop create \
  --region <region> --zone <zone> \
  --name test-spark \
  --framework Hadoop \
  --framework-version 3.3.4-udh3.2 \
  --password 'YourPass123!' \
  --cluster-case Spark \
  --master-node-type o.hadoop2m.xlarge \
  --core-node-type o.hadoop2m.xlarge

# Hbase 模板创建
go run . uhadoop create \
  --region <region> --zone <zone> \
  --name test-hbase \
  --framework Hadoop \
  --framework-version 3.3.4-udh3.2 \
  --password 'YourPass123!' \
  --cluster-case Hbase \
  --master-node-type o.hadoop2m.xlarge \
  --core-node-type o.hadoop2m.xlarge

# Core-Hadoop 模板创建
go run . uhadoop create \
  --region <region> --zone <zone> \
  --name test-core \
  --framework Hadoop \
  --framework-version 3.3.4-udh3.2 \
  --password 'YourPass123!' \
  --cluster-case Core-Hadoop \
  --master-node-type o.hadoop2m.xlarge \
  --core-node-type o.hadoop2m.xlarge

# 手动指定 app-config
go run . uhadoop create \
  --region <region> --zone <zone> \
  --name test-manual \
  --framework Hadoop \
  --framework-version 3.3.4-udh3.2 \
  --password 'YourPass123!' \
  --master-node-type o.hadoop2m.xlarge \
  --core-node-type o.hadoop2m.xlarge \
  --app-config Spark#3.5.3 \
  --app-config Hive#3.1.3 \
  --app-config Zookeeper#3.8.4 \
  --app-config Hue#4.11.0 \
  --app-config Hdfs#3.3.4 \
  --app-config Yarn#3.3.4

# MR 集群（含 Task 节点）
go run . uhadoop create \
  --region <region> --zone <zone> \
  --name test-mr \
  --framework MR \
  --framework-version 3.3.4-udh3.2 \
  --password 'YourPass123!' \
  --storage-cluster-id <storage-cluster-id> \
  --cluster-case Spark \
  --master-node-type o.hadoop2m.xlarge \
  --core-node-type o.hadoop2m.xlarge \
  --task-node-type o.hadoop2m.xlarge --task-count 2

# 缺参立即报错
go run . uhadoop create --region cn-bj2 --zone cn-bj2-02
# → --name is required
```

---

#### 2.6 delete — 删除集群

| 参数 | 必填 | 说明 |
|------|:--:|------|
| `<instance-id>` | ✅ | 位置参数 |
| `--region` | ✅ | 地域 |
| `--zone` | ✅ | 可用区 |
| `--release-eip` | | 是否释放绑定的 EIP |

```bash
# help
go run . uhadoop delete --help

# 删除（保留 EIP）
go run . uhadoop delete <instance-id> --region <region> --zone <zone>

# 删除并释放 EIP
go run . uhadoop delete <instance-id> --region <region> --zone <zone> --release-eip
```

---

### 三、自动化校验

```bash
# 编译
go build ./...

# 运行 golden 测试
go test ./hack/snapshot -v -run 'ProductGoldens/uhadoop'
go test ./hack/snapshot -v -run 'ProductCompletionGoldens/uhadoop'

# 全部 snapshot
go test ./hack/snapshot -v
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

> Hadoop 框架自动追加 Hdfs 组件，MR/HDFS 框架不追加。

---

### 五、必填参数校验汇总

| 命令 | 必填 |
|------|------|
| `list-framework-app` | `--region` `--zone` |
| `list-node-type` | `--region` `--zone` |
| `list` | `--region` |
| `describe <id>` | `<id>` `--region` `--zone` |
| `create` | `--name` `--framework` `--framework-version` `--password` `--region` `--zone` + `--cluster-case` 或 `--app-config` |
| `delete <id>` | `<id>` `--region` `--zone` |
| `add-node` | `--instance-id` `--node-role` `--node-type` `--region` `--zone` |
| `restart-service` | `--instance-id` `--service-name` `--region` `--zone` |
| `upgrade-node` | `--instance-id` `--node-role` `--node-type` `--region` `--zone` |
| `upgrade-disk` | `--instance-id` `--node-role` `--data-disk-size` `--region` `--zone` |

> 缺任何必填参数都会立即报错退出，不会发起 API 调用。

---

### 测试结果记录

| # | 命令 | 状态 | 备注 |
|---|------|------|------|
| 1 | `list-framework-app --help` | ⬜ | |
| 2 | `list-framework-app --region <r> --zone <z>` | ⬜ | |
| 3 | `list-framework-app` (缺参) | ⬜ | 预期报错 |
| 4 | `list-node-type --help` | ⬜ | |
| 5 | `list-node-type --region <r> --zone <z>` | ⬜ | |
| 6 | `list --help` | ⬜ | |
| 7 | `list --region <r>` | ⬜ | |
| 8 | `describe --help` | ⬜ | |
| 9 | `describe <id> --region <r> --zone <z>` | ⬜ | |
| 10 | `restart-service --help` | ⬜ | |
| 11 | `add-node --help` | ⬜ | |
| 12 | `add-node` (缺参) | ⬜ | 预期报错 |
| 13 | `upgrade-node --help` | ⬜ | |
| 14 | `upgrade-disk --help` | ⬜ | |
| 15 | `create --help` | ⬜ | |
| 16 | `create` (缺参) | ⬜ | 预期报错 |
| 17 | `create --cluster-case Spark` | ⬜ | |
| 18 | `create --cluster-case Hbase` | ⬜ | |
| 19 | `create --cluster-case Core-Hadoop` | ⬜ | |
| 20 | `create --app-config ...` | ⬜ | |
| 21 | `delete --help` | ⬜ | |
| 22 | `go build ./...` | ⬜ | |
| 23 | `go test ./hack/snapshot -v` | ⬜ | |

> ⬜ = 待测试 &ensp; ✅ = 通过 &ensp; ❌ = 失败
