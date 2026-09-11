package system

import (
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
)

// Info 为系统资源使用信息。
type Info struct {
	CPU  float64 `json:"cpu"`
	Mem  float64 `json:"mem"`
	Disk float64 `json:"disk"`
	Load float64 `json:"load"`
}

// Collect 采集当前系统资源使用率。
func Collect() Info {
	var info Info
	if per, err := cpu.Percent(0, false); err == nil && len(per) > 0 {
		info.CPU = per[0]
	}
	if vm, err := mem.VirtualMemory(); err == nil {
		info.Mem = vm.UsedPercent
	}
	if du, err := disk.Usage("/"); err == nil {
		info.Disk = du.UsedPercent
	}
	if la, err := load.Avg(); err == nil {
		info.Load = la.Load1
	}
	return info
}
