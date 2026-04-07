[中文](#中文说明) | [English](#english)

# ETL-Go

<a id="中文说明"></a>

[切换到 English](#english) | [ETL-GO 文档站](https://etl.ziyi.chat)

ETL-Go 是一个面向数据集成场景的现代化 ETL 平台，提供可视化任务编排、可扩展组件体系、REST API、任务调度、运行日志、文件资产管理和模板化工作流能力。

它适合这些场景：

- 在数据库、文件和分析系统之间稳定搬运数据
- 用低门槛方式搭建定时同步、清洗、脱敏、格式转换流程
- 需要一套可二开、可嵌入、可扩展的 Go ETL 基础设施
- 希望同时拥有 Web 管理台和后端 API

## 🚀 在线 Demo 体验

**体验地址**：[https://demo.ziyi.chat/](https://demo.ziyi.chat/)

**账号**：admin  
**密码**：password123

> ⚠️ 重要提示：
> - Demo 为公开系统，请勿输入个人数据库账号密码，注意保护数据库信息安全。
> - Demo 系统数据会定时清空。
## 为什么选择 ETL-Go

- 可视化与可编程兼得：既能通过 Web 页面管理任务，也能通过 `/api/v1` 接口接入自己的平台。
- 组件化架构：数据源、Source、Processor、Sink、Executor、Variable 全部解耦，便于扩展。
- 并发 pipeline 引擎：基于 Goroutine 和 Channel 的流水线执行模型，具备批量写入、上下游解耦和取消传播能力。
- 面向生产的任务能力：支持手动执行、定时调度、运行记录、任务模板、任务文件管理和失败排障。
- 内置文件资产中心：上传一次，在任务配置、数据源配置、运行日志中统一复用。
- 更强的类型安全与生命周期管理：核心组件契约已完成强类型化与 context-aware 改造，数据库/HTTP 阻塞操作可感知取消信号。
- 国际化就绪：前后端都支持中英文能力扩展。

## 核心特性

### 1. 可视化工作流编排

- 创建手动任务与定时任务
- 使用任务模板快速复用已有流程
- 从已有任务复制生成新任务
- 通过分区化任务配置弹窗管理复杂流程
- 在保存前预览最终提交结构

### 2. 完整的 ETL 执行链路

一个任务可以由这些阶段组成：

`Before Executor -> Source -> Processors -> Sink -> After Executor`

你可以按需启用：

- `Executor`
  - 运行前准备 SQL
  - 运行后收尾 SQL
- `Source`
  - 从数据库查询读取数据
  - 从 CSV / JSON 文件读取数据
  - 从 HTTP API 拉取数据
  - 消费 Kafka Topic 消息
  - 从 Redis 读取 Hash / List / String 数据
- `Processors`
  - 类型转换
  - 行过滤
  - 数据脱敏
  - 列重命名
  - 列选择
- `Sink`
  - 写入数据库
  - 导出 CSV
  - 导出 JSON
  - 写入 Doris
  - 推送到 HTTP API
  - 发布消息到 Kafka Topic
  - 写入 Redis（Hash / List / String）

### 3. 丰富的数据连接能力

内置 DataSource：

- MySQL
- PostgreSQL
- SQLite
- Doris
- Kafka
- Redis

内置 Source：

- SQL Source
- CSV Source
- JSON Source
- HTTP Source
- Kafka Source
- Redis Source

内置 Sink：

- SQL Sink
- CSV Sink
- JSON Sink
- Doris Stream Load Sink
- HTTP Sink
- Kafka Sink
- Redis Sink

内置 Variable：

- SQL Variable for MySQL / PostgreSQL / SQLite

### 4. 面向真实业务的任务管理

- 任务列表分页与筛选
- 批量启动、批量停止、批量删除
- 手动执行与 Cron 调度
- 任务模板保存、复用、删除
- 任务运行记录跳转与联查
- 运行参数、执行结果、任务文件统一查看

### 5. 文件驱动型数据集成体验

ETL-Go 不把文件上传做成孤立功能，而是把文件当成平台里的“资产”来管理：

- 文件管理页面支持上传、搜索、删除、下载
- 任务参数中的 `file_id` / `file_ids` 可直接选择文件，而不是手填 ID
- 数据源配置中的文件型参数也使用统一文件库
- 大文件上传链路已做专门优化，减少超时和边界问题
- 运行记录与输出文件支持关联查看

### 6. 工程级稳定性改进

当前版本的核心引擎已经具备这些特征：

- factory registry 并发安全
- 组件类型列表稳定排序
- datasource 契约强类型化
- `context.Context` 贯穿 Source / Processor / Sink / Executor / Variable
- SQL / HTTP 阻塞操作支持取消传播
- 任务级 datasource 共享租约，避免多个组件共享同一数据源时提前关闭底层连接
- 更稳妥的 JSON / CSV / SQL 输出顺序与资源释放行为

### 7. 两步验证（TOTP）

支持基于时间的一次性密码（TOTP）二步验证，与 Google Authenticator、Authy 等 App 兼容。

- 在 `config.yaml` 中将 `totpEnabled` 设为 `true` 即可启用，`false` 时完全不影响登录流程
- 首次启动自动生成随机 TOTP Secret，也可自行填写已有密钥
- 启用后，登录时先验证账号密码，通过后前端自动进入验证码输入步骤，验证通过才下发 JWT
- 验证码输入步骤有独立的 IP 限流保护，防止暴力破解

**配置方式：**

```yaml
totpEnabled: true
totpSecret: "YOUR_BASE32_SECRET"  # 将此值录入 Google Authenticator / Authy
```

**API 流程：**

```text
POST /api/v1/login → { requires_2fa: true, pre_auth_token: "..." }
POST /api/v1/verify-2fa { pre_auth_token, code } → { token, refresh_token }
```

## 项目结构

```text
.
├── components/          # 内置组件实现
│   ├── datasource/
│   ├── executor/
│   ├── processors/
│   ├── sinks/
│   ├── sources/
│   └── variable/
├── etl/
│   ├── core/            # 核心接口与抽象
│   ├── factory/         # 组件注册与创建
│   └── pipeline/        # 并发执行引擎
├── server/
│   ├── api/             # REST API
│   ├── config/          # 配置加载
│   ├── model/           # 数据模型
│   ├── router/          # 路由注册
│   ├── task/            # 任务编排与执行
│   ├── types/           # 请求/响应结构
│   └── utils/
├── web/                 # 前端管理台
└── main.go              # 服务入口
```

## 架构亮点

### Pipeline 引擎

ETL-Go 使用分阶段并发模型执行任务：

- Source 负责读取
- Processor 链负责转换
- Sink 负责批量写入
- 所有阶段通过 Channel 解耦
- 取消信号通过 `context.Context` 向全链路传递

这使它既适合处理中小规模定时同步，也适合作为团队内部数据自动化平台的基础引擎。

### 组件工厂

所有内置和自定义组件都通过 factory 注册。你不需要在业务层写一堆 `switch case`，而是直接通过类型名称创建组件实例。

### 任务装配层

任务执行前会统一完成：

- 参数构建
- 文件型参数解析
- 变量替换
- datasource 初始化
- 共享 datasource 复用
- pipeline 组装与执行

## 快速开始

### 方式一：直接运行后端服务

适合只使用 API，或者已有独立前端项目。

```bash
git clone https://github.com/BernardSimon/etl-go.git
cd etl-go
go build -o etl-go .
./etl-go
```

默认 API 地址：

- `http://localhost:8080`

### 方式二：前后端分开开发运行

适合本地开发、联调和二次开发。

#### 1. 启动后端

```bash
go build -o etl-go .
./etl-go
```

#### 2. 启动前端

```bash
cd web
npm install
npm run dev
```

默认地址：

- 前端：`http://localhost:5173`
- 后端：`http://localhost:8080`

### 方式三：内建 Web 静态资源

如果你想让后端直接托管前端静态文件，可以先构建前端：

```bash
cd web
npm install
npm run build
cd ..
go build -o etl-go .
./etl-go
```

根据配置项 `runWeb` 和 `webUrl`，后端可以同时提供 Web 页面服务。

## 默认登录信息

首次启动使用 `config.yaml` 中的默认管理员账号：

- 用户名：`admin`
- 密码：`password123`

请在生产环境中务必修改。

## 配置说明

项目默认使用根目录下的 [config.yaml](/Users/szy/Desktop/code/etl-go/config.yaml)。

示例：

```yaml
username: admin
password: password123
jwtSecret: your-jwt-secret
apiSecret: ""
aesKey: your-aes-key
initDb: false
logLevel: dev

log:
  filename: ./log/app.log
  maxSize: 20
  maxBackups: 3
  maxAge: 7
  compress: true

database:
  driver: sqlite        # sqlite（默认）| mysql | postgres
  path: ./data.db       # 仅 driver=sqlite 时有效
  # dsn: ""             # driver=mysql/postgres 时填写连接字符串
  maxOpenConns: 10
  maxIdleConns: 5
  connMaxLifetime: 300

pipeline:
  batchSize: 1000
  channelSize: 10000

serverUrl: 0.0.0.0:8080
runWeb: false
webUrl: 0.0.0.0:8081

corsOrigins:
  - http://localhost:8081
  - http://localhost:5173

totpEnabled: false
totpSecret: "YOUR_BASE32_SECRET"
```

**切换到 MySQL：**

```yaml
database:
  driver: mysql
  dsn: "user:password@tcp(127.0.0.1:3306)/etl?charset=utf8mb4&parseTime=True&loc=Local"
  maxOpenConns: 20
  maxIdleConns: 10
```

**切换到 PostgreSQL：**

```yaml
database:
  driver: postgres
  dsn: "host=127.0.0.1 user=etl password=password dbname=etl port=5432 sslmode=disable"
  maxOpenConns: 20
  maxIdleConns: 10
```

### 关键配置项

- `database`
  - 平台元数据库配置，支持 `sqlite`（默认）、`mysql`、`postgres`
  - `driver`：数据库类型，留空等同于 `sqlite`
  - `path`：SQLite 数据库文件路径，仅 `driver=sqlite` 时有效
  - `dsn`：MySQL / PostgreSQL 连接字符串，`driver` 为这两者时必填
- `pipeline.batchSize`
  - Sink 每次批量写入的记录数
- `pipeline.channelSize`
  - pipeline 各阶段通道缓冲大小
- `serverUrl`
  - API 服务监听地址
- `runWeb`
  - 是否由后端直接提供 Web 页面
- `webUrl`
  - 内建 Web 服务监听地址
- `apiSecret`
  - API 签名鉴权密钥；为空时表示不启用签名鉴权
- `corsOrigins`
  - 允许跨域访问的前端地址
- `totpEnabled`
  - 是否启用 TOTP 两步验证；默认 `false`
- `totpSecret`
  - Base32 编码的 TOTP 密钥，录入 Google Authenticator / Authy 使用；首次启动自动随机生成

### 环境变量覆盖

这些环境变量会覆盖配置文件中的对应值：

- `ETL_USERNAME`
- `ETL_PASSWORD`
- `ETL_JWT_SECRET`
- `ETL_API_SECRET`
- `ETL_AES_KEY`
- `ETL_SERVER_URL`
- `ETL_LOG_LEVEL`

## REST API

后端 API 使用统一前缀：

- `/api/v1`

主要能力包括：

- 登录与鉴权
- 数据源管理
- 系统变量管理
- 任务管理
- 任务模板管理
- 运行记录管理
- 文件管理
- 组件元数据查询

示例登录请求（未启用 2FA）：

```bash
curl 'http://localhost:8080/api/v1/login' \
  -H 'Accept-Language: zh' \
  -H 'Content-Type: application/json' \
  --data-raw '{"username":"admin","password":"password123"}'
```

启用 2FA 后的登录流程：

```bash
# 第一步：账号密码登录，返回 pre_auth_token
curl 'http://localhost:8080/api/v1/login' \
  -H 'Content-Type: application/json' \
  --data-raw '{"username":"admin","password":"password123"}'
# → {"code":0,"data":{"requires_2fa":true,"pre_auth_token":"<token>"}}

# 第二步：提交验证码，返回正式 JWT
curl 'http://localhost:8080/api/v1/verify-2fa' \
  -H 'Content-Type: application/json' \
  --data-raw '{"pre_auth_token":"<token>","code":"123456"}'
# → {"code":0,"data":{"token":"<jwt>","refresh_token":"<refresh>"}}
```

### API 签名鉴权

对于非 Web 场景的服务端 API 调用，受保护接口除了支持 `Authorization` token，也支持 query 参数签名鉴权。

行为说明：

- 当 `apiSecret` 已配置且请求中带有 `timestamp` 或 `sign` 时，服务端按签名方式校验
- 当 `apiSecret` 未配置时，签名鉴权整体关闭，服务端继续按原有 token 方式鉴权
- `apiSecret` 只用于调用方本地和服务端本地计算签名，不应放进 query、body 或 header

启用方式：

- 在 `config.yaml` 中配置 `apiSecret`
- 或设置环境变量 `ETL_API_SECRET`
- 当两者同时存在时，以环境变量 `ETL_API_SECRET` 为准
- 当 `apiSecret` 为空字符串或未填写时，签名鉴权不启用，接口仍只按原有 token 方式鉴权

签名参数：

- `timestamp`
- `sign`

校验规则：

- `timestamp` 使用 Unix 秒级时间戳
- 客户端与服务端时间差必须在 60 秒内
- `sign` 使用 MD5
- 请求 query 中不传 `apiSecret`
- 签名内容由“除 `sign` 外的全部 query 参数” + “请求 body” + “配置中的 `apiSecret`”共同组成

签名拼接规则：

1. 取全部 query 参数，排除 `sign`
2. 按参数名升序排序；同名多值时按值升序排序
3. 拼接为 `k=v&k2=v2`
4. body 为空时按空字符串参与签名；body 为 JSON 时服务端会先压缩成无空白的紧凑 JSON 后再参与签名
5. 将服务端配置的 `apiSecret` 追加到待签名字符串末尾
6. 最终待签名字符串格式为：`<sorted_query_string>&body=<normalized_body>&secret=<apiSecret>`
7. 对该字符串做 MD5，得到 `sign`

示例：

假设：

- `apiSecret=demo-secret`
- 请求路径：`/api/v1/tasks?page_no=1&page_size=10`
- 请求 body：`{"name":"demo","enabled":true}`
- 当前时间戳：`1712300000`

则参与签名的 query 为：

```text
page_no=1&page_size=10&timestamp=1712300000
```

规范化后的 body 为：

```text
{"name":"demo","enabled":true}
```

最终待签名字符串为：

```text
page_no=1&page_size=10&timestamp=1712300000&body={"name":"demo","enabled":true}&secret=demo-secret
```

可以这样生成签名：

```bash
SIGN_SOURCE='page_no=1&page_size=10&timestamp=1712300000&body={"name":"demo","enabled":true}&secret=demo-secret'
SIGN=$(printf '%s' "$SIGN_SOURCE" | openssl dgst -md5 | awk '{print $2}')
```

调用示例：

```bash
TIMESTAMP=$(date +%s)
BODY='{"name":"demo","enabled":true}'
QUERY="page_no=1&page_size=10&timestamp=${TIMESTAMP}"
SIGN_SOURCE="${QUERY}&body=${BODY}&secret=demo-secret"
SIGN=$(printf '%s' "$SIGN_SOURCE" | openssl dgst -md5 | awk '{print $2}')

curl "http://localhost:8080/api/v1/tasks?${QUERY}&sign=${SIGN}" \
  -H 'Content-Type: application/json' \
  --data-raw "${BODY}"
```

## Web 管理台

前端管理台覆盖了日常使用的核心操作：

- 数据源管理
- 变量管理
- 工作流管理
- 任务模板
- 运行日志
- 文件管理
- 国际化切换

当前前端已经具备这些体验特性：

- RESTful API 全量适配
- 中英文国际化补齐
- 表单字段级错误回填
- 文件选择统一文件库
- 移动端与可访问性增强
- 运行日志详情与自动刷新

## 使用路径建议

### 场景一：数据库到数据库同步

1. 创建源数据库 DataSource
2. 创建目标数据库 DataSource
3. 新建任务
4. 配置 SQL Source
5. 可选添加 Processor 链
6. 配置 SQL Sink
7. 保存并执行

### 场景二：文件导入数据库

1. 在文件管理中上传 CSV / JSON 文件
2. 创建任务或数据源配置
3. 在文件参数中选择文件
4. 配置目标数据库 Sink
5. 手动执行或定时执行

### 场景三：从 HTTP API 拉取数据入库

1. 新建任务
2. 配置 HTTP Source，填写 API 地址、认证头、分页方式
3. 可选添加 Processor 链进行数据转换
4. 配置目标数据库 Sink
5. 手动执行或定时执行

### 场景四：数据清洗与脱敏导出

1. 使用 SQL Source 读取原始数据
2. 添加 `filterRows`、`convertType`、`maskData`
3. 输出到 CSV / JSON / Doris

### 场景五：Kafka 消息数据落库

1. 创建 Kafka DataSource（配置 broker 地址）
2. 新建任务，配置 Kafka Source（指定 topic、consumer group、消息数量）
3. 可选添加 Processor 链做类型转换、字段过滤
4. 配置数据库 Sink
5. 手动执行或定时执行

### 场景六：数据库数据写入 Redis 缓存

1. 创建源数据库 DataSource 和 Redis DataSource
2. 新建任务，配置 SQL Source 读取数据
3. 配置 Redis Sink（hash 模式，指定 key_field 和 key_prefix）
4. 手动执行或定时执行

## 内置组件一览

### DataSource

- MySQL
- PostgreSQL
- SQLite
- Doris
- Kafka
- Redis

### Source

- SQL
- CSV
- JSON
- HTTP
- Kafka
- Redis

### Processor

- `convertType`
- `filterRows`
- `maskData`
- `renameColumn`
- `selectColumns`

### Sink

- SQL
- CSV
- JSON
- Doris
- HTTP
- Kafka
- Redis

### Executor

- SQL Executor

### Variable

- SQL Variable

## HTTP 组件详细说明

### HTTP Source

从 HTTP API 拉取 JSON 数据作为数据源，支持分页和嵌套数据提取。

#### 参数

| 参数 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| `url` | 是 | - | 请求地址 |
| `method` | 否 | `GET` | HTTP 方法（GET / POST） |
| `headers` | 否 | - | 请求头，JSON 格式，如 `{"Authorization": "Bearer xxx"}` |
| `body` | 否 | - | 请求体（POST 时使用），JSON 字符串 |
| `pagination_type` | 否 | `none` | 分页方式：`none` / `offset` / `page` / `cursor` |
| `page_size` | 否 | `100` | 每页记录数 |
| `cursor_field` | 否 | `next_cursor` | 游标字段名（cursor 分页时使用），支持点分路径 |
| `data_path` | 否 | - | 数据数组在响应 JSON 中的路径，如 `data.items` |

#### 分页模式

- **`none`**：只请求一次
- **`offset`**：自动附加 `?offset=N&limit=M`，当返回数据量 < page_size 时停止
- **`page`**：自动附加 `?page=N&page_size=M`，当返回数据量 < page_size 时停止
- **`cursor`**：从响应中提取 cursor_field 值附加到下一次请求，cursor 为空时停止

#### 示例

```yaml
# 基础用法 - 从 API 获取 JSON 数组
source:
  type: http
  config:
    url: "https://api.example.com/users"

# 带认证和分页
source:
  type: http
  config:
    url: "https://api.example.com/orders"
    headers: '{"Authorization": "Bearer my-token"}'
    pagination_type: offset
    page_size: "200"
    data_path: data.list
```

### HTTP Sink

将数据推送到 HTTP API，支持自定义请求体结构、签名验证和多种认证方式。

#### 参数

| 参数 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| `url` | 是 | - | 目标 API 地址 |
| `method` | 否 | `POST` | HTTP 方法（POST / PUT / PATCH） |
| `headers` | 否 | - | 自定义请求头，JSON 格式 |
| `auth_type` | 否 | `none` | 认证方式：`none` / `bearer` / `basic` / `api_key` |
| `auth_value` | 否 | - | 认证凭据：Token / `user:password` / API Key 值 |
| `api_key_name` | 否 | `X-API-Key` | API Key 的 Header 名称 |
| `body_template` | 否 | - | 请求体模板（Go template 语法），为空时直接发送 JSON 数组 |
| `send_mode` | 否 | `batch` | 发送模式：`batch`（整批数组）/ `single`（逐条对象） |

#### body_template 详解

通过 Go template 语法自定义请求体结构。不配置时，直接将数据作为 JSON 数组发送。

**可用变量：**

| 变量 | 说明 |
|------|------|
| `.DataJSON` | 数据的 JSON 字符串（batch 模式为数组，single 模式为单个对象） |
| `.Timestamp` | 当前 Unix 时间戳（秒） |
| `.TimestampMs` | 当前 Unix 时间戳（毫秒） |
| `.ID` | 批次 ID |
| `.Count` | 当前批次记录数 |

**内置签名函数：**

| 函数 | 用法 | 说明 |
|------|------|------|
| `hmacSHA256` | `{{hmacSHA256 .DataJSON "secret"}}` | HMAC-SHA256 签名 |
| `md5` | `{{md5 .DataJSON}}` | MD5 哈希 |
| `sha256` | `{{sha256 .DataJSON}}` | SHA-256 哈希 |
| `concat` | `{{concat "a" "b"}}` | 拼接多个字符串 |
| `toString` | `{{toString .Timestamp}}` | 将任意值转为字符串 |

#### 示例

```yaml
# 基础用法 - 直接发送 JSON 数组
sink:
  type: http
  config:
    url: "https://api.example.com/import"
    auth_type: bearer
    auth_value: "my-token"

# 自定义包装结构
sink:
  type: http
  config:
    url: "https://api.example.com/import"
    body_template: '{"code": 0, "data": {{.DataJSON}}}'

# 带时间戳和 HMAC 签名
sink:
  type: http
  config:
    url: "https://api.example.com/import"
    body_template: >
      {"timestamp": {{.Timestamp}},
       "sign": "{{hmacSHA256 (concat .DataJSON (toString .Timestamp)) "secret-key"}}",
       "data": {{.DataJSON}}}

# 逐条发送 + 自定义字段
sink:
  type: http
  config:
    url: "https://api.example.com/record"
    send_mode: single
    body_template: '{"app_id": "myapp", "ts": {{.Timestamp}}, "record": {{.DataJSON}}}'
```

## Kafka 组件详细说明

### Kafka DataSource

管理 Kafka broker 连接配置，供 Kafka Source 和 Kafka Sink 共用。

#### 参数

| 参数 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| `brokers` | 是 | `localhost:9092` | Broker 地址，多个用逗号分隔，如 `host1:9092,host2:9092` |
| `sasl_mechanism` | 否 | - | SASL 认证方式：`PLAIN` / `SCRAM-SHA-256` / `SCRAM-SHA-512`，留空禁用 |
| `sasl_username` | 否 | - | SASL 用户名 |
| `sasl_password` | 否 | - | SASL 密码 |
| `tls_enabled` | 否 | `false` | 是否启用 TLS |

### Kafka Source

从 Kafka Topic 消费消息，每条消息转换为一条 Record。

#### 参数

| 参数 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| `topic` | 是 | - | 要消费的 Kafka Topic |
| `group_id` | 否 | `etl-go-consumer` | Consumer Group ID |
| `value_format` | 否 | `json` | 消息格式：`json`（解析 JSON 字段）/ `string`（整体存入 `value` 字段） |
| `key_field` | 否 | - | 非空时，将消息 Key 存入该字段名 |
| `max_messages` | 否 | `0` | 最多读取多少条消息，`0` 表示读到超时为止 |
| `timeout_seconds` | 否 | `30` | 等待单条消息的超时秒数，超时视为正常结束 |
| `start_offset` | 否 | `earliest` | 起始位置：`earliest`（从头）/ `latest`（仅新消息） |

### Kafka Sink

将 Record 序列化后发布到 Kafka Topic。

#### 参数

| 参数 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| `topic` | 是 | - | 目标 Kafka Topic |
| `value_format` | 否 | `json` | 消息格式：`json`（序列化为 JSON 对象）/ `string`（取 `value` 字段原始字符串） |
| `key_field` | 否 | - | 非空时，取该字段的值作为消息 Key |

## Redis 组件详细说明

### Redis DataSource

管理 Redis 连接，供 Redis Source 和 Redis Sink 共用。

#### 参数

| 参数 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| `addr` | 是 | `localhost:6379` | Redis 服务地址（host:port） |
| `password` | 否 | - | Redis 密码（AUTH），不需要时留空 |
| `db` | 否 | `0` | Redis 数据库编号（0-15） |
| `tls_enabled` | 否 | `false` | 是否启用 TLS |

### Redis Source

从 Redis 读取数据，支持三种模式。

#### 参数

| 参数 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| `mode` | 否 | `hash_scan` | 读取模式：`hash_scan` / `list` / `string_scan` |
| `scan_match` | 否 | `*` | Key 匹配模式（Glob 风格），用于 `hash_scan` 和 `string_scan` |
| `scan_count` | 否 | `100` | 每次 SCAN 返回的 Key 数量提示 |
| `key` | 否 | - | 目标 List Key，`list` 模式必填 |
| `list_start` | 否 | `0` | LRANGE 起始索引（`list` 模式） |
| `list_stop` | 否 | `-1` | LRANGE 结束索引，`-1` 表示末尾（`list` 模式） |
| `value_format` | 否 | `json` | `list` 模式下的元素格式：`json` / `string` |

#### 模式说明

- **`hash_scan`**：SCAN 匹配的 Key，对每个 Key 执行 HGETALL，生成一条 Record，额外包含 `_key` 字段
- **`list`**：LRANGE 指定 Key，每个元素解析为 JSON Record（或 `{"value": ...}` 原始字符串）
- **`string_scan`**：SCAN 匹配的 Key，GET 每个值，生成 `{"key": ..., "value": ...}` Record

### Redis Sink

将 Record 写入 Redis，支持三种模式，使用 Pipeline 批量执行。

#### 参数

| 参数 | 必填 | 默认值 | 说明 |
|------|------|--------|------|
| `mode` | 否 | `hash` | 写入模式：`hash` / `list` / `string` |
| `key_field` | 条件必填 | - | 用于生成 Redis Key 的字段名（`hash` 和 `string` 模式必填） |
| `key_prefix` | 否 | - | Redis Key 前缀，如 `user:`，最终 Key = prefix + record[key_field] |
| `key` | 条件必填 | - | 目标 List Key（`list` 模式必填） |
| `value_field` | 否 | `value` | 用于 `string` 模式写入的值字段名 |

#### 模式说明

- **`hash`**：HSET `key_prefix + record[key_field]`，其余字段作为 Hash 字段写入
- **`list`**：将 Record 序列化为 JSON 后 RPUSH 到指定 Key
- **`string`**：SET `key_prefix + record[key_field]` = `record[value_field]`

## 安全能力

- JWT 鉴权
- 管理员登录保护
- 敏感字段加密存储
- SQL 执行安全校验
- 文件访问路径解析控制
- 国际化错误返回

## 可扩展性

ETL-Go 非常适合做行业内平台化封装。你可以：

- 新增自定义 DataSource
- 新增自定义 Source / Sink / Processor
- 注册自己的 Variable / Executor
- 复用现有 pipeline 引擎
- 基于 `/api/v1` 封装自己的门户或 SaaS 管理后台

组件通过 `etl/factory` 注册，遵循统一接口即可接入系统。

## 开发与测试

### 构建

```bash
go build ./...
```

### 测试

```bash
go test ./...
```

### 前端构建

```bash
cd web
npm install
npm run build
```

## 适用人群

- 数据工程师
- 后端工程师
- 平台工程团队
- 中小团队内部数据平台建设者
- 需要低成本搭建 ETL 中台的业务团队

## Roadmap 方向

ETL-Go 适合继续演进到：

- 更多内置组件
- 更复杂的任务模板体系
- 更强的审计与权限能力
- 更丰富的数据质量校验
- 更细粒度的运行观测与告警

## 贡献

欢迎提交 Issue、讨论设计方向，或直接发起 Pull Request。

如果你在使用 ETL-Go 构建自己的数据平台，也非常欢迎分享实践经验。

## 许可证

本项目使用 Apache License 2.0。

## 致谢

项目最初受到 [go-pocket-etl](https://github.com/changhe626/go-pocket-etl) 的启发，并在此基础上持续演进为一套更完整的任务平台与可视化 ETL 系统。

---

<a id="english"></a>


# ETL-Go

[Switch to 中文](#中文说明) | [ETL-GO Documentation](https://etl.ziyi.chat/en)

ETL-Go is a modern ETL platform for data integration scenarios. It provides visual workflow orchestration, an extensible component system, REST APIs, task scheduling, runtime logs, file asset management, and template-based workflows.

It is a good fit for scenarios such as:

- Moving data reliably between databases, files, and analytics systems
- Building scheduled sync, cleansing, masking, and format-conversion flows with a low learning curve
- Needing a Go-based ETL foundation that is customizable, embeddable, and extensible
- Wanting both a Web admin console and backend APIs



## 🚀 Online Demo

**Demo URL**: [https://demo.ziyi.chat/](https://demo.ziyi.chat/)

**Username**: admin  
**Password**: password123

> ⚠️ **Important Notice:**
> - The demo is a public system. **Do not enter any personal database credentials.**
> - Demo data is periodically cleared.
## Why ETL-Go

- Visual and programmable: manage tasks from the Web UI or integrate with your own platform through `/api/v1`.
- Component-based architecture: DataSource, Source, Processor, Sink, Executor, and Variable are fully decoupled for easier extension.
- Concurrent pipeline engine: built on Goroutines and Channels, with batch writes, stage decoupling, and cancellation propagation.
- Production-oriented task features: supports manual runs, scheduled execution, run records, task templates, file management, and troubleshooting workflows.
- Built-in file asset center: upload once and reuse files across task configs, datasource configs, and runtime logs.
- Stronger type safety and lifecycle management: core contracts are strongly typed and context-aware, and blocking database/HTTP operations can respond to cancellation.
- Ready for internationalization: both backend and frontend are prepared for Chinese and English support.

## Core Features

### 1. Visual Workflow Orchestration

- Create manual and scheduled tasks
- Reuse existing flows quickly with task templates
- Duplicate existing tasks to create new ones
- Manage complex flows with partitioned task configuration dialogs
- Preview the final payload before saving

### 2. End-to-End ETL Execution Chain

A task can consist of the following stages:

`Before Executor -> Source -> Processors -> Sink -> After Executor`

You can enable the following as needed:

- `Executor`
  - Pre-run SQL preparation
  - Post-run SQL cleanup
- `Source`
  - Read data from database queries
  - Read data from CSV / JSON files
  - Fetch data from HTTP APIs
  - Consume messages from Kafka Topics
  - Read Hash / List / String data from Redis
- `Processors`
  - Type conversion
  - Row filtering
  - Data masking
  - Column renaming
  - Column selection
- `Sink`
  - Write to databases
  - Export to CSV
  - Export to JSON
  - Write to Doris
  - Push to HTTP APIs
  - Publish messages to Kafka Topics
  - Write to Redis (Hash / List / String)

### 3. Rich Data Connectivity

Built-in DataSource:

- MySQL
- PostgreSQL
- SQLite
- Doris
- Kafka
- Redis

Built-in Source:

- SQL Source
- CSV Source
- JSON Source
- HTTP Source
- Kafka Source
- Redis Source

Built-in Sink:

- SQL Sink
- CSV Sink
- JSON Sink
- Doris Stream Load Sink
- HTTP Sink
- Kafka Sink
- Redis Sink

Built-in Variable:

- SQL Variable for MySQL / PostgreSQL / SQLite

### 4. Task Management for Real Business Use

- Paginated task list with filters
- Batch start, stop, and delete
- Manual execution and Cron scheduling
- Save, reuse, and delete task templates
- Jump from tasks to run records and cross-reference execution history
- Unified views for runtime parameters, execution results, and task files

### 5. File-Driven Data Integration Experience

ETL-Go does not treat file upload as an isolated feature. Instead, files are managed as platform assets:

- Upload, search, delete, and download from the file management page
- Select files directly for `file_id` / `file_ids` task parameters instead of manually entering IDs
- Reuse the same file library for file-based datasource parameters
- Optimized large-file upload pipeline to reduce timeouts and edge-case issues
- Link runtime records with output files for easier inspection

### 6. Engineering-Grade Stability Improvements

The current engine already includes:

- Concurrent-safe factory registry
- Stable ordering of component type lists
- Strongly typed datasource contracts
- `context.Context` propagated through Source / Processor / Sink / Executor / Variable
- Cancellation propagation for blocking SQL / HTTP operations
- Shared datasource leases at the task level to avoid closing underlying connections too early
- More reliable JSON / CSV / SQL output ordering and resource cleanup

### 7. Two-Factor Authentication (TOTP)

Optional TOTP-based two-factor authentication compatible with Google Authenticator, Authy, and other standard authenticator apps.

- Set `totpEnabled: true` in `config.yaml` to enable; when `false`, the login flow is unchanged
- A random TOTP secret is generated automatically on first startup, or you can supply your own
- When enabled, the login flow first validates the username and password, then prompts the user for a 6-digit code before issuing a JWT
- The code verification step has its own per-IP rate limiting to prevent brute-force attacks

**Configuration:**

```yaml
totpEnabled: true
totpSecret: "YOUR_BASE32_SECRET"  # Scan or enter this value in Google Authenticator / Authy
```

**API flow:**

```text
POST /api/v1/login → { requires_2fa: true, pre_auth_token: "..." }
POST /api/v1/verify-2fa { pre_auth_token, code } → { token, refresh_token }
```

## Project Structure

```text
.
├── components/          # Built-in component implementations
│   ├── datasource/
│   ├── executor/
│   ├── processors/
│   ├── sinks/
│   ├── sources/
│   └── variable/
├── etl/
│   ├── core/            # Core interfaces and abstractions
│   ├── factory/         # Component registration and creation
│   └── pipeline/        # Concurrent execution engine
├── server/
│   ├── api/             # REST API
│   ├── config/          # Configuration loading
│   ├── model/           # Data models
│   ├── router/          # Route registration
│   ├── task/            # Task orchestration and execution
│   ├── types/           # Request/response structures
│   └── utils/
├── web/                 # Frontend admin console
└── main.go              # Service entrypoint
```

## Architecture Highlights

### Pipeline Engine

ETL-Go uses a staged concurrent model to execute tasks:

- Source handles reading
- The Processor chain handles transformation
- Sink handles batch writes
- All stages are decoupled through Channels
- Cancellation signals are propagated through `context.Context`

This makes it suitable for both small to medium scheduled sync jobs and internal data automation platforms.

### Component Factory

All built-in and custom components are registered through the factory. You do not need large `switch case` blocks in business logic. Components are created directly by type name.

### Task Assembly Layer

Before execution, the task assembly layer handles:

- Parameter construction
- File parameter resolution
- Variable substitution
- Datasource initialization
- Shared datasource reuse
- Pipeline assembly and execution

## Quick Start

### Option 1: Run the Backend Service Directly

Suitable if you only need the API or already have a separate frontend.

```bash
git clone https://github.com/BernardSimon/etl-go.git
cd etl-go
go build -o etl-go .
./etl-go
```

Default API address:

- `http://localhost:8080`

### Option 2: Run Backend and Frontend Separately for Development

Suitable for local development, integration testing, and customization.

#### 1. Start the Backend

```bash
go build -o etl-go .
./etl-go
```

#### 2. Start the Frontend

```bash
cd web
npm install
npm run dev
```

Default addresses:

- Frontend: `http://localhost:5173`
- Backend: `http://localhost:8080`

### Option 3: Serve Built-In Web Static Assets

If you want the backend to serve the frontend static files directly, build the frontend first:

```bash
cd web
npm install
npm run build
cd ..
go build -o etl-go .
./etl-go
```

The backend can also serve the Web UI according to the `runWeb` and `webUrl` settings.

## Default Login

On first startup, use the default administrator account from [config.yaml](/Users/szy/Desktop/code/etl-go/config.yaml):

- Username: `admin`
- Password: `password123`

Be sure to change these credentials in production.

## Configuration

The project uses [config.yaml](/Users/szy/Desktop/code/etl-go/config.yaml) in the repository root by default.

Example:

```yaml
username: admin
password: password123
jwtSecret: your-jwt-secret
aesKey: your-aes-key
initDb: false
logLevel: dev

log:
  filename: ./log/app.log
  maxSize: 20
  maxBackups: 3
  maxAge: 7
  compress: true

database:
  driver: sqlite        # sqlite (default) | mysql | postgres
  path: ./data.db       # only used when driver=sqlite
  # dsn: ""             # required when driver=mysql or driver=postgres
  maxOpenConns: 10
  maxIdleConns: 5
  connMaxLifetime: 300

pipeline:
  batchSize: 1000
  channelSize: 10000

serverUrl: 0.0.0.0:8080
runWeb: false
webUrl: 0.0.0.0:8081

corsOrigins:
  - http://localhost:8081
  - http://localhost:5173

totpEnabled: false
totpSecret: "YOUR_BASE32_SECRET"
```

**Switch to MySQL:**

```yaml
database:
  driver: mysql
  dsn: "user:password@tcp(127.0.0.1:3306)/etl?charset=utf8mb4&parseTime=True&loc=Local"
  maxOpenConns: 20
  maxIdleConns: 10
```

**Switch to PostgreSQL:**

```yaml
database:
  driver: postgres
  dsn: "host=127.0.0.1 user=etl password=password dbname=etl port=5432 sslmode=disable"
  maxOpenConns: 20
  maxIdleConns: 10
```

### Key Configuration Items

- `database`
  - Platform metadata database. Supports `sqlite` (default), `mysql`, and `postgres`
  - `driver`: Database type; omitting this field defaults to `sqlite`
  - `path`: SQLite file path; only used when `driver=sqlite`
  - `dsn`: Connection string for MySQL / PostgreSQL; required when using either of those drivers
- `pipeline.batchSize`
  - Number of records written per sink batch
- `pipeline.channelSize`
  - Buffer size of pipeline channels
- `serverUrl`
  - Listening address for the API service
- `runWeb`
  - Whether the backend should serve the Web UI directly
- `webUrl`
  - Listening address for the built-in Web service
- `corsOrigins`
  - Frontend origins allowed for cross-origin requests
- `totpEnabled`
  - Whether to enable TOTP two-factor authentication; defaults to `false`
- `totpSecret`
  - Base32-encoded TOTP secret to register in Google Authenticator / Authy; generated automatically on first startup

### Environment Variable Overrides

These environment variables override values in the config file:

- `ETL_USERNAME`
- `ETL_PASSWORD`
- `ETL_JWT_SECRET`
- `ETL_AES_KEY`
- `ETL_SERVER_URL`
- `ETL_LOG_LEVEL`

## REST API

The backend API uses a unified prefix:

- `/api/v1`

Main capabilities include:

- Login and authentication
- Datasource management
- System variable management
- Task management
- Task template management
- Run record management
- File management
- Component metadata queries

Example login request (2FA disabled):

```bash
curl 'http://localhost:8080/api/v1/login' \
  -H 'Content-Type: application/json' \
  --data-raw '{"username":"admin","password":"password123"}'
```

When 2FA is enabled, login is a two-step flow:

```bash
# Step 1: submit credentials, receive a pre-auth token
curl 'http://localhost:8080/api/v1/login' \
  -H 'Content-Type: application/json' \
  --data-raw '{"username":"admin","password":"password123"}'
# → {"code":0,"data":{"requires_2fa":true,"pre_auth_token":"<token>"}}

# Step 2: submit the authenticator code to receive the final JWT
curl 'http://localhost:8080/api/v1/verify-2fa' \
  -H 'Content-Type: application/json' \
  --data-raw '{"pre_auth_token":"<token>","code":"123456"}'
# → {"code":0,"data":{"token":"<jwt>","refresh_token":"<refresh>"}}
```

## Web Admin Console

The frontend admin console covers the core daily operations:

- Datasource management
- Variable management
- Workflow management
- Task templates
- Runtime logs
- File management
- Language switching

The current frontend already includes:

- Full RESTful API integration
- Complete Chinese/English internationalization
- Field-level form error feedback
- Unified file library selection
- Better mobile and accessibility support
- Runtime log details and auto refresh

## Suggested Usage Paths

### Scenario 1: Database-to-Database Sync

1. Create a source database DataSource
2. Create a target database DataSource
3. Create a new task
4. Configure SQL Source
5. Optionally add a Processor chain
6. Configure SQL Sink
7. Save and run

### Scenario 2: Import Files into a Database

1. Upload CSV / JSON files in file management
2. Create a task or datasource configuration
3. Select the file in file parameters
4. Configure the target database Sink
5. Run manually or on schedule

### Scenario 3: Fetch Data from HTTP API into Database

1. Create a new task
2. Configure HTTP Source with API URL, auth headers, and pagination type
3. Optionally add a Processor chain for data transformation
4. Configure the target database Sink
5. Run manually or on schedule

### Scenario 4: Data Cleansing, Masking, and Export

1. Read raw data with SQL Source
2. Add `filterRows`, `convertType`, and `maskData`
3. Output to CSV / JSON / Doris

### Scenario 5: Kafka Message Data Ingestion

1. Create a Kafka DataSource (configure broker addresses)
2. Create a new task and configure Kafka Source (topic, consumer group, message limit)
3. Optionally add a Processor chain for type conversion or field filtering
4. Configure a database Sink
5. Run manually or on schedule

### Scenario 6: Write Database Data to Redis Cache

1. Create a source database DataSource and a Redis DataSource
2. Create a new task and configure SQL Source to read data
3. Configure Redis Sink (hash mode, set key_field and key_prefix)
4. Run manually or on schedule

## Built-In Components

### DataSource

- MySQL
- PostgreSQL
- SQLite
- Doris
- Kafka
- Redis

### Source

- SQL
- CSV
- JSON
- HTTP
- Kafka
- Redis

### Processor

- `convertType`
- `filterRows`
- `maskData`
- `renameColumn`
- `selectColumns`

### Sink

- SQL
- CSV
- JSON
- Doris
- HTTP
- Kafka
- Redis

### Executor

- SQL Executor

### Variable

- SQL Variable

## HTTP Component Reference

### HTTP Source

Fetch JSON data from HTTP APIs as a data source, with pagination and nested data extraction.

#### Parameters

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `url` | Yes | - | HTTP request URL |
| `method` | No | `GET` | HTTP method (GET / POST) |
| `headers` | No | - | Request headers in JSON format, e.g. `{"Authorization": "Bearer xxx"}` |
| `body` | No | - | Request body for POST requests, JSON string |
| `pagination_type` | No | `none` | Pagination: `none` / `offset` / `page` / `cursor` |
| `page_size` | No | `100` | Records per page |
| `cursor_field` | No | `next_cursor` | Cursor field name in response (for cursor pagination), supports dot-separated path |
| `data_path` | No | - | Dot-separated path to data array in response, e.g. `data.items` |

#### Pagination Modes

- **`none`**: Single request only
- **`offset`**: Appends `?offset=N&limit=M`, stops when returned count < page_size
- **`page`**: Appends `?page=N&page_size=M`, stops when returned count < page_size
- **`cursor`**: Extracts cursor_field from response for next request, stops when cursor is empty

#### Examples

```yaml
# Basic - fetch a JSON array from an API
source:
  type: http
  config:
    url: "https://api.example.com/users"

# With auth and pagination
source:
  type: http
  config:
    url: "https://api.example.com/orders"
    headers: '{"Authorization": "Bearer my-token"}'
    pagination_type: offset
    page_size: "200"
    data_path: data.list
```

### HTTP Sink

Push data to HTTP APIs with custom body structure, signature verification, and multiple auth methods.

#### Parameters

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `url` | Yes | - | Target API URL |
| `method` | No | `POST` | HTTP method (POST / PUT / PATCH) |
| `headers` | No | - | Custom request headers in JSON format |
| `auth_type` | No | `none` | Authentication: `none` / `bearer` / `basic` / `api_key` |
| `auth_value` | No | - | Credential: token string / `user:password` / API key value |
| `api_key_name` | No | `X-API-Key` | Header name for API key auth |
| `body_template` | No | - | Request body template (Go template syntax), sends raw JSON array if empty |
| `send_mode` | No | `batch` | Send mode: `batch` (JSON array) / `single` (one request per record) |

#### body_template Reference

Customize the request body structure using Go template syntax. When not configured, data is sent as a plain JSON array.

**Available Variables:**

| Variable | Description |
|----------|-------------|
| `.DataJSON` | JSON string of the data (array in batch mode, object in single mode) |
| `.Timestamp` | Current Unix timestamp (seconds) |
| `.TimestampMs` | Current Unix timestamp (milliseconds) |
| `.ID` | Batch ID |
| `.Count` | Number of records in the current batch |

**Built-in Signing Functions:**

| Function | Usage | Description |
|----------|-------|-------------|
| `hmacSHA256` | `{{hmacSHA256 .DataJSON "secret"}}` | HMAC-SHA256 signature |
| `md5` | `{{md5 .DataJSON}}` | MD5 hash |
| `sha256` | `{{sha256 .DataJSON}}` | SHA-256 hash |
| `concat` | `{{concat "a" "b"}}` | Concatenate strings |
| `toString` | `{{toString .Timestamp}}` | Convert any value to string |

#### Examples

```yaml
# Basic - send JSON array directly
sink:
  type: http
  config:
    url: "https://api.example.com/import"
    auth_type: bearer
    auth_value: "my-token"

# Custom wrapper structure
sink:
  type: http
  config:
    url: "https://api.example.com/import"
    body_template: '{"code": 0, "data": {{.DataJSON}}}'

# With timestamp and HMAC signature
sink:
  type: http
  config:
    url: "https://api.example.com/import"
    body_template: >
      {"timestamp": {{.Timestamp}},
       "sign": "{{hmacSHA256 (concat .DataJSON (toString .Timestamp)) "secret-key"}}",
       "data": {{.DataJSON}}}

# Single-record mode with custom fields
sink:
  type: http
  config:
    url: "https://api.example.com/record"
    send_mode: single
    body_template: '{"app_id": "myapp", "ts": {{.Timestamp}}, "record": {{.DataJSON}}}'
```

## Kafka Component Reference

### Kafka DataSource

Manages Kafka broker connection configuration, shared by Kafka Source and Kafka Sink.

#### Parameters

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `brokers` | Yes | `localhost:9092` | Broker addresses, comma-separated, e.g. `host1:9092,host2:9092` |
| `sasl_mechanism` | No | - | SASL auth: `PLAIN` / `SCRAM-SHA-256` / `SCRAM-SHA-512`, leave empty to disable |
| `sasl_username` | No | - | SASL username |
| `sasl_password` | No | - | SASL password |
| `tls_enabled` | No | `false` | Enable TLS |

### Kafka Source

Consumes messages from a Kafka Topic and converts each message into a Record.

#### Parameters

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `topic` | Yes | - | Kafka Topic to consume |
| `group_id` | No | `etl-go-consumer` | Consumer Group ID |
| `value_format` | No | `json` | Message format: `json` (parse JSON into fields) / `string` (store raw value in `value` field) |
| `key_field` | No | - | If set, the message Key is stored in this field name |
| `max_messages` | No | `0` | Maximum messages to read; `0` means read until timeout |
| `timeout_seconds` | No | `30` | Timeout in seconds waiting for each message; timeout is treated as normal end |
| `start_offset` | No | `earliest` | Starting position: `earliest` (from beginning) / `latest` (new messages only) |

### Kafka Sink

Serializes Records and publishes them to a Kafka Topic.

#### Parameters

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `topic` | Yes | - | Target Kafka Topic |
| `value_format` | No | `json` | Message format: `json` (serialize as JSON object) / `string` (use raw `value` field) |
| `key_field` | No | - | If set, the value of this field is used as the Kafka message Key |

## Redis Component Reference

### Redis DataSource

Manages Redis connection configuration, shared by Redis Source and Redis Sink.

#### Parameters

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `addr` | Yes | `localhost:6379` | Redis server address (host:port) |
| `password` | No | - | Redis password (AUTH), leave empty if not required |
| `db` | No | `0` | Redis database number (0-15) |
| `tls_enabled` | No | `false` | Enable TLS |

### Redis Source

Reads data from Redis in three modes.

#### Parameters

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `mode` | No | `hash_scan` | Read mode: `hash_scan` / `list` / `string_scan` |
| `scan_match` | No | `*` | Key pattern (glob-style) for `hash_scan` and `string_scan`, e.g. `user:*` |
| `scan_count` | No | `100` | Hint for number of keys per SCAN call |
| `key` | No | - | Target List key; required for `list` mode |
| `list_start` | No | `0` | LRANGE start index (`list` mode) |
| `list_stop` | No | `-1` | LRANGE stop index, `-1` means end (`list` mode) |
| `value_format` | No | `json` | Element format for `list` mode: `json` / `string` |

#### Mode Reference

- **`hash_scan`**: SCAN matching keys, HGETALL each key, produce one Record per key with an extra `_key` field
- **`list`**: LRANGE the specified key, parse each element as a JSON Record (or `{"value": ...}` for string format)
- **`string_scan`**: SCAN matching keys, GET each value, produce `{"key": ..., "value": ...}` Records

### Redis Sink

Writes Records to Redis using pipeline batching. Supports three modes.

#### Parameters

| Parameter | Required | Default | Description |
|-----------|----------|---------|-------------|
| `mode` | No | `hash` | Write mode: `hash` / `list` / `string` |
| `key_field` | Conditional | - | Field name used to generate the Redis key (required for `hash` and `string` modes) |
| `key_prefix` | No | - | Prefix for Redis keys, e.g. `user:`. Final key = prefix + record[key_field] |
| `key` | Conditional | - | Target List key (required for `list` mode) |
| `value_field` | No | `value` | Field name used as the Redis string value (`string` mode) |

#### Mode Reference

- **`hash`**: HSET `key_prefix + record[key_field]` with all other fields as Hash fields
- **`list`**: Serialize Record as JSON and RPUSH to the specified key
- **`string`**: SET `key_prefix + record[key_field]` = `record[value_field]`

## Security

- JWT authentication
- Admin login protection
- Encrypted storage for sensitive fields
- SQL execution safety validation
- File access path resolution control
- Internationalized error responses

## Extensibility

ETL-Go is well suited for packaging into an industry-specific platform. You can:

- Add custom DataSource implementations
- Add custom Source / Sink / Processor implementations
- Register your own Variable / Executor
- Reuse the existing pipeline engine
- Build your own portal or SaaS admin platform on top of `/api/v1`

Components are registered through `etl/factory` and only need to implement the unified interfaces.

## Development and Testing

### Build

```bash
go build ./...
```

### Test

```bash
go test ./...
```

### Frontend Build

```bash
cd web
npm install
npm run build
```

## Who Is It For

- Data engineers
- Backend engineers
- Platform engineering teams
- Small and medium teams building internal data platforms
- Business teams needing a low-cost ETL platform foundation

## Roadmap

ETL-Go can continue evolving toward:

- More built-in components
- More powerful task template systems
- Stronger audit and permission features
- Richer data quality validation
- More detailed runtime observability and alerting

## Contributing

Issues, design discussions, and Pull Requests are all welcome.

If you are building your own data platform with ETL-Go, sharing your experience is also very welcome.

## License

This project is licensed under the Apache License 2.0.

## Acknowledgements

The project was originally inspired by [go-pocket-etl](https://github.com/changhe626/go-pocket-etl) and has since evolved into a more complete task platform and visual ETL system.
