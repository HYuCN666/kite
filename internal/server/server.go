package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/HYuCN666/volans/internal/api/middleware"
	v1 "github.com/HYuCN666/volans/internal/api/v1"
	"github.com/HYuCN666/volans/internal/auth"
	"github.com/HYuCN666/volans/internal/config"
	"github.com/HYuCN666/volans/internal/model"
	"github.com/HYuCN666/volans/internal/stats"
	"github.com/HYuCN666/volans/internal/store"
	"github.com/HYuCN666/volans/internal/ws"
	"github.com/HYuCN666/volans/internal/xray"
)

const statsAPIAddr = "127.0.0.1:10085"

// Server 为 HTTP 服务入口。
type Server struct {
	cfg     *config.Config
	db      *store.Store
	auth    *auth.Manager
	xray    *xray.Manager
	stats   *stats.Collector
	hub     *ws.Hub
	handler *v1.Handler
	remote  *remoteStats
}

// New 构造服务。
func New(cfg *config.Config, db *store.Store) *Server {
	s := &Server{
		cfg:    cfg,
		db:     db,
		auth:   auth.NewManager(cfg.Secret, 24*time.Hour),
		xray:   xray.NewManager(cfg.XrayPath, cfg.DataDir),
		stats:  stats.NewCollector(statsAPIAddr),
		hub:    ws.NewHub(),
		remote: newRemoteStats(),
	}
	build := xray.BuildOptions{
		APIListen: "127.0.0.1",
		APIPort:   10085,
		AccessLog: filepath.Join(cfg.DataDir, "xray", "access.log"),
		ErrorLog:  filepath.Join(cfg.DataDir, "xray", "error.log"),
	}
	s.handler = v1.New(db, s.auth, s.xray, s.stats, build)
	return s
}

// Run 启动 HTTP 服务与后台维护任务。
func (s *Server) Run() error {
	if err := s.bootstrap(); err != nil {
		return err
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	s.routes(r)

	go s.maintenance(context.Background())

	addr := fmt.Sprintf("%s:%d", s.cfg.Bind, s.cfg.Port)
	if s.cfg.TLSCert != "" && s.cfg.TLSKey != "" {
		return r.RunTLS(addr, s.cfg.TLSCert, s.cfg.TLSKey)
	}
	return r.Run(addr)
}

func (s *Server) bootstrap() error {
	n, err := s.db.CountAdmins()
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	hash, err := s.auth.HashPassword("admin")
	if err != nil {
		return err
	}
	if _, err := s.db.CreateAdmin("admin", hash, "admin"); err != nil {
		return err
	}
	log.Println("已创建默认管理员 admin/admin，请尽快修改密码")
	return nil
}

func (s *Server) routes(r *gin.Engine) {
	r.GET("/api/v1/health", s.health)
	r.GET("/sub/:token", s.handler.Sub)
	r.GET("/ws", s.wsHandler)

	v1g := r.Group("/api/v1")
	v1g.POST("/auth/login", middleware.NewRateLimiter(5, time.Minute).Middleware(), s.handler.Login)

	authed := v1g.Group("")
	authed.Use(middleware.Auth(s.auth))
	{
		authed.GET("/auth/me", s.handler.Me)
		authed.PUT("/auth/password", s.handler.ChangePassword)

		authed.GET("/auth/2fa/status", s.handler.TwoFAStatus)
		authed.POST("/auth/2fa/enable", s.handler.TwoFAEnable)
		authed.POST("/auth/2fa/confirm", s.handler.TwoFAConfirm)
		authed.POST("/auth/2fa/disable", s.handler.TwoFADisable)

		authed.GET("/admins", s.handler.ListAdmins)
		authed.POST("/admins", s.handler.CreateAdmin)
		authed.PUT("/admins/:id", s.handler.UpdateAdmin)
		authed.DELETE("/admins/:id", s.handler.DeleteAdmin)

		authed.GET("/node/status", s.handler.Status)
		authed.POST("/node/control", s.handler.Control)
		authed.POST("/node/install", s.handler.Install)

		authed.GET("/servers", s.handler.ListServers)
		authed.POST("/servers", s.handler.CreateServer)
		authed.PUT("/servers/:id", s.handler.UpdateServer)
		authed.DELETE("/servers/:id", s.handler.DeleteServer)
		authed.POST("/servers/test", s.handler.TestServer)
		authed.GET("/servers/:id/status", s.handler.ServerStatus)
		authed.POST("/servers/:id/install", s.handler.ServerInstall)
		authed.POST("/servers/:id/control", s.handler.ServerControl)

		authed.GET("/inbounds", s.handler.ListInbounds)
		authed.POST("/inbounds", s.handler.CreateInbound)
		authed.GET("/inbounds/:id", s.handler.GetInbound)
		authed.PUT("/inbounds/:id", s.handler.UpdateInbound)
		authed.DELETE("/inbounds/:id", s.handler.DeleteInbound)
		authed.PATCH("/inbounds/:id/status", s.handler.SetInboundStatus)

		authed.GET("/users", s.handler.ListUsers)
		authed.POST("/users", s.handler.CreateUser)
		authed.GET("/users/:id", s.handler.GetUser)
		authed.PUT("/users/:id", s.handler.UpdateUser)
		authed.DELETE("/users/:id", s.handler.DeleteUser)
		authed.PATCH("/users/:id/status", s.handler.SetUserStatus)
		authed.POST("/users/:id/reset-traffic", s.handler.ResetUserTraffic)
		authed.GET("/users/:id/subscription", s.handler.GetUserSubscription)

		authed.GET("/stats/overview", s.handler.Overview)

		authed.GET("/settings", s.handler.GetSettings)
		authed.PUT("/settings", s.handler.UpdateSettings)
	}

	s.serveStatic(r)
}

func (s *Server) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"status": "up"}})
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(*http.Request) bool { return true },
}

