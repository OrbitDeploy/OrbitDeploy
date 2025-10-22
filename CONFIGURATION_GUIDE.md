 # OrbitDeploy Configuration Guide

This document will guide you through the installation and configuration of the OrbitDeploy application.

## 1. Installation

We provide a convenient `install.sh` script to install OrbitDeploy as a systemd service. Before running the installation script, please ensure you have successfully compiled the `orbit-deploy` binary according to the [Build Instructions](README.md#building-from-source).

Run the installation script as the root user:

```bash
sudo ./install.sh
```

This script will automatically perform the following tasks:

1.  **Create Service User**: Creates a system user named `orbit-deploy` to securely run the service.
2.  **Create Directories**:
    *   `/etc/orbit-deploy/`: For storing configuration files.
    *   `/var/lib/orbit-deploy/`: For storing data, such as the SQLite database file.
3.  **Install Binary**: Copies the `orbit-deploy` executable to `/usr/local/bin/`.
4.  **Create Configuration File**: Creates a default `config.toml` configuration file in the `/etc/orbit-deploy/` directory.
5.  **Create and Start Systemd Service**: Sets up a systemd service named `orbit-deploy.service`, starts it, and enables it to run on boot.

### Service Management

After installation, you can use standard `systemctl` commands to manage the OrbitDeploy service.

*   **Check Service Status**:

    ```bash
    sudo systemctl status orbit-deploy
    ```

*   **View Real-time Logs**:

    ```bash
    sudo journalctl -u orbit-deploy -f
    ```

*   **Restart Service**:

    ```bash
    sudo systemctl restart orbit-deploy
    ```

## 2. Configuration

When installing with the `install.sh` script, the recommended way to configure the application is by **directly modifying the TOML configuration file**.

**Configuration File Path**: `/etc/orbit-deploy/config.toml`

After modifying the configuration file, you must **restart the service** for the new settings to take effect.

```bash
sudo systemctl restart orbit-deploy
```

### Configuration Options

The following options are available in the `config.toml` file. Note that some settings, particularly for security, must be set via environment variables.

#### Basic Application Configuration

These settings can be configured in `config.toml`.

| TOML Key (`config.toml`) | Environment Variable | Description | Default Value |
| :--- | :--- | :--- | :--- |
| `server_addr` | `SERVER_ADDR` | The listening address and port for the application service. | `:8285` |
| `db_path` | `DB_PATH` | The storage path for the SQLite database file. | `/var/lib/orbit-deploy/orbit_deploy.db` |
| `log_level` | `LOG_LEVEL` | The logging level. Options are `debug`, `info`, `warn`, `error`. | `info` |

#### Security and JWT Configuration

These are critical security-related settings. They **must be set as environment variables** and should be changed for any production environment.

| Environment Variable | Description | Default (Insecure) Value |
| :--- | :--- | :--- |
| `ORBIT_ENCRYPTION_KEY` | The master key for encrypting sensitive data (like credentials). **Must be changed in production**. | A fixed development key |
| `JWT_ACCESS_SECRET` | The secret key for signing access tokens. | `access-secret-key-change-in-production` |
| `JWT_REFRESH_SECRET` | The secret key for signing refresh tokens. | `refresh-secret-key-change-in-production` |
| `JWT_ACCESS_TTL` | The time-to-live for access tokens. Format is a duration string. | `15m` (15 minutes) |
| `JWT_REFRESH_TTL` | The time-to-live for refresh tokens. | `720h` (30 days) |
| `JWT_ISSUER` | The issuer claim for JWTs. | `go-webui` |
| `JWT_AUDIENCE` | The audience claim for JWTs. | `go-webui-users` |

### Production Environment Example

Below is an example of a secure configuration. For the secret keys, it is recommended to generate strong random strings using a command like `openssl rand -base64 32`.

**1. Basic Configuration (`config.toml`)**

```toml
# /etc/orbit-deploy/config.toml

# Basic Configuration
server_addr = ":8285"
db_path = "/var/lib/orbit-deploy/orbit_deploy.db"
log_level = "info"
```

**2. Security & JWT Configuration (Environment Variables)**

For security, it is **strongly recommended to set these environment variables within the systemd service file** rather than exporting them globally.

To do this, edit the service file:
```bash
sudo systemctl edit orbit-deploy.service
```

And add the `[Service]` section with your environment variables:

```ini
# /etc/systemd/system/orbit-deploy.service.d/override.conf

[Service]
Environment="ORBIT_ENCRYPTION_KEY=your_strong_random_string_here"
Environment="JWT_ACCESS_SECRET=another_strong_random_string_here"
Environment="JWT_REFRESH_SECRET=a_third_strong_random_string_here"
Environment="JWT_ACCESS_TTL=30m"
Environment="JWT_REFRESH_TTL=168h" # 7 days
```

After creating or editing this file, reload systemd and restart the service to apply all changes:

```bash
sudo systemctl daemon-reload
sudo systemctl restart orbit-deploy
```
