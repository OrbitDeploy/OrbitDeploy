#!/bin/bash

# This script installs the OrbitDeploy application as a systemd service.

# --- Configuration ---
BINARY_NAME="orbit-deploy"
INSTALL_PATH="/usr/local/bin"
SERVICE_NAME="orbit-deploy"
SERVICE_USER="orbit-deploy"
CONFIG_DIR="/etc/orbit-deploy"
DATA_DIR="/var/lib/orbit-deploy"

# --- Pre-flight Checks ---
if [ "$(id -u)" -ne 0 ]; then
  echo "This script must be run as root. Please use sudo." >&2
  exit 1
fi

if ! command -v systemctl &> /dev/null; then
  echo "systemd is not available on this system. This script requires systemd." >&2
  exit 1
fi

if [ ! -f "./${BINARY_NAME}" ]; then
  echo "The binary '${BINARY_NAME}' was not found in the current directory." >&2
  echo "Please build the application first and run this script from the same directory." >&2
  exit 1
fi

# --- Installation ---
echo "Installing OrbitDeploy..."

# 1. Create user and groups
echo "- Creating user '${SERVICE_USER}'..."
if ! id "${SERVICE_USER}" &>/dev/null; then
  useradd -r -s /bin/false "${SERVICE_USER}"
else
  echo "  User '${SERVICE_USER}' already exists. Skipping."
fi

# 2. Create directories
echo "- Creating directories..."
mkdir -p "${CONFIG_DIR}"
mkdir -p "${DATA_DIR}"
chown -R "${SERVICE_USER}:${SERVICE_USER}" "${CONFIG_DIR}"
chown -R "${SERVICE_USER}:${SERVICE_USER}" "${DATA_DIR}"

# 3. Install binary
echo "- Installing binary to ${INSTALL_PATH}/${BINARY_NAME}..."
cp "./${BINARY_NAME}" "${INSTALL_PATH}/${BINARY_NAME}"
chmod +x "${INSTALL_PATH}/${BINARY_NAME}"

# 4. Create systemd service file
echo "- Creating systemd service file..."
cat > "/etc/systemd/system/${SERVICE_NAME}.service" << EOL
[Unit]
Description=OrbitDeploy Service
After=network.target

[Service]
Type=simple
User=${SERVICE_USER}
Group=${SERVICE_USER}
WorkingDirectory=${DATA_DIR}
ExecStart=${INSTALL_PATH}/${BINARY_NAME} -config ${CONFIG_DIR}/config.toml
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOL

# 5. Create a default config file if it doesn't exist
if [ ! -f "${CONFIG_DIR}/config.toml" ]; then
  echo "- Creating default config file at ${CONFIG_DIR}/config.toml..."
  cat > "${CONFIG_DIR}/config.toml" << EOL
# OrbitDeploy Configuration

# Server address to listen on
server_addr = ":8285"

# Path to the SQLite database
db_path = "${DATA_DIR}/orbit_deploy.db"

# Log level (debug, info, warn, error)
log_level = "info"
EOL
  chown "${SERVICE_USER}:${SERVICE_USER}" "${CONFIG_DIR}/config.toml"
fi

# 6. Reload systemd, enable and start the service
echo "- Reloading systemd and starting service..."
systemctl daemon-reload
systemctl enable "${SERVICE_NAME}.service"
systemctl start "${SERVICE_NAME}.service"

# --- Post-installation ---
echo ""
echo "Installation complete!"
echo ""
echo "To check the status of the service, run:"
echo "  systemctl status ${SERVICE_NAME}"
echo ""
echo "To view the logs, run:"
echo "  journalctl -u ${SERVICE_NAME} -f"
echo ""
echo "The configuration file is located at: ${CONFIG_DIR}/config.toml"
