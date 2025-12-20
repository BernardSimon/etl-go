# etl-go - 开箱即用的开源ETL工具

etl-go是一个现代化、高性能、易于使用的开源ETL（Extract, Transform, Load）工具，旨在帮助开发者和数据工程师轻松构建和管理数据处理流程。

## 使用文档
访问 [ETL-GO文档站](http://etl.ziyi.chat/)查看使用文档。
## 🌟 特性

- **开箱即用**：内置多种常用的数据源、处理器和目标组件
- **高扩展性**：插件化架构，支持自定义组件开发
- **可视化配置**：提供Web界面进行任务配置和监控
- **多数据源支持**：MySQL、PostgreSQL、SQLite、Doris、CSV、JSON等
- **丰富的处理器**：数据类型转换、行过滤、数据脱敏、列重命名等
- **任务调度**：支持定时任务和手动触发
- **变量管理**：动态配置和SQL变量支持
- **文件管理**：内置文件上传和管理功能
- **日志监控**：完善的日志记录和任务执行监控

## 🏗️ 架构设计

etl-go采用模块化设计，主要包含以下核心组件：

### 核心模块 (etl/core)
- `datasource`: 数据源抽象层
- `source`: 数据提取组件
- `processor`: 数据处理组件
- `sink`: 数据加载组件
- `executor`: 执行器组件
- `variable`: 变量管理组件
- `record`: 数据记录模型
- `params`: 参数定义模型

### 工厂模式 (etl/factory)
统一的组件注册和创建机制，支持动态加载各类ETL组件。

### 执行引擎 (etl/pipeline)
基于Go协程和通道的高性能并发执行引擎，支持流水线式数据处理。

## 🔧 支持的组件

### 数据源 (DataSource)
- MySQL
- PostgreSQL
- SQLite
- Doris

### 数据输入 (Source)
- SQL查询（MySQL、PostgreSQL、SQLite）
- CSV文件
- JSON文件

### 数据处理 (Processor)
- convertType: 数据类型转换
- filterRows: 行过滤
- maskData: 数据脱敏（MD5、SHA256）
- renameColumn: 列重命名
- selectColumns: 列选择

### 数据输出 (Sink)
- SQL表（MySQL、PostgreSQL、SQLite）
- CSV文件
- JSON文件
- Doris快速输出(stream_load)

### 执行器 (Executor)
- SQL执行（MySQL、PostgreSQL、SQLite）

### 变量 (Variable)
- SQL查询变量（MySQL、PostgreSQL、SQLite）

## 🚀 快速开始
### 安装部署（下载编译包）

1. **下载最新版本**
    - 访问 [GitHub Releases 页面](https://github.com/BernardSimon/etl-go/releases)，根据你的操作系统和架构选择最新的发布版本进行下载并解压。

2. **运行服务**
    - Windows 用户可以直接双击运行 `etl-go.exe`。
    - macOS/Linux 用户需打开终端，进入项目目录，并执行命令：
      ```bash
      ./etl-go
      ```

3. **访问 Web 界面**
    - 打开浏览器访问 [http://localhost:8081](http://localhost:8081)。
    - 默认登录凭据为：
        - 用户名：`admin`
        - 密码：`password123`

### 🚀 安装部署（手动编译）

#### 环境要求
- Go 1.24+
- SQLite (默认数据库)

#### 编译步骤

1. **克隆项目源码**
   ```bash
   git clone https://github.com/BernardSimon/etl-go.git
   cd etl-go
   ```

2. **构建前端资源**
   ```bash
   cd ./web
   pnpm install
   pnpm run build
   ```


3. **编译后端服务**
   ```bash
   cd ..  # 返回项目根目录
   go build -o etl-go .
   ```


4. **启动服务**
   ```bash
   ./etl-go
   ```

### 配置文件 (config.yaml)
```yaml
username: admin                    # 管理员用户名
password: password123             # 管理员密码
jwtSecret: <your-jwt-secret>      # JWT密钥
aesKey: <your-aes-key>            # AES加密密钥
initDb: false                     # 是否初始化数据库
logLevel: dev                     # 日志级别 (dev|prod)
serverUrl: localhost:8080         # API服务地址
runWeb: false                     # 是否启动Web界面
webUrl: localhost:8081            # Web界面地址
```


## 🖥️ Web界面

etl-go提供了直观的Web管理界面，包括：
- 数据源管理
- 变量配置
- 任务创建与调度
- 执行日志查看
- 文件管理

访问 `http://localhost:8081` (默认地址) 即可使用。

## 🔧 API接口


## 📋 使用示例

### 1. 创建数据源
通过Web界面或API创建MySQL数据源，配置连接信息。

### 2. 创建ETL任务
配置任务流程：
```
Executor (SQL执行) →Source (SQL查询) → Processor (数据转换) → Sink (目标表) → Executor (SQL执行)
```

### 3. 设置调度
可设置Cron表达式进行定时执行，或手动触发执行。

### 4. 监控执行
通过Web界面查看任务执行状态和日志。

## 🔒 安全特性

- JWT Token认证
- 敏感信息AES加密存储
- SQL注入防护
- 文件访问权限控制

## 🛠️ 开发指南

```
见官网-文档-开发指南
```

## 🤝 贡献

欢迎提交Issue和Pull Request来改进etl-go！

## 📄 许可证

本项目采用Apache License 2.0许可证。

## 🙏 致谢

本项目基于[Go-Pocket-Etl]("https://github.com/changhe626/go-pocket-etl")开发，感谢其作者Changhe626提供的ETL管道代码。