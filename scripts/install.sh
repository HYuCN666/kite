#!/usr/bin/env bash
set -euo pipefail

# Kite 一键安装脚本（Ubuntu / Debian）
# 用法：bash install.sh

ARCH="$(uname -m)"
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  *) echo "不支持的架构: $ARCH" >&2; exit 1 ;;
esac

if [[ "$(id -u)" -ne 0 ]]; then
  echo "请以 root 权限运行" >&2
  exit 1
fi

INSTALL_DIR="/opt/kite"
DATA_DIR="/etc/kite"

echo "==> 创建目录"
mkdir -p "$INSTALL_DIR" "$DATA_DIR"

echo "==> 下载 Xray"
XRAY_URL="https://github.com/XTLS/Xray-core/releases/latest/download/Xray-linux-${ARCH}.zip"
curl -L -o /tmp/xray.zip "$XRAY_URL"
unzip -o /tmp/xray.zip -d /tmp/xray
install -m 755 /tmp/xray/xray /usr/local/bin/xray

echo "==> 安装 Kite"
# 此处替换为实际的 release 下载地址
# curl -L -o "$INSTALL_DIR/kite" "https://github.com/HYuCN666/kite/releases/latest/download/kite-linux-${ARCH}"
# install -m 755 "$INSTALL_DIR/kite" /usr/local/bin/kite

echo "==> 创建 systemd 服务"
cat > /etc/systemd/system/kite.service <<EOF
[Unit]
Description=Kite
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/kite --data "$DATA_DIR"
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable kite
systemctl start kite

echo "==> 安装完成，面板地址: http://127.0.0.1:8080"
