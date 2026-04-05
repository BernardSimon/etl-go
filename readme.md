[# 中文](#中文说明) | [English](#english)

# ETL-Go

<a id="中文说明"></a>

[切换到 English](#english)

ETL-Go 是一个面向数据集成场景的现代化 ETL 平台，提供可视化任务编排、可扩展组件体系、REST API、任务调度、运行日志、文件资产管理和模板化工作流能力。

它适合这些场景：

- 在数据库、文件和分析系统之间稳定搬运数据
- 用低门槛方式搭建定时同步、清洗、脱敏、格式转换流程
- 需要一套可二开、可嵌入、可扩展的 Go ETL 基础设施
- 希望同时拥有 Web 管理台和后端 API

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

### 3. 丰富的数据连接能力

内置 DataSource：

- MySQL
- PostgreSQL
- SQLite
- Doris

内置 Source：

- SQL Source
- CSV Source
- JSON Source
- HTTP Source

内置 Sink：

- SQL Sink
- CSV Sink
- JSON Sink
- Doris Stream Load Sink
- HTTP Sink

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
  path: ./data.db
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
```

### 关键配置项

- `database`
  - 平台元数据库配置，默认是 SQLite
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
- `corsOrigins`
  - 允许跨域访问的前端地址

### 环境变量覆盖

这些环境变量会覆盖配置文件中的对应值：

- `ETL_USERNAME`
- `ETL_PASSWORD`
- `ETL_JWT_SECRET`
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

示例登录请求：

```bash
curl 'http://localhost:8080/api/v1/login' \
  -H 'Accept-Language: zh' \
  -H 'Content-Type: application/json' \
  --data-raw '{"username":"admin","password":"password123"}'
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

## 内置组件一览

### DataSource

- MySQL
- PostgreSQL
- SQLite
- Doris

### Source

- SQL
- CSV
- JSON
- HTTP

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

[Switch to 中文](#中文说明)

ETL-Go is a modern ETL platform for data integration scenarios. It provides visual workflow orchestration, an extensible component system, REST APIs, task scheduling, runtime logs, file asset management, and template-based workflows.

It is a good fit for scenarios such as:

- Moving data reliably between databases, files, and analytics systems
- Building scheduled sync, cleansing, masking, and format-conversion flows with a low learning curve
- Needing a Go-based ETL foundation that is customizable, embeddable, and extensible
- Wanting both a Web admin console and backend APIs

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

### 3. Rich Data Connectivity

Built-in DataSource:

- MySQL
- PostgreSQL
- SQLite
- Doris

Built-in Source:

- SQL Source
- CSV Source
- JSON Source
- HTTP Source

Built-in Sink:

- SQL Sink
- CSV Sink
- JSON Sink
- Doris Stream Load Sink
- HTTP Sink

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
  path: ./data.db
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
```

### Key Configuration Items

- `database`
  - Platform metadata database configuration, SQLite by default
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

Example login request:

```bash
curl 'http://localhost:8080/api/v1/login' \
  -H 'Accept-Language: zh' \
  -H 'Content-Type: application/json' \
  --data-raw '{"username":"admin","password":"password123"}'
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

## Built-In Components

### DataSource

- MySQL
- PostgreSQL
- SQLite
- Doris

### Source

- SQL
- CSV
- JSON
- HTTP

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
