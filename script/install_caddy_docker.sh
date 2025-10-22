#!/bin/bash

# ==============================================================================
# 脚本名称: install_caddy_docker.sh
# 描述:     使用 Podman + Quadlet 方式安装 Caddy Cloudflare 镜像
# 作者:     Auto-generated
# 日期:     2025-01-01
# ==============================================================================

# 在执行命令时输出，并在出错时立即退出
set -e
set -o pipefail

# --- 权限检查 ---
if [ "$(id -u)" -ne 0 ]; then
  echo "错误：此脚本必须以 root 权限运行。" >&2
  echo "请尝试使用 'sudo ./install_caddy_docker.sh' 命令运行。" >&2
  exit 1
fi

echo ">>> 步骤 1: 检查 Podman 是否已安装..."
if ! command -v podman &> /dev/null; then
    echo "错误：Podman 未安装。请先安装 Podman。" >&2
    exit 1
fi

echo "✅ Podman 已安装"
podman version --format "{{.Client.Version}}"

echo ""
echo ">>> 步骤 2: 创建 Caddy 系统目录..."

# 创建 Quadlet 配置目录
mkdir -p /etc/containers/systemd
mkdir -p /etc/caddy
mkdir -p /var/lib/caddy/data
mkdir -p /var/lib/caddy/config

# 设置权限（容器会以 caddy 用户运行）
chown -R 1000:1000 /var/lib/caddy
chmod 755 /etc/caddy

echo ""
echo ">>> 步骤 3: 创建 Caddy Quadlet 配置文件..."

# 创建 Caddy 容器的 Quadlet 配置
cat <<EOF > /etc/containers/systemd/caddy.container
[Unit]
Description=Caddy Web Server with Cloudflare DNS
Documentation=https://caddyserver.com/docs/
After=network-online.target
Wants=network-online.target

[Container]
Image=ghcr.io/caddybuilds/caddy-cloudflare:latest
ContainerName=caddy
AutoUpdate=registry

# 网络配置
PublishPort=80:80
PublishPort=443:443
PublishPort=443:443/udp

# 卷挂载
Volume=/etc/caddy:/etc/caddy:ro
Volume=/var/lib/caddy/data:/data
Volume=/var/lib/caddy/config:/config

# 环境变量 (如果需要 Cloudflare API Token)
# Environment=CLOUDFLARE_API_TOKEN=your_cloudflare_api_token

# 安全配置
AddCapability=NET_ADMIN

# 重启策略
Restart=unless-stopped

[Service]
Restart=always
TimeoutStopSec=30

[Install]
WantedBy=multi-user.target default.target
EOF

echo ""
echo ">>> 步骤 4: 创建默认 Caddyfile..."

if [ ! -f /etc/caddy/Caddyfile ]; then
cat <<EOF > /etc/caddy/Caddyfile
# /etc/caddy/Caddyfile
# Caddy 配置文件 - 支持 Cloudflare DNS-01 ACME 挑战
# 更多信息请访问：https://caddyserver.com/docs/caddyfile

{
	# 全局配置 - 使用 Cloudflare DNS 进行 ACME 挑战
	# 取消注释以下行并设置 CLOUDFLARE_API_TOKEN 环境变量
	# acme_dns cloudflare {env.CLOUDFLARE_API_TOKEN}
	
	# 信任 Cloudflare IP 范围
	servers {
		trusted_proxies cloudflare
		client_ip_headers Cf-Connecting-Ip
	}
}

# 默认响应
:80 {
	header Server "Caddy with Cloudflare DNS"
	respond "Hello from Caddy! Cloudflare DNS support enabled."
}

# 示例 HTTPS 站点配置（需要有效域名和 Cloudflare API Token）
# example.com {
#     root * /srv
#     file_server
#     encode gzip
#     
#     # 使用全局 DNS 挑战配置，无需在此重复
# }
EOF

chown root:root /etc/caddy/Caddyfile
chmod 644 /etc/caddy/Caddyfile
fi

echo ""
echo ">>> 步骤 5: 拉取 Caddy Cloudflare 镜像..."
podman pull ghcr.io/caddybuilds/caddy-cloudflare:latest

echo ""
echo ">>> 步骤 6: 重新加载 systemd 并启用 Caddy 服务..."
systemctl daemon-reload
systemctl enable --now caddy.service

echo ""
echo ">>> 步骤 7: 等待服务启动并验证..."
sleep 5

# 检查服务状态
if systemctl is-active --quiet caddy.service; then
    echo "✅ Caddy 服务已成功启动"
else
    echo "❌ Caddy 服务启动失败"
    systemctl status caddy.service --no-pager
    exit 1
fi

echo ""
echo ">>> 步骤 8: 验证容器运行状态..."
podman ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep caddy || echo "容器列表中未找到 caddy"

echo ""
echo -e "\033[32m安装完成！\033[0m"
echo "Caddy 已使用 Docker/Podman + Quadlet 方式安装完成。"
echo ""
echo "配置信息："
echo "- Quadlet 配置文件: /etc/containers/systemd/caddy.container"
echo "- Caddyfile 配置: /etc/caddy/Caddyfile"
echo "- 数据目录: /var/lib/caddy/data"
echo "- 配置目录: /var/lib/caddy/config"
echo ""
echo "使用说明："
echo "1. 修改 /etc/caddy/Caddyfile 以配置您的网站"
echo "2. 如需 Cloudflare DNS-01 支持，请:"
echo "   - 获取 Cloudflare API Token"
echo "   - 在 /etc/containers/systemd/caddy.container 中设置 CLOUDFLARE_API_TOKEN"
echo "   - 在 Caddyfile 中启用 acme_dns cloudflare"
echo "3. 重启服务: sudo systemctl restart caddy.service"
echo ""
echo "您可以通过访问 http://<您的服务器IP> 来测试默认页面。"

exit 0