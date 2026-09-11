package stats

import (
	"context"
	"strings"
	"time"

	statscmd "github.com/xtls/xray-core/app/stats/command"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Collector 从 Xray Stats API 采集流量数据。
type Collector struct {
	addr   string
	conn   *grpc.ClientConn
	client statscmd.StatsServiceClient
}

// Snapshot 表示一次采集到的用户流量增量。
type Snapshot struct {
	Email    string
	Uplink   int64
	Downlink int64
}

// NewCollector 创建采集器。
func NewCollector(addr string) *Collector {
	return &Collector{addr: addr}
}

// Connect 建立 gRPC 连接。
func (c *Collector) Connect() error {
	if c.conn != nil {
		return nil
	}
	conn, err := grpc.Dial(c.addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return err
	}
	c.conn = conn
	c.client = statscmd.NewStatsServiceClient(conn)
	return nil
}

// Close 关闭连接。
func (c *Collector) Close() error {
	if c.conn == nil {
		return nil
	}
	err := c.conn.Close()
	c.conn = nil
	return err
}

// Collect 拉取一次流量增量快照（读取后重置）。
func (c *Collector) Collect() ([]Snapshot, error) {
	if err := c.Connect(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.QueryStats(ctx, &statscmd.QueryStatsRequest{Pattern: "user>>>", Reset_: true})
	if err != nil {
		return nil, err
	}

	byEmail := map[string]*Snapshot{}
	for _, st := range resp.GetStat() {
		email, dir, ok := parseName(st.GetName())
		if !ok {
			continue
		}
		snap := byEmail[email]
		if snap == nil {
			snap = &Snapshot{Email: email}
			byEmail[email] = snap
		}
		if dir == "uplink" {
			snap.Uplink = st.GetValue()
		} else {
			snap.Downlink = st.GetValue()
		}
	}

	out := make([]Snapshot, 0, len(byEmail))
	for _, s := range byEmail {
		out = append(out, *s)
	}
	return out, nil
}

// parseName 解析形如 user>>>[email]>>>traffic>>>uplink 的计数名。
func parseName(name string) (email, dir string, ok bool) {
	parts := strings.Split(name, ">>>")
	if len(parts) != 4 || parts[0] != "user" || parts[2] != "traffic" {
		return "", "", false
	}
	if parts[3] != "uplink" && parts[3] != "downlink" {
		return "", "", false
	}
	return parts[1], parts[3], true
}

// OnlineDevices 返回某用户当前在线设备（独立 IP）数量。
func (c *Collector) OnlineDevices(email string) (int, error) {
	if err := c.Connect(); err != nil {
		return 0, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.GetStatsOnlineIpList(ctx, &statscmd.GetStatsRequest{Name: email})
	if err != nil {
		// 用户不在线时返回 NotFound，视为 0 设备。
		return 0, nil
	}
	return len(resp.GetIps()), nil
}
