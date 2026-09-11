#!/usr/bin/env bash
set -euo pipefail

# Volans 一键安装脚本（Ubuntu / Debian）
# 用法：curl -fsSL https://raw.githubusercontent.com/HYuCN666/volans/main/install.sh | bash

REPO="HYuCN666/volans"

if [[ "$(id -u)" -ne 0 ]]; then
  echo "请以 root 权限运行: sudo bash install.sh" >&2
  exit 1
fi

ARCH="$(uname -m)"
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  *) echo "不支持的架构: $ARCH" >&2; exit 1 ;;
esac

INSTALL_DIR="/opt/volans"
DATA_DIR="/etc/volans"

echo "==> 下载 Volans 发行版"
PKG="volans-linux-${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/latest/download/${PKG}"
curl -L --fail -o "/tmp/${PKG}" "$URL"

echo "==> 安装到 ${INSTALL_DIR}"
mkdir -p "$INSTALL_DIR"
tar -xzf "/tmp/${PKG}" -C "$INSTALL_DIR"
ln -sf "${INSTALL_DIR}/volans/volans" /usr/local/bin/volans

echo "==> 创建 systemd 服务"
cat > /etc/systemd/system/volans.service <<EOF
[Unit]
Description=Volans
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/volans --bind 0.0.0.0 --port 8080 --data ${DATA_DIR} --web ${INSTALL_DIR}/volans/web/dist
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now volans

echo ""
echo "==> 安装完成"
echo "    面板地址: http://<服务器IP>:8080"
echo "    默认账号: admin / admin（登录后请立即修改密码）"
echo "    Xray 内核请在面板「入站节点」页内一键安装"
