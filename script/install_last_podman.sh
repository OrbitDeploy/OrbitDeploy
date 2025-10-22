#!/bin/bash

# 获取操作系统信息
# 检查 /etc/os-release 文件是否存在
if [ -f /etc/os-release ]; then
    # 使用 source 命令加载文件中的变量
    # 这会比多次调用 awk 更高效
    source /etc/os-release
else
    echo "错误：无法找到 /etc/os-release 文件。"
    # 如果文件不存在，可以尝试其他方法或直接退出
    # 例如，尝试 lsb_release
    if command -v lsb_release &> /dev/null; then
        ID=$(lsb_release -is | tr '[:upper:]' '[:lower:]') # 转为小写以匹配 os-release 风格
        VERSION_ID=$(lsb_release -rs)
    else
        echo "无法确定操作系统版本，请手动检查。"
        exit 1
    fi
fi

echo "---"
echo "正在检测操作系统..."
# 直接使用加载后的变量 $ID 和 $VERSION_ID
echo "ID: $ID"
echo "VERSION_ID: $VERSION_ID"
echo "---"


# 检查是否为 Debian 13+ 或 Ubuntu 25.04+
if { [ "$ID" = "debian" ] && [ "$VERSION_ID" -ge 13 ]; } || \
   { [ "$ID" = "ubuntu" ] && [ "$(printf '%s\n' "$VERSION_ID" "25.04" | sort -V | tail -n 1)" = "$VERSION_ID" ]; }; then
    
    # 这里是条件满足时要执行的代码
    echo "操作系统版本符合要求 (Debian 13+ 或 Ubuntu 25.04+)"

    echo "检测到 Debian 13+ 或 Ubuntu 25.04+。"
    echo "将直接从官方软件源安装 Podman..."
    echo "---"

    # 更新软件包列表并安装必要的工具
    apt-get update
    apt-get install -y curl wget git nano unzip

    # 安装 Podman 相关组件
    echo "---"
    echo "正在安装 Podman 相关组件..."
    apt-get upgrade -y
    apt-get install -y podman  netavark aardvark-dns 

elif [ "$ID" = "ubuntu" ] && [ "$VERSION_ID" = "24.04" ]; then
    echo "检测到 Ubuntu 24.04，将通过 Plucky 存储库安装较新版本的 Podman..."
    echo "---"

    # 定义文件路径
    PINNING_FILE="/etc/apt/preferences.d/podman-plucky.pref"
    SOURCE_LIST="/etc/apt/sources.list.d/plucky.list"

    # 更新软件包列表并安装必要的工具
    apt-get update
    apt-get install -y curl wget git nano 

    # 写入 Plucky APT 来源列表
    echo "添加 Plucky 存储库到 $SOURCE_LIST..."
    echo "deb http://archive.ubuntu.com/ubuntu plucky main universe" > "$SOURCE_LIST"

    # 写入 APT 固定规则
    echo "写入 APT 固定规则到 $PINNING_FILE..."
    cat <<EOF > "$PINNING_FILE"
Package: podman buildah golang-github-containers-common crun libgpgme11t64 libgpg-error0 golang-github-containers-image catatonit conmon containers-storage
Pin: release n=plucky
Pin-Priority: 991

Package: libsubid4 netavark passt aardvark-dns containernetworking-plugins libslirp0 slirp4netns
Pin: release n=plucky
Pin-Priority: 991

Package: *
Pin: release n=plucky
Pin-Priority: 400
EOF

    # 更新 APT 缓存
    echo "---"
    echo "更新 APT 软件包列表..."
    apt-get update

    # 安装 Podman 相关组件
    echo "---"
    echo "正在安装 Podman 相关组件..."
    apt-get install -y podman netavark aardvark-dns

elif [ "$ID" = "fedora" ] || [ "$ID" = "rhel" ] || [ "$ID" = "almalinux" ] || [ "$ID" = "rocky" ] || [ "$ID_LIKE" = "*rhel*" ] || [ "$ID_LIKE" = "*fedora*" ]; then
    echo "检测到基于 DNF 的发行版: $NAME $VERSION_ID"
    echo "将从官方软件源安装 Podman..."
    echo "---"

    # 更新系统并安装必要的工具
    dnf update -y
    dnf install -y curl wget git nano unzip

    # 安装 Podman 相关组件
    echo "---"
    echo "正在安装 Podman 相关组件..."
    dnf install -y podman netavark aardvark-dns

elif [ "$ID" = "arch" ] || [ "$ID_LIKE" = "arch" ]; then
    echo "检测到 Arch Linux 系统。"
    echo "将从官方软件源安装 Podman 5.0..."
    echo "---"

    # 更新系统并安装必要的工具
    pacman -Syu --noconfirm
    pacman -S --noconfirm curl wget git nano unzip

    # 安装 Podman 相关组件
    echo "---"
    echo "正在安装 Podman 相关组件..."
    pacman -S --noconfirm podman netavark aardvark-dns

else

    # 检查是否为 Debian 12
    if [ "$ID" = "debian" ] && [ "$VERSION_ID" = "12" ]; then
        echo "检测到 Debian 12，将通过第三方存储库安装 Podman..."
        echo "---"

        # 更新软件包列表并安装必要的工具
        echo "---"
        echo "# 更新软件包列表并安装必要的工具..."
        apt-get update
        apt-get install -y curl wget git gpg gnupg2 software-properties-common apt-transport-https lsb-release ca-certificates bc

        # 下载并导入存储库的 GPG 密钥
        echo "---"
        echo "# 下载并导入存储库的 GPG 密钥..."
        wget http://downloadcontent.opensuse.org/repositories/home:/alvistack/Debian_12/Release.key -O alvistack_key
        cat alvistack_key | gpg --dearmor | tee /etc/apt/trusted.gpg.d/alvistack.gpg >/dev/null

        # 将存储库添加到系统
        echo "---"
        echo "# 将存储库添加到系统..."
        echo "deb http://downloadcontent.opensuse.org/repositories/home:/alvistack/Debian_12/ /" | tee /etc/apt/sources.list.d/alvistack.list >/dev/null

        # 再次更新软件包列表并安装 Podman 相关组件
        echo "---"
        echo "# 再次更新软件包列表并安装 Podman 相关组件..."
        apt-get update && apt-get upgrade -y
        apt-get install -y podman netavark aardvark-dns criu curl wget git nano unzip
    else
        echo "检测到系统: $ID $VERSION_ID"
        echo "⚠️  此脚本仅支持 Debian 12、Debian 13+、Ubuntu 24.04、Ubuntu 25.04+、基于 DNF 的发行版 或 Arch Linux 的自动安装。"
        echo "请寻找其他方式安装最新版本 Podman（需 5.0.0 以上）。"
        echo "---"
        echo "建议："
        echo "1. 查看官方文档: https://podman.io/getting-started/installation"
        echo "2. 编译安装最新版本"
        echo "3. 使用 Snap/Flatpak 等包管理器"
        exit 1
    fi

fi

# 验证安装
echo "---
# 验证安装..."
podman version

echo "---
# Podman 已成功安装。"