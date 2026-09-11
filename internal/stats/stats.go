package stats

// Collector 负责从 Xray Stats API 采集并聚合流量数据。
type Collector struct {
	addr string
}

// NewCollector 创建采集器。
func NewCollector(addr string) *Collector {
	return &Collector{addr: addr}
}

// Snapshot 表示一次采集到的用户流量增量。
type Snapshot struct {
	Email    string
	Uplink   int64
	Downlink int64
}

// Collect 拉取一次流量快照。
func (c *Collector) Collect() ([]Snapshot, error) {
	return nil, nil
}
