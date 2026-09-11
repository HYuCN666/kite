package xray

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/HYuCN666/volans/internal/xray/config"
	"github.com/HYuCN666/volans/internal/xray/proc"
)

// Manager 负责 Xray 内核的安装、进程控制与配置生成。
type Manager struct {
	binPath    string
	configPath string
	proc       *proc.Proc
}

// NewManager 创建内核管理器。
func NewManager(binPath, dataDir string) *Manager {
	return &Manager{
		binPath:    binPath,
		configPath: filepath.Join(dataDir, "xray", "config.json"),
		proc:       proc.New(binPath),
	}
}

// Installed 返回内核是否已安装。
func (m *Manager) Installed() bool {
	_, err := os.Stat(m.binPath)
	return err == nil
}

// Start 启动进程。
func (m *Manager) Start() error {
	return m.proc.Start()
}

// Stop 停止进程。
func (m *Manager) Stop() error {
	return m.proc.Stop()
}

// Restart 重启进程。
func (m *Manager) Restart() error {
	return m.proc.Restart()
}

// Reload 触发热重载，不中断现有连接。
func (m *Manager) Reload() error {
	return m.proc.Reload()
}

// Running 返回进程是否在运行。
func (m *Manager) Running() bool {
	return m.proc.IsRunning()
}

// WriteConfig 将配置写入文件，并在内核已安装时触发热重载。
func (m *Manager) WriteConfig(cfg *config.Config) error {
	if err := os.MkdirAll(filepath.Dir(m.configPath), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(m.configPath, data, 0o600); err != nil {
		return err
	}
	if !m.Installed() {
		return nil
	}
	if m.Running() {
		return m.Reload()
	}
	return m.Start()
}
