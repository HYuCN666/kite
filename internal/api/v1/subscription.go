package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/HYuCN666/volans/internal/model"
	"github.com/HYuCN666/volans/internal/subscription"
)

// Sub 处理公开的订阅下载（/sub/:token）。
func (h *Handler) Sub(c *gin.Context) {
	token := c.Param("token")
	sub, err := h.store.GetSubscriptionByToken(token)
	if err != nil {
		c.String(http.StatusInternalServerError, "internal error")
		return
	}
	if sub == nil {
		c.String(http.StatusNotFound, "subscription not found")
		return
	}

	u, err := h.store.GetUser(sub.UserID)
	if err != nil || u == nil {
		c.String(http.StatusNotFound, "user not found")
		return
	}
	if !u.Enabled {
		c.String(http.StatusForbidden, "user disabled")
		return
	}

	in, err := h.store.GetInbound(u.InboundID)
	if err != nil || in == nil {
		c.String(http.StatusNotFound, "inbound not found")
		return
	}

	host := h.resolveHost(in)
	if host == "" {
		c.String(http.StatusInternalServerError, "server host not configured")
		return
	}

	node, err := subscription.BuildNode(*in, *u, host)
	if err != nil {
		c.String(http.StatusInternalServerError, "build node failed")
		return
	}
	nodes := []*subscription.Node{node}

	format := c.DefaultQuery("format", "v2rayn")
	switch format {
	case "clash":
		body, err := subscription.RenderClash(nodes)
		if err != nil {
			c.String(http.StatusInternalServerError, "render failed")
			return
		}
		c.Data(http.StatusOK, "text/yaml; charset=utf-8", []byte(body))
	default:
		c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(subscription.RenderV2RayN(nodes)))
	}
}

func (h *Handler) resolveHost(in *model.Inbound) string {
	if in.TLSServerName != "" {
		return in.TLSServerName
	}
	if host, _ := h.store.GetSetting("public_host"); host != "" {
		return host
	}
	return ""
}

// Overview 返回流量与用户概览。
func (h *Handler) Overview(c *gin.Context) {
	up, down, active, err := h.store.TrafficOverview()
	if err != nil {
		fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	ok(c, gin.H{
		"total_uplink":   up,
		"total_downlink": down,
		"active_users":   active,
	})
}
