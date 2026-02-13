# Ruijie Go - 燕山大学锐捷V2网络认证命令行工具

这是燕山大学锐捷V2网络认证系统的Go语言命令行工具，提供与Python版本相同的功能。

## 功能特性

- **网络登录**: 通过CAS-SSO直接登录，使用AES-ECB加密认证
- **服务选择**: 支持多种网络服务（校园网、中国联通、中国电信、中国移动）
- **状态检查**: 查看当前登录状态和用户信息
- **账户信息**: 获取详细的账户信息
- **网络登出**: 安全退出网络连接
- **代理支持**: 支持HTTP/HTTPS/SOCKS5代理
- **交互式操作**: 支持交互式输入用户名、密码和服务选择

## 安装

### 从源码编译

```bash
git clone <repository-url>
cd ruijie-go
go build -o ruijie-go
```

### 直接下载

从 [Releases](releases) 页面下载对应平台的预编译二进制文件。

## 使用方法

### 基本命令

```bash
# 登录（交互式）
./ruijie-go login

# 使用用户名密码登录
./ruijie-go login -u 1145141919810 -p mypassword

# 登录到指定服务
./ruijie-go login -s campus
./ruijie-go login -s 1  # 使用数字别名

# 交互式服务选择
./ruijie-go login -s

# 检查登录状态
./ruijie-go status

# 查看账户信息
./ruijie-go info

# 登出
./ruijie-go logout

# 显示帮助
./ruijie-go --help
```

### 服务别名

支持以下服务别名，方便非中文终端使用：

- `campus` 或 `1` → 校园网
- `unicom` 或 `2` → 中国联通
- `telecom` 或 `3` → 中国电信
- `mobile` 或 `4` → 中国移动

### 环境变量

可以通过环境变量设置默认值：

```bash
export RUIJIE_USERNAME=your_username
export RUIJIE_PASSWORD=your_password
export RUIJIE_SERVICE=校园网
export RUIJIE_VERBOSE=true
export HTTP_PROXY=http://proxy.example.com:8080
export HTTPS_PROXY=https://proxy.example.com:8080
```

### 代理设置

```bash
# 使用命令行参数
./ruijie-go login --proxy socks5://127.0.0.1:1080

# 使用环境变量
export HTTP_PROXY=http://proxy.example.com:8080
export HTTPS_PROXY=https://proxy.example.com:8080
```

### 详细输出

```bash
# 启用详细输出模式
./ruijie-go login -v
./ruijie-go status --verbose
```

## 认证流程

工具使用CAS-SSO直接登录流程（与浏览器实际使用的流程一致）：

1. 重定向到门户获取会话信息（sessionId等参数）
2. 访问 `cas-sso/login` 页面，提取AES密钥（croypto）和流程密钥（execution）
3. 使用AES-ECB加密密码，提交登录表单
4. 验证登录成功（检查ticket或auth-success重定向）
5. 选择网络服务并完成认证

## 配置文件

支持YAML格式的配置文件，默认位置：`~/.ruijie-go.yaml`

```yaml
username: your_username
password: your_password
service: 校园网
verbose: false
proxy: ""
```

## 错误处理

工具提供友好的错误消息：

- 网络连接错误
- 认证失败
- 验证码错误
- 服务器错误
- 门户访问失败

## 开发

### 项目结构

```
ruijie-go/
├── main.go                 # 程序入口
├── cmd/                    # CLI命令
│   ├── root.go            # 根命令
│   ├── login.go           # 登录命令
│   ├── logout.go          # 登出命令
│   ├── status.go          # 状态命令
│   └── info.go            # 信息命令
├── internal/
│   ├── client/            # 客户端实现
│   │   ├── ruijie.go      # 锐捷客户端（含CAS-SSO登录）
│   │   └── cas.go         # （已废弃）
│   ├── config/            # 配置管理
│   │   └── config.go
│   └── utils/             # 工具函数
│       ├── crypto.go      # AES-ECB加密工具
│       ├── captcha.go     # 验证码处理（已废弃）
│       └── display.go     # 输出格式化
├── go.mod
└── README.md
```

### 依赖

- `github.com/spf13/cobra` - CLI框架
- `github.com/spf13/viper` - 配置管理
- `github.com/go-resty/resty/v2` - HTTP客户端
- `github.com/PuerkitoBio/goquery` - HTML解析
- `golang.org/x/term` - 终端输入处理

### 构建

```bash
# 开发构建
go build -o ruijie-go

# 发布构建
go build -ldflags "-s -w" -o ruijie-go

# 交叉编译
GOOS=linux GOARCH=amd64 go build -o ruijie-go-linux-amd64
GOOS=windows GOARCH=amd64 go build -o ruijie-go-windows-amd64.exe
GOOS=darwin GOARCH=amd64 go build -o ruijie-go-darwin-amd64
```

## 许可证

本项目采用与原Python版本相同的许可证。

## 贡献

欢迎提交Issue和Pull Request！

## 更新日志

### v2.0.0
- 使用CAS-SSO直接登录替代旧的CAS+SAM认证流程
- 加密方式从AES-CBC切换为AES-ECB（匹配新的认证接口）
- 移除验证码处理（新流程不需要）
- 修复重定向URL追踪问题
- 修复JavaScript重定向解析问题

### v1.0.0
- 完整的Go语言重写
- 保持与Python版本的功能兼容性
- 改进的错误处理和用户体验
- 跨平台支持
- 单二进制文件分发
