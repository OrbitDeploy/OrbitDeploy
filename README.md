# OrbitDeploy

**Language / 语言版本:** [🇺🇸 English](README.md) | [🇨🇳 中文](README_CN.md)

### 🚀 What is OrbitDeploy?

OrbitDeploy is a self-hosted, web-based platform for managing containerized application deployments. It provides a modern web interface to manage container configurations, reverse proxy, and application deployments with ease.

### ✨ Key Features

- **Reverse Proxy**: Automatic domain management and HTTPS with Caddy
- **Multi-Node Support**: Deploy applications across multiple VPS nodes (in development)
- **Database Support**: Built-in support for PostgreSQL database deployments 
- **CLI Tool**: Command-line interface for automation and scripting (in development)
- **Self-Hosted**: Complete control over your deployment infrastructure
- **Performance Firstd**: Adopt SPA (Single Page Application) for the frontend + SQLite. Due to low performance requirements, this allows more server resources to be allocated for application deployment.

### 🛠️ Technology Stack

**Backend:**
- Go (Echo framework)
- SQLite database
- Podman for container orchestration
- Systemd Quadlets for container management

**Frontend:**
- SolidJS
- TypeScript
- Vite

**Infrastructure:**
- Caddy (reverse proxy and automatic HTTPS)
- Podman (container engine)

### 📦 Quick Start

#### Prerequisites

- Go 1.24 or later
- Node.js and npm (for frontend development)
- Post-2023 Linux distributions or Rolling Release distributions.
- Support Systemd
  
#### Building from Source

1. Clone the repository:
```bash
git clone https://github.com/OrbitDeploy/OrbitDeploy.git
cd OrbitDeploy
```

2. Build the frontend:
```bash
cd frontend
npm install
npm run build
cd ..
```

3. Build and run the backend:
```bash
go build -o orbitdeploy

# Set required environment variables for production
export ORBIT_ENCRYPTION_KEY="your-secure-random-key-here"
export JWT_ACCESS_SECRET="your-jwt-access-secret-here"
export JWT_REFRESH_SECRET="your-jwt-refresh-secret-here"

./orbitdeploy
```

The server will start on port `:8285` by default.

> **Security Note**: See [Environment Variables Configuration](DOC/environment-variables.md) for details on configuring secrets properly for production deployments.

#### Using the CLI

```bash
cd orbitctl
go build -o orbitctl
./orbitctl --help
```



### 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

### 🔒 Security

For security issues, please see our [Security Policy](SECURITY.md).
