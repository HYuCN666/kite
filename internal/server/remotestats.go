package server

import (
	"fmt"
	"sync"

	"github.com/HYuCN666/volans/internal/sshx"
	"github.com/HYuCN666/volans/internal/stats"
	"github.com/HYuCN666/volans/internal/store"
)

// remoteCollector 为某台远端服务器的统计采集器（含 SSH 隧道）。
type remoteCollector struct {
	ssh  *sshx.Client
	coll *stats.Collector
}

// remoteStats 管理各远端服务器的 stats 采集器缓存。
type remoteStats struct {
	mu    sync.Mutex
	nodes map[int64]*remoteCollector
}

func newRemoteStats() *remoteStats {
	return &remoteStats{nodes: make(map[int64]*remoteCollector)}
}

// get 懒创建并返回某服务器的采集器。
func (rs *remoteStats) get(db *store.Store, serverID int64) (*stats.Collector, error) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if n, ok := rs.nodes[serverID]; ok {
		return n.coll, nil
	}

	sv, err := db.GetServer(serverID)
	if err != nil || sv == nil {
		return nil, fmt.Errorf("server %d not found", serverID)
	}
	client := sshx.New(sshx.Config{
		Host:       sv.Host,
		Port:       sv.Port,
		Username:   sv.Username,
		AuthType:   sv.AuthType,
		Password:   sv.Password,
		PrivateKey: sv.PrivateKey,
	})
	addr, err := client.ForwardLocal(statsAPIAddr)
	if err != nil {
		client.Close()
		return nil, err
	}
	coll := stats.NewCollector(addr)
	rs.nodes[serverID] = &remoteCollector{ssh: client, coll: coll}
	return coll, nil
}

// drop 移除并关闭某服务器的采集器。
func (rs *remoteStats) drop(serverID int64) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if n, ok := rs.nodes[serverID]; ok {
		n.ssh.Close()
		delete(rs.nodes, serverID)
	}
}

// dropAll 关闭全部远端采集器。
func (rs *remoteStats) dropAll() {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	for id, n := range rs.nodes {
		n.ssh.Close()
		delete(rs.nodes, id)
	}
}
