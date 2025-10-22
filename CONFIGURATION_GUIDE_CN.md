 # OrbitDeploy 配置指南

本文档将指导您如何安装和配置 OrbitDeploy 应用。

## 1. 安装

我们提供了一个便捷的 `install.sh` 脚本来将 OrbitDeploy 安装为一个 systemd 系统服务。在运行安装脚本之前，请确保您已经根据 [构建说明](README.md#构建) 成功编译了 `orbit-deploy` 二进制文件。

以 root 用户身份运行安装脚本：

```bash
sudo ./install.sh
```

该脚本会自动完成以下工作：

1.  **创建服务用户**：创建一个名为 `orbit-deploy` 的系统用户，用于安全地运行服务。
2.  **创建目录**：
    *   `/etc/orbit-deploy/`：用于存放配置文件。
    *   `/var/lib/orbit-deploy/`：用于存放数据，例如 SQLite 数据库文件。
3.  **安装二进制文件**：将 `orbit-deploy` 可执行文件复制到 `/usr/local/bin/`。
4.  **创建配置文件**：在 `/etc/orbit-deploy/` 目录下创建一个默认的 `config.toml` 配置文件。
5.  **创建并启动 systemd 服务**：设置一个名为 `orbit-deploy.service` 的 systemd 服务，并启动它，同时设置开机自启。

### 服务管理

安装完成后，您可以使用标准的 `systemctl` 命令来管理 OrbitDeploy 服务。

*   **查看服务状态**：

    ```bash
    sudo systemctl status orbit-deploy
    ```

*   **查看实时日志**：

    ```bash
    sudo journalctl -u orbit-deploy -f
    ```

*   **重启服务**：

    ```bash
    sudo systemctl restart orbit-deploy
    ```

## 2. 配置

当使用 `install.sh` 脚本安装时，推荐的配置方式是**直接修改 TOML 配置文件**。

**配置文件路径**：`/etc/orbit-deploy/config.toml`

每次修改完配置文件后，您都需要**重启服务**来让新的配置生效。

```bash
sudo systemctl restart orbit-deploy
```

### 配置项说明

以下是 `config.toml` 文件中可用的配置项。这些配置项也同样可以通过**环境变量**进行设置（环境变量的优先级更高），这在 Docker 部署或开发环境中非常有用。

#### 基础应用配置

| TOML 键 (`config.toml`) | 环境变量 | 描述 | 默认值 |
| :--- | :--- | :--- | :--- |
| `server_addr` | `SERVER_ADDR` | 应用服务的监听地址和端口。 | `:8285` |
| `db_path` | `DB_PATH` | SQLite 数据库文件的存储路径。 | `/var/lib/orbit-deploy/orbit_deploy.db` |
| `log_level` | `LOG_LEVEL` | 日志级别，可选值为 `debug`, `info`, `warn`, `error`。 | `info` |

#### 安全与 JWT 配置

这些是与安全相关的关键配置，**只能通过环境变量设置**，强烈建议您在生产环境中修改它们。

| 环境变量 | 描述 | 默认值 (不安全) |
| :--- | :--- | :--- |
| `ORBIT_ENCRYPTION_KEY` | 用于加密敏感数据（如凭证）的主密钥。**生产环境必须修改**。 | 一个固定的开发密钥 |
| `JWT_ACCESS_SECRET` | 用于签发访问令牌（Access Token）的密钥。 | `access-secret-key-change-in-production` |
| `JWT_REFRESH_SECRET` | 用于签发刷新令牌（Refresh Token）的密钥。 | `refresh-secret-key-change-in-production` |
| `JWT_ACCESS_TTL` | 访问令牌的有效时间。格式为持续时间字符串。 | `15m` (15分钟) |
| `JWT_REFRESH_TTL` | 刷新令牌的有效时间。 | `720h` (30天) |
| `JWT_ISSUER` | JWT 的签发者声明。 | `go-webui` |
| `JWT_AUDIENCE` | JWT 的受众声明。 | `go-webui-users` |

### 生产环境配置示例

以下是一个安全的 `config.toml` 配置文件示例。对于密钥字段，建议使用 `openssl rand -base64 32` 等命令生成强随机字符串。

```toml
# /etc/orbit-deploy/config.toml

# 基础配置
server_addr = ":8285"
db_path = "/var/lib/orbit-deploy/orbit_deploy.db"
log_level = "info"
```

**安全与 JWT 配置（通过环境变量设置）**

为了安全起见，**强烈建议您通过 systemd service 文件或 shell 环境来设置以下环境变量**，而不是将它们直接写入 `config.toml`。

```bash
# /etc/systemd/system/orbit-deploy.service (示例)

[Service]
Environment="ORBIT_ENCRYPTION_KEY=您生成的强随机字符串"
Environment="JWT_ACCESS_SECRET=另一个您生成的强随机字符串"
Environment="JWT_REFRESH_SECRET=再一个您生成的强随机字符串"
Environment="JWT_ACCESS_TTL=30m"
Environment="JWT_REFRESH_TTL=168h" # 7天
# ... 其他配置
```

编辑并保存文件后，请务必重启服务：

```bash
sudo systemctl daemon-reload
sudo systemctl restart orbit-deploy
```
