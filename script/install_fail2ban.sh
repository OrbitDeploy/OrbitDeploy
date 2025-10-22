#!/bin/bash

# ==============================================================================
# Fail2Ban Installer and Basic Configuration Script
#
# Author: Gemini (as a Professional Linux Sysadmin)
# Description: This script automatically detects the Linux distribution
#              (supporting DNF, Pacman, APT systems) and installs Fail2Ban.
#              It then creates a basic 'jail.local' to protect SSHD,
#              following best practices.
#
# Supported Systems:
#   - DNF-based: Fedora, CentOS Stream, RHEL, Rocky Linux, AlmaLinux
#   - Pacman-based: Arch Linux
#   - APT-based: Ubuntu, Debian
# ==============================================================================

# --- Function to print colored messages ---
print_msg() {
    local color="$1"
    local message="$2"
    case "$color" in
        "green") echo -e "\e[32m${message}\e[0m" ;;
        "red") echo -e "\e[31m${message}\e[0m" ;;
        "yellow") echo -e "\e[33m${message}\e[0m" ;;
        *) echo "${message}" ;;
    esac
}

# --- 1. Check for Root Privileges ---
if [ "$(id -u)" -ne 0 ]; then
    print_msg "red" "错误：此脚本必须以 root 权限运行。"
    print_msg "red" "请尝试使用 'sudo ./install_fail2ban.sh' 运行。"
    exit 1
fi

# --- 2. Detect Package Manager and OS ---
PKG_MANAGER=""
OS_ID=""

if [ -f /etc/os-release ]; then
    # shellcheck source=/dev/null
    source /etc/os-release
    OS_ID=$ID
else
    print_msg "red" "无法检测到操作系统类型，/etc/os-release 文件不存在。"
    exit 1
fi

print_msg "green" "检测到操作系统: $OS_ID"

case "$OS_ID" in
    fedora|centos|rhel|rocky|almalinux)
        PKG_MANAGER="dnf"
        ;;
    arch)
        PKG_MANAGER="pacman"
        ;;
    ubuntu|debian)
        PKG_MANAGER="apt"
        ;;
    *)
        print_msg "red" "错误：不支持的操作系统 '$OS_ID'。"
        print_msg "red" "此脚本仅支持基于 DNF, Pacman, APT 的系统。"
        exit 1
        ;;
esac

print_msg "green" "使用的包管理器: $PKG_MANAGER"


# --- 3. Install Fail2Ban ---
print_msg "yellow" "正在安装 Fail2Ban..."

case "$PKG_MANAGER" in
    "dnf")
        dnf install -y fail2ban
        ;;
    "pacman")
        pacman -Syu --noconfirm fail2ban
        ;;
    "apt")
        apt-get update
        apt-get install -y fail2ban
        ;;
esac

# Verify installation
if ! command -v fail2ban-client &> /dev/null; then
    print_msg "red" "Fail2Ban 安装失败，请检查上面的错误信息。"
    exit 1
fi

print_msg "green" "Fail2Ban 安装成功！"


# --- 4. Configure Fail2Ban (Best Practice: Create jail.local) ---
print_msg "yellow" "正在配置 '/etc/fail2ban/jail.local'..."

JAIL_LOCAL_FILE="/etc/fail2ban/jail.local"

# Create the jail.local file using a heredoc.
# This configuration will:
# - Set default ban time to 1 hour.
# - Ban IPs after 5 failed attempts within a 10-minute window.
# - Enable SSHD protection.
cat > "$JAIL_LOCAL_FILE" << EOF
# ==============================================================================
# Fail2Ban LOCAL configuration file.
#
# Created by install_fail2ban.sh script.
#
# NOTE: This file overrides settings in /etc/fail2ban/jail.conf.
#       Do NOT edit jail.conf, as it may be overwritten by package updates.
#       All your custom changes should be placed in this file.
# ==============================================================================

[DEFAULT]
# Ban time in seconds. 1h = 3600, 1d = 86400.
bantime = 1h

# The time window (in seconds) for detecting attacks.
findtime = 10m

# Number of failures before an IP is banned.
maxretry = 5

# Whitelist your own IP addresses to avoid getting locked out.
# Separate multiple IPs with spaces.
# Example: ignoreip = 127.0.0.1/8 192.168.1.100
ignoreip = 127.0.0.1/8

[sshd]
# Enable the SSHD jail.
enabled = true

# If you use a non-standard SSH port, change it here.
# Example: port = 2222
# Or for multiple ports: port = ssh,2222
# port = ssh

# Action to take. default is iptables-multiport
# action = %(action_)s
EOF

print_msg "green" "'$JAIL_LOCAL_FILE' 创建成功。"
print_msg "yellow" "重要提示: 如果您有固定的公网IP，请编辑 '$JAIL_LOCAL_FILE' 文件，将您的IP添加到 'ignoreip' 列表中以防被误封。"


# --- 5. Enable and Start the Fail2Ban Service ---
print_msg "yellow" "正在启用并启动 Fail2Ban 服务..."

systemctl enable fail2ban
systemctl restart fail2ban

# --- 6. Final Verification ---
if systemctl is-active --quiet fail2ban; then
    print_msg "green" "✅ Fail2Ban 服务已成功启动并设置为开机自启。"
    echo "-----------------------------------------------------"
    print_msg "yellow" "您可以使用以下命令来检查 SSHD 防护状态:"
    echo "  sudo fail2ban-client status sshd"
    echo "-----------------------------------------------------"
else
    print_msg "red" "❌ 错误：Fail2Ban 服务启动失败。"
    print_msg "red" "请使用 'systemctl status fail2ban' 和 'journalctl -u fail2ban' 命令检查日志以排查问题。"
fi

exit 0