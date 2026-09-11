package v1

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/HYuCN666/volans/internal/model"
)

type userRequest struct {
	InboundID          int64      `json:"inbound_id"`
	Remark             string     `json:"remark"`
	QuotaBytes         int64      `json:"quota_bytes"`
	SpeedLimitUplink   int64      `json:"speed_limit_uplink"`
	SpeedLimitDownlink int64      `json:"speed_limit_downlink"`
	ExpireAt           *time.Time `json:"expire_at"`
	Enabled            bool       `json:"enabled"`
}

func (r *userRequest) validate() string {
	if r.InboundID <= 0 {
		return "缺少入站"
	}
	return ""
}

// ListUsers 返回用户列表。
func (h *Handler) ListUsers(c *gin.Context) {
	var inboundID int64
	if v := c.Query("inbound_id"); v != "" {
		inboundID, _ = strconv.ParseInt(v, 10, 64)
	}
	users, err := h.store.ListUsers(inboundID)
	if err != nil {
		fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	out := make([]gin.H, 0, len(users))
	for _, u := range users {
		out = append(out, userView(&u, nil))
	}
	ok(c, out)
}

// GetUser 返回用户详情。
func (h *Handler) GetUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid id")
		return
	}
	u, err := h.store.GetUser(id)
	if err != nil {
		fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	if u == nil {
		fail(c, http.StatusNotFound, 40400, "用户不存在")
		return
	}
	sub, _ := h.store.GetSubscriptionByUserID(id)
	ok(c, userView(u, sub))
}

// CreateUser 创建用户。
func (h *Handler) CreateUser(c *gin.Context) {
	var req userRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid request")
		return
	}
	if msg := req.validate(); msg != "" {
		fail(c, http.StatusBadRequest, 40001, msg)
		return
	}

	u := &model.User{
		InboundID:          req.InboundID,
		Remark:             req.Remark,
		QuotaBytes:         req.QuotaBytes,
		SpeedLimitUplink:   req.SpeedLimitUplink,
		SpeedLimitDownlink: req.SpeedLimitDownlink,
		ExpireAt:           req.ExpireAt,
		Enabled:            req.Enabled,
	}
	u.UUID = uuid.New().String()
	u.Email = u.UUID

	id, err := h.store.CreateUser(u)
	if err != nil {
		fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	token, err := randomToken()
	if err != nil {
		fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	if err := h.store.CreateSubscription(id, token); err != nil {
		fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	if err := h.rebuild(); err != nil {
		fail(c, http.StatusInternalServerError, 50010, err.Error())
		return
	}
	ok(c, gin.H{"id": id})
}

// UpdateUser 更新用户。
func (h *Handler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid id")
		return
	}
	var req userRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid request")
		return
	}

	u, err := h.store.GetUser(id)
	if err != nil || u == nil {
		fail(c, http.StatusNotFound, 40400, "用户不存在")
		return
	}
	u.Remark = req.Remark
	u.QuotaBytes = req.QuotaBytes
	u.SpeedLimitUplink = req.SpeedLimitUplink
	u.SpeedLimitDownlink = req.SpeedLimitDownlink
	u.ExpireAt = req.ExpireAt
	u.Enabled = req.Enabled

	if err := h.store.UpdateUser(u); err != nil {
		fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	if err := h.rebuild(); err != nil {
		fail(c, http.StatusInternalServerError, 50010, err.Error())
		return
	}
	ok(c, nil)
}

// DeleteUser 删除用户。
func (h *Handler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid id")
		return
	}
	if err := h.store.DeleteUser(id); err != nil {
		fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	if err := h.rebuild(); err != nil {
		fail(c, http.StatusInternalServerError, 50010, err.Error())
		return
	}
	ok(c, nil)
}

// SetUserStatus 启用/停用用户。
func (h *Handler) SetUserStatus(c *gin.Context) {
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
	if err := h.store.SetUserEnabled(id, req.Enabled); err != nil {
		fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	if err := h.rebuild(); err != nil {
		fail(c, http.StatusInternalServerError, 50010, err.Error())
		return
	}
	ok(c, nil)
}

// ResetUserTraffic 重置用户流量。
func (h *Handler) ResetUserTraffic(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid id")
		return
	}
	if err := h.store.ResetUserTraffic(id); err != nil {
		fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	ok(c, nil)
}

// GetUserSubscription 返回用户订阅信息。
func (h *Handler) GetUserSubscription(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid id")
		return
	}
	sub, err := h.store.GetSubscriptionByUserID(id)
	if err != nil {
		fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	if sub == nil {
		fail(c, http.StatusNotFound, 40400, "订阅不存在")
		return
	}
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	base := scheme + "://" + c.Request.Host
	ok(c, gin.H{
		"token": sub.Token,
		"url":   base + "/sub/" + sub.Token,
	})
}

func userView(u *model.User, sub *model.Subscription) gin.H {
	v := gin.H{
		"id":                   u.ID,
		"inbound_id":           u.InboundID,
		"remark":               u.Remark,
		"uuid":                 u.UUID,
		"quota_bytes":          u.QuotaBytes,
		"used_uplink":          u.UsedUplink,
		"used_downlink":        u.UsedDownlink,
		"speed_limit_uplink":   u.SpeedLimitUplink,
		"speed_limit_downlink": u.SpeedLimitDownlink,
		"expire_at":            u.ExpireAt,
		"enabled":              u.Enabled,
	}
	if sub != nil {
		v["subscription_token"] = sub.Token
	}
	return v
}

func randomToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
