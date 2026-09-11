package v1

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/HYuCN666/volans/internal/model"
)

type inboundRequest struct {
	Remark         string                `json:"remark"`
	Protocol       string                `json:"protocol"`
	Port           int                   `json:"port"`
	Listen         string                `json:"listen"`
	Transport      string                `json:"transport"`
	StreamSettings *model.StreamSettings `json:"stream_settings"`
	TLSEnabled     bool                  `json:"tls_enabled"`
	TLSCert        string                `json:"tls_cert"`
	TLSKey         string                `json:"tls_key"`
	TLSServerName  string                `json:"tls_server_name"`
	EnableSniffing bool                  `json:"enable_sniffing"`
	Enabled        bool                  `json:"enabled"`
}

func (r *inboundRequest) toModel(in *model.Inbound) error {
	in.Remark = r.Remark
	in.Protocol = r.Protocol
	in.Port = r.Port
	in.Listen = r.Listen
	in.Transport = r.Transport
	in.TLSEnabled = r.TLSEnabled
	in.TLSCert = r.TLSCert
	in.TLSKey = r.TLSKey
	in.TLSServerName = r.TLSServerName
	in.EnableSniffing = r.EnableSniffing
	in.Enabled = r.Enabled

	if r.StreamSettings == nil {
		in.StreamSettings = ""
		return nil
	}
	b, err := json.Marshal(r.StreamSettings)
	if err != nil {
		return err
	}
	in.StreamSettings = string(b)
	return nil
}

func (r *inboundRequest) validate() string {
	if r.Protocol != "vless" {
		return "暂仅支持 vless 协议"
	}
	if r.Port < 1 || r.Port > 65535 {
		return "端口超出范围"
	}
	if r.Transport != "tcp" && r.Transport != "ws" {
		return "暂仅支持 tcp/ws 传输"
	}
	if r.TLSEnabled && (r.TLSCert == "" || r.TLSKey == "") {
		return "启用 TLS 需填写证书与私钥路径"
	}
	return ""
}

// ListInbounds 返回入站列表。
func (h *Handler) ListInbounds(c *gin.Context) {
	inbounds, err := h.store.ListInbounds()
	if err != nil {
		fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	users, _ := h.store.ListUsers(0)
	counts := map[int64]int{}
	for _, u := range users {
		counts[u.InboundID]++
	}

	out := make([]gin.H, 0, len(inbounds))
	for _, in := range inbounds {
		out = append(out, gin.H{
			"id":          in.ID,
			"tag":         in.Tag,
			"remark":      in.Remark,
			"protocol":    in.Protocol,
			"port":        in.Port,
			"listen":      in.Listen,
			"transport":   in.Transport,
			"tls_enabled": in.TLSEnabled,
			"enabled":     in.Enabled,
			"user_count":  counts[in.ID],
		})
	}
	ok(c, out)
}

// GetInbound 返回入站详情。
func (h *Handler) GetInbound(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid id")
		return
	}
	in, err := h.store.GetInbound(id)
	if err != nil {
		fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	if in == nil {
		fail(c, http.StatusNotFound, 40400, "入站不存在")
		return
	}
	ok(c, in)
}

// CreateInbound 创建入站。
func (h *Handler) CreateInbound(c *gin.Context) {
	var req inboundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid request")
		return
	}
	if msg := req.validate(); msg != "" {
		fail(c, http.StatusBadRequest, 40001, msg)
		return
	}

	tag, err := h.store.NextInboundTag()
	if err != nil {
		fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	in := &model.Inbound{Tag: tag}
	if err := req.toModel(in); err != nil {
		fail(c, http.StatusBadRequest, 40001, err.Error())
		return
	}

	id, err := h.store.CreateInbound(in)
	if err != nil {
		fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	if err := h.rebuild(); err != nil {
		fail(c, http.StatusInternalServerError, 50010, err.Error())
		return
	}
	ok(c, gin.H{"id": id})
}

// UpdateInbound 更新入站。
func (h *Handler) UpdateInbound(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid id")
		return
	}
	var req inboundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid request")
		return
	}
	if msg := req.validate(); msg != "" {
		fail(c, http.StatusBadRequest, 40001, msg)
		return
	}

	in, err := h.store.GetInbound(id)
	if err != nil || in == nil {
		fail(c, http.StatusNotFound, 40400, "入站不存在")
		return
	}
	if err := req.toModel(in); err != nil {
		fail(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	if err := h.store.UpdateInbound(in); err != nil {
		fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	if err := h.rebuild(); err != nil {
		fail(c, http.StatusInternalServerError, 50010, err.Error())
		return
	}
	ok(c, nil)
}

// DeleteInbound 删除入站。
func (h *Handler) DeleteInbound(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid id")
		return
	}
	if err := h.store.DeleteInbound(id); err != nil {
		fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	if err := h.rebuild(); err != nil {
		fail(c, http.StatusInternalServerError, 50010, err.Error())
		return
	}
	ok(c, nil)
}

// SetInboundStatus 启用/停用入站。
func (h *Handler) SetInboundStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid id")
		return
	}
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid request")
		return
	}
	in, err := h.store.GetInbound(id)
	if err != nil || in == nil {
		fail(c, http.StatusNotFound, 40400, "入站不存在")
		return
	}
	in.Enabled = req.Enabled
	if err := h.store.UpdateInbound(in); err != nil {
		fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	if err := h.rebuild(); err != nil {
		fail(c, http.StatusInternalServerError, 50010, err.Error())
		return
	}
	ok(c, nil)
}