func (s *Server) wsHandler(c *gin.Context) {
	token := c.Query("token")
	if _, err := s.auth.Parse(token); err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	s.hub.Register(conn)
	defer s.hub.Unregister(conn)

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

// serveStatic 服务前端产物（若存在）。
func (s *Server) serveStatic(r *gin.Engine) {
	dist := s.cfg.WebDir
	if _, err := os.Stat(dist); err != nil {
		return
	}
	r.NoRoute(func(c *gin.Context) {
		p := filepath.Join(dist, filepath.Clean("/"+c.Request.URL.Path))
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			c.File(p)
			return
		}
		c.File(filepath.Join(dist, "index.html"))
	})
}

// maintenance 周期采集流量并执行配额/到期停用。
func (s *Server) maintenance(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.applyTraffic()
		}
	}
}

func (s *Server) applyTraffic() {
	snaps := s.collectAllStats()

	if len(snaps) > 0 {
		users, err := s.db.ListUsers(0)
		if err != nil {
			return
		}
		emailToID := make(map[string]int64, len(users))
		for _, u := range users {
			emailToID[u.Email] = u.ID
		}

		var totalUp, totalDown int64
		for _, snap := range snaps {
			totalUp += snap.Uplink
			totalDown += snap.Downlink
			if id, ok := emailToID[snap.Email]; ok {
				_ = s.db.AddUserTraffic(id, snap.Uplink, snap.Downlink)
			}
		}

		s.hub.Broadcast(map[string]any{
			"type": "traffic",
			"data": map[string]any{
				"uplink":   totalUp,
				"downlink": totalDown,
				"time":     time.Now().Unix(),
			},
		})
	}

	changed := false
	users, _ := s.db.ListUsers(0)
	now := time.Now()
	for _, u := range users {
		if !u.Enabled {
			continue
		}
		stop := (u.QuotaBytes > 0 && u.UsedUplink+u.UsedDownlink >= u.QuotaBytes) ||
			(u.ExpireAt != nil && u.ExpireAt.Before(now))
		if u.MaxDevices > 0 {
			if coll, err := s.collectorForUser(&u); err == nil {
				if devices, err := coll.OnlineDevices(u.Email); err == nil && devices > int(u.MaxDevices) {
					stop = true
				}
			}
		}
		if stop {
			_ = s.db.SetUserEnabled(u.ID, false)
			changed = true
		}
	}
	if changed {
		_ = s.handler.Rebuild()
	}
}

// collectAllStats 汇总本机与所有远端服务器的流量快照。
func (s *Server) collectAllStats() []stats.Snapshot {
	var all []stats.Snapshot
	if snaps, err := s.stats.Collect(); err == nil {
		all = append(all, snaps...)
	}

	servers, err := s.db.ListServers()
	if err != nil {
		return all
	}
	for _, sv := range servers {
		if !sv.Enabled {
			continue
		}
		coll, err := s.remote.get(s.db, sv.ID)
		if err != nil {
			continue
		}
		snaps, err := coll.Collect()
		if err != nil {
			s.remote.drop(sv.ID)
			continue
		}
		all = append(all, snaps...)
	}
	return all
}

// collectorForUser 返回用户所属服务器对应的统计采集器。
func (s *Server) collectorForUser(u *model.User) (*stats.Collector, error) {
	in, err := s.db.GetInbound(u.InboundID)
	if err != nil || in == nil {
		return nil, fmt.Errorf("inbound not found")
	}
	if in.ServerID == 0 {
		return s.stats, nil
	}
	return s.remote.get(s.db, in.ServerID)
}
