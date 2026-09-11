package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/HYuCN666/kite/internal/api/middleware"
	v1 "github.com/HYuCN666/kite/internal/api/v1"
	"github.com/HYuCN666/kite/internal/auth"
	"github.com/HYuCN666/kite/internal/config"
	"github.com/HYuCN666/kite/internal/stats"
	"github.com/HYuCN666/kite/internal/store"
	"github.com/HYuCN666/kite/internal/xray"
)

const statsAPIAddr = "127.0.0.1:10085"

// Server 为 HTTP 服务入口。
type Server struct {
	cfg     *config.Config
	db      *store.Store
	auth    *auth.Manager
	xray    *xray.Manager
	stats   *stats.Collector
	handler *v1.Handler
}

// New 构造服务。
func New(cfg *config.Config, db *store.Store) *Server {
	s := &Server{
		cfg:   cfg,
		db:    db,
		auth:  auth.NewManager(cfg.Secret, 24*time.Hour),
		xray:  xray.NewManager(cfg.XrayPath, cfg.DataDir),
		stats: stats.NewCollector(statsAPIAddr),
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
	if _, err := s.db.CreateAdmin("admin", hash); err != nil {
		return err
	}
	log.Println("已创建默认管理员 admin/admin，请尽快修改密码")
	return nil
}

func (s *Server) routes(r *gin.Engine) {
	r.GET("/api/v1/health", s.health)
	r.GET("/sub/:token", s.handler.Sub)

	v1g := r.Group("/api/v1")
	v1g.POST("/auth/login", s.handler.Login)

	authed := v1g.Group("")
	authed.Use(middleware.Auth(s.auth))
	{
		authed.GET("/auth/me", s.handler.Me)
		authed.PUT("/auth/password", s.handler.ChangePassword)

		authed.GET("/node/status", s.handler.Status)
		authed.POST("/node/control", s.handler.Control)
		authed.POST("/node/install", s.handler.Install)

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
}

func (s *Server) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"status": "up"}})
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
	snaps, err := s.stats.Collect()
	if err != nil {
		return
	}
	if len(snaps) == 0 {
		return
	}

	users, err := s.db.ListUsers(0)
	if err != nil {
		return
	}
	emailToID := make(map[string]int64, len(users))
	for _, u := range users {
		emailToID[u.Email] = u.ID
	}

	for _, snap := range snaps {
		if id, ok := emailToID[snap.Email]; ok {
			_ = s.db.AddUserTraffic(id, snap.Uplink, snap.Downlink)
		}
	}

	changed := false
	users, _ = s.db.ListUsers(0)
	now := time.Now()
	for _, u := range users {
		if !u.Enabled {
			continue
		}
		stop := (u.QuotaBytes > 0 && u.UsedUplink+u.UsedDownlink >= u.QuotaBytes) ||
			(u.ExpireAt != nil && u.ExpireAt.Before(now))
		if stop {
			_ = s.db.SetUserEnabled(u.ID, false)
			changed = true
		}
	}
	if changed {
		_ = s.handler.Rebuild()
	}
}
