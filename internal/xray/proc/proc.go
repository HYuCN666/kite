package proc

import (
	"fmt"
	"os/exec"
	"runtime"
)

// Proc 管理 Xray 进程生命周期。
type Proc struct {
	binPath string
}

// New 创建进程管理器。
func New(binPath string) *Proc {
	return &Proc{binPath: binPath}
}

// Start 启动进程。
func (p *Proc) Start() error {
	return p.run("start")
}

// Stop 停止进程。
func (p *Proc) Stop() error {
	return p.run("stop")
}

// Restart 重启进程。
func (p *Proc) Restart() error {
	return p.run("restart")
}

// Reload 发送 SIGHUP 触发热重载，不中断现有连接。
func (p *Proc) Reload() error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("process control only supported on linux")
	}
	return exec.Command("systemctl", "kill", "-s", "HUP", "xray").Run()
}

// IsRunning 返回进程是否在运行。
func (p *Proc) IsRunning() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	err := exec.Command("systemctl", "is-active", "--quiet", "xray").Run()
	return err == nil
}

func (p *Proc) run(action string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("process control only supported on linux")
	}
	return exec.Command("systemctl", action, "xray").Run()
}
