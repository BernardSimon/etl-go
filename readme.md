# ETL-Go

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

内置 Sink：

- SQL Sink
- CSV Sink
- JSON Sink
- Doris Stream Load Sink

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

### 场景三：数据清洗与脱敏导出

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

### Executor

- SQL Executor

### Variable

- SQL Variable

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
