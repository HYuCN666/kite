package v1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/HYuCN666/volans/internal/acme"
	"github.com/HYuCN666/volans/internal/auth"
	"github.com/HYuCN666/volans/internal/model"
	"github.com/HYuCN666/volans/internal/sshx"
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
	acme  *acme.Issuer
	build xray.BuildOptions
}

// New 构造 Handler。
func New(s *store.Store, a *auth.Manager, x *xray.Manager, sc *stats.Collector, ac *acme.Issuer, b xray.BuildOptions) *Handler {
	return &Handler{store: s, auth: a, xray: x, stats: sc, acme: ac, build: b}
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

// rebuild 从数据库重新生成并写入各服务器的 Xray 配置。
func (h *Handler) rebuild() error {
	inbounds, err := h.store.ListInbounds()
	if err != nil {
		return err
	}
	users, err := h.store.ListUsers(0)
	if err != nil {
		return err
	}

	byServer := map[int64][]model.Inbound{}
	for _, in := range inbounds {
		byServer[in.ServerID] = append(byServer[in.ServerID], in)
	}
	usersByInbound := map[int64][]model.User{}
	for _, u := range users {
		usersByInbound[u.InboundID] = append(usersByInbound[u.InboundID], u)
	}

	for serverID, serverInbounds := range byServer {
		var serverUsers []model.User
		for _, in := range serverInbounds {
			serverUsers = append(serverUsers, usersByInbound[in.ID]...)
		}

		cfg, err := xray.BuildConfig(serverInbounds, serverUsers, h.build)
		if err != nil {
			return err
		}

		if serverID == 0 {
			if err := h.xray.WriteConfig(cfg); err != nil {
				return err
			}
			continue
		}

		remote, err := h.remoteFor(serverID)
		if err != nil {
			return err
		}
		if err := remote.WriteConfig(cfg); err != nil {
			remote.Close()
			return err
		}
		remote.Close()
	}
	return nil
}

// remoteFor 根据服务器 ID 构造远端管理器。
func (h *Handler) remoteFor(serverID int64) (*xray.Remote, error) {
	sv, err := h.store.GetServer(serverID)
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
	return xray.NewRemote(client, "/usr/local/bin/xray", "/etc/xray/config.json"), nil
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
