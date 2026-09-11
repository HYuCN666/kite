package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/HYuCN666/kite/internal/api/middleware"
	"github.com/HYuCN666/kite/internal/auth"
	"github.com/HYuCN666/kite/internal/config"
	"github.com/HYuCN666/kite/internal/store"
	"github.com/HYuCN666/kite/internal/xray"
)

// Server 为 HTTP 服务入口。
type Server struct {
	cfg  *config.Config
	db   *store.Store
	auth *auth.Manager
	xray *xray.Manager
}

// New 构造服务。
func New(cfg *config.Config, db *store.Store) *Server {
	return &Server{
		cfg:  cfg,
		db:   db,
		auth: auth.NewManager(cfg.Secret, 24*time.Hour),
		xray: xray.NewManager(cfg.XrayPath, cfg.DataDir),
	}
}

// Run 启动 HTTP 服务。
func (s *Server) Run() error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	s.routes(r)

	addr := fmt.Sprintf("%s:%d", s.cfg.Bind, s.cfg.Port)
	return r.Run(addr)
}

func (s *Server) routes(r *gin.Engine) {
	r.GET("/api/v1/health", s.health)

	v1 := r.Group("/api/v1")
	v1.POST("/auth/login", s.login)

	authed := v1.Group("")
	authed.Use(middleware.Auth(s.auth))
	{
		authed.GET("/auth/me", s.me)
	}
}

func (s *Server) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"status": "up"}})
}

func (s *Server) login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40001, "message": "invalid request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"token": ""}})
}

func (s *Server) me(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"status": "up"}})
}
