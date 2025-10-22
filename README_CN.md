# OrbitDeploy

## 🚀 OrbitDeploy 是什么？

OrbitDeploy 是一个自托管的、基于 Web 的平台，用于管理容器化应用程序的部署。它提供了一个现代化的 Web 界面，可以轻松管理容器配置、反向代理和应用程序部署。

## ✨ 主要特性

- **反向代理**: 使用 Caddy 自动管理域名和 HTTPS
- **多节点支持**: 跨多个 VPS 节点部署应用程序 开发中
- **数据库支持**: 内置对PostgreSQL数据库部署的支持
- **CLI 工具**: 用于自动化和脚本编写的命令行界面 开发中
- **自托管**: 完全控制您的部署基础设施
- **性能优先**: 采用spa实现前端+sqlite，低性能要求，将服务器剩余性能更多用于部署应用程序。

## 🛠️ 技术栈

**后端：**
- Go (Echo 框架)
- SQLite 数据库
- Podman 用于容器编排
- Systemd Quadlets 用于容器管理

**前端：**
- SolidJS
- TypeScript
- Vite

**基础设施：**
- Caddy (反向代理和自动 HTTPS)
- Podman (容器引擎)

## 📦 快速开始

### 前置要求

- Go 1.24 或更高版本
- Node.js 和 npm（用于前端开发）
- 2023以后的liunx发行版或者滚动更新的发行版
- Support Systemd

### 从源码构建

1. 克隆仓库：
```bash
git clone https://github.com/youfun/OrbitDeploy.git
cd OrbitDeploy
```

2. 构建前端：
```bash
cd frontend
npm install
npm run build
cd ..
```

3. 构建并运行后端：
```bash
go build -o orbitdeploy

# 为生产环境设置必需的环境变量
export ORBIT_ENCRYPTION_KEY="your-secure-random-key-here"
export JWT_ACCESS_SECRET="your-jwt-access-secret-here"
export JWT_REFRESH_SECRET="your-jwt-refresh-secret-here"

./orbitdeploy
```

服务器默认将在端口 `:8285` 上启动。

> **安全提示**: 有关如何为生产部署正确配置密钥的详细信息，请参阅[环境变量配置](DOC/environment-variables.md)。

### 使用 CLI

```bash
cd orbitctl
go build -o orbitctl
./orbitctl --help
```


## 🤝 贡献

我们欢迎贡献！详情请参见我们的[贡献指南](CONTRIBUTING.md)。

## 📄 许可证

本项目采用 MIT 许可证 - 详情请参见 [LICENSE](LICENSE) 文件。

## 🔒 安全

有关安全问题，请参见我们的[安全政策](SECURITY_CN.md)。
