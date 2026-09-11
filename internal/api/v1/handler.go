package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/HYuCN666/volans/internal/auth"
	"github.com/HYuCN666/volans/internal/stats"
	"github.com/HYuCN666/volans/internal/store"
	"github.com/HYuCN666/volans/internal/xray"
)

// Handler 聚合所有 API handler 依赖。
type Handler struct {
	store *store.Store
	auth  *auth.Manager
	xray  *xray.Manager
	stats *stats.Collector
	build xray.BuildOptions
}

// New 构造 Handler。
func New(s *store.Store, a *auth.Manager, x *xray.Manager, sc *stats.Collector, b xray.BuildOptions) *Handler {
	return &Handler{store: s, auth: a, xray: x, stats: sc, build: b}
}

func respond(c *gin.Context, status, code int, message string, data any) {
	c.JSON(status, gin.H{"code": code, "message": message, "data": data})
}

func ok(c *gin.Context, data any) {
	respond(c, http.StatusOK, 0, "ok", data)
}

func fail(c *gin.Context, status, code int, message string) {
	respond(c, status, code, message, nil)
}

// rebuild 从数据库重新生成并写入 Xray 配置。
func (h *Handler) rebuild() error {
	inbounds, err := h.store.ListInbounds()
	if err != nil {
		return err
	}
	users, err := h.store.ListUsers(0)
	if err != nil {
		return err
	}
	cfg, err := xray.BuildConfig(inbounds, users, h.build)
	if err != nil {
		return err
	}
	return h.xray.WriteConfig(cfg)
}

// Rebuild 导出重建操作，供后台任务调用。
func (h *Handler) Rebuild() error {
	return h.rebuild()
}

// mustClaims 从上下文读取已校验的 JWT claims。
func mustClaims(c *gin.Context) *auth.Claims {
	v, _ := c.Get("claims")
	claims, _ := v.(*auth.Claims)
	return claims
}
