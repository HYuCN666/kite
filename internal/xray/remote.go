package xray

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/HYuCN666/volans/internal/sshx"
	"github.com/HYuCN666/volans/internal/xray/config"
)

// Remote 通过 SSH 管理远端 Xray 节点。
type Remote struct {
	client     *sshx.Client
	binPath    string
	configPath string
}

// NewRemote 创建远端管理器。
func NewRemote(client *sshx.Client, binPath, configPath string) *Remote {
	return &Remote{client: client, binPath: binPath, configPath: configPath}
}

// Close 关闭底层连接。
func (r *Remote) Close() error { return r.client.Close() }

// Installed 返回远端内核是否已安装。
func (r *Remote) Installed() bool {
	out, err := r.client.Run(fmt.Sprintf("test -f %s && echo yes", r.binPath))
	return err == nil && strings.Contains(out, "yes")
}

// Version 返回远端内核版本。
func (r *Remote) Version() (string, error) {
	out, err := r.client.Run(r.binPath + " version")
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`(?i)xray\s+([0-9]+\.[0-9]+\.[0-9]+)`)
	if m := re.FindStringSubmatch(out); m != nil {
		return m[1], nil
	}
	return "", fmt.Errorf("cannot parse version from %q", out)
}

// Running 返回远端进程是否运行。
func (r *Remote) Running() bool {
	out, err := r.client.Run("systemctl is-active --quiet xray && echo yes")
	return err == nil && strings.Contains(out, "yes")
}

// Start 启动远端进程。
func (r *Remote) Start() error { _, err := r.client.Run("systemctl start xray"); return err }

// Stop 停止远端进程。
func (r *Remote) Stop() error { _, err := r.client.Run("systemctl stop xray"); return err }

// Restart 重启远端进程。
func (r *Remote) Restart() error { _, err := r.client.Run("systemctl restart xray"); return err }

// Reload 热重载远端进程。
func (r *Remote) Reload() error {
	_, err := r.client.Run("systemctl kill -s HUP xray")
	return err
}

// Install 在远端下载并安装 Xray。
func (r *Remote) Install() error {
	asset, err := r.remoteAssetName()
	if err != nil {
		return err
	}
	url := fmt.Sprintf("https://github.com/XTLS/Xray-core/releases/latest/download/%s", asset)
	cmd := fmt.Sprintf(
		"mkdir -p $(dirname %s) && curl -L -o /tmp/xray.zip %s && unzip -o /tmp/xray.zip -d /tmp/xray-tmp && install -m 755 /tmp/xray-tmp/xray %s",
		r.binPath, url, r.binPath,
	)
	if _, err := r.client.Run(cmd); err != nil {
		return err
	}

	unit := fmt.Sprintf(`[Unit]
Description=Xray Service
After=network.target

[Service]
Type=simple
ExecStart=%s run -config %s
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
`, r.binPath, r.configPath)
	if err := r.client.WriteFile("/etc/systemd/system/xray.service", []byte(unit)); err != nil {
		return err
	}
	_, err = r.client.Run("systemctl daemon-reload && systemctl enable xray")
	return err
}

// WriteConfig 写入远端配置并热重载。
func (r *Remote) WriteConfig(cfg *config.Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := r.client.WriteFile(r.configPath, data); err != nil {
		return err
	}
	if r.Running() {
		return r.Reload()
	}
	return r.Start()
}

func (r *Remote) remoteAssetName() (string, error) {
	out, err := r.client.Run("uname -m")
	if err != nil {
		return "", err
	}
	switch strings.TrimSpace(out) {
	case "x86_64", "amd64":
		return "Xray-linux-64.zip", nil
	case "aarch64", "arm64":
		return "Xray-linux-arm64-v8a.zip", nil
	default:
		return "", fmt.Errorf("unsupported remote arch: %s", out)
	}
}
