package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// TwoFAStatus 返回当前账号 2FA 启用状态。
func (h *Handler) TwoFAStatus(c *gin.Context) {
	claims := mustClaims(c)
	admin, _ := h.store.GetAdminByID(claims.UserID)
	ok(c, gin.H{"enabled": admin != nil && admin.TOTPSecret != ""})
}

// TwoFAEnable 生成 TOTP 密钥与二维码地址（未激活）。
func (h *Handler) TwoFAEnable(c *gin.Context) {
	claims := mustClaims(c)
	secret := h.auth.GenerateTOTPSecret()
	ok(c, gin.H{
		"secret": secret,
		"url":    h.auth.TOTPURL(secret, "Volans", claims.Username),
	})
}

// TwoFAConfirm 校验动态码并激活 2FA。
func (h *Handler) TwoFAConfirm(c *gin.Context) {
	var req struct {
		Secret string `json:"secret"`
		Code   string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid request")
		return
	}
	if !h.auth.ValidateTOTP(req.Secret, req.Code, 1) {
		fail(c, http.StatusBadRequest, 40001, "动态验证码错误")
		return
	}
	claims := mustClaims(c)
	if err := h.store.UpdateAdminTOTPSecret(claims.UserID, req.Secret); err != nil {
		fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	ok(c, nil)
}

// TwoFADisable 校验动态码并关闭 2FA。
func (h *Handler) TwoFADisable(c *gin.Context) {
	var req struct {
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid request")
		return
	}
	claims := mustClaims(c)
	admin, err := h.store.GetAdminByID(claims.UserID)
	if err != nil || admin == nil {
		fail(c, http.StatusInternalServerError, 50000, "internal error")
		return
	}
	if !h.auth.ValidateTOTP(admin.TOTPSecret, req.Code, 1) {
		fail(c, http.StatusBadRequest, 40001, "动态验证码错误")
		return
	}
	if err := h.store.UpdateAdminTOTPSecret(admin.ID, ""); err != nil {
		fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	ok(c, nil)
}

// ListAdmins 返回管理员列表（仅 admin）。
func (h *Handler) ListAdmins(c *gin.Context) {
	if !isAdmin(c) {
		fail(c, http.StatusForbidden, 40300, "需要管理员权限")
		return
	}
	admins, err := h.store.ListAdmins()
	if err != nil {
		fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	out := make([]gin.H, 0, len(admins))
	for _, a := range admins {
		out = append(out, gin.H{
			"id": a.ID, "username": a.Username, "role": a.Role,
			"totp_enabled": a.TOTPSecret != "", "created_at": a.CreatedAt,
		})
	}
	ok(c, out)
}

// CreateAdmin 创建管理员（仅 admin）。
func (h *Handler) CreateAdmin(c *gin.Context) {
	if !isAdmin(c) {
		fail(c, http.StatusForbidden, 40300, "需要管理员权限")
		return
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid request")
		return
	}
	if req.Username == "" || len(req.Password) < 8 {
		fail(c, http.StatusBadRequest, 40001, "用户名必填，密码至少 8 位")
		return
	}
	if req.Role != "operator" {
		req.Role = "admin"
	}
	hash, err := h.auth.HashPassword(req.Password)
	if err != nil {
		fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	id, err := h.store.CreateAdmin(req.Username, hash, req.Role)
	if err != nil {
		fail(c, http.StatusConflict, 40900, "用户名已存在")
		return
	}
	ok(c, gin.H{"id": id})
}

// UpdateAdmin 更新管理员角色或密码（仅 admin）。
func (h *Handler) UpdateAdmin(c *gin.Context) {
	if !isAdmin(c) {
		fail(c, http.StatusForbidden, 40300, "需要管理员权限")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid id")
		return
	}
	var req struct {
		Role     string `json:"role"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid request")
		return
	}
	if req.Role == "admin" || req.Role == "operator" {
		if err := h.store.UpdateAdminRole(id, req.Role); err != nil {
			fail(c, http.StatusInternalServerError, 50000, err.Error())
			return
		}
	}
	if req.Password != "" {
		if len(req.Password) < 8 {
			fail(c, http.StatusBadRequest, 40001, "密码至少 8 位")
			return
		}
		hash, err := h.auth.HashPassword(req.Password)
		if err != nil {
			fail(c, http.StatusInternalServerError, 50000, err.Error())
			return
		}
		if err := h.store.UpdateAdminPassword(id, hash); err != nil {
			fail(c, http.StatusInternalServerError, 50000, err.Error())
			return
		}
	}
	ok(c, nil)
}

// DeleteAdmin 删除管理员（仅 admin，不能删除自己）。
func (h *Handler) DeleteAdmin(c *gin.Context) {
	if !isAdmin(c) {
		fail(c, http.StatusForbidden, 40300, "需要管理员权限")
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid id")
		return
	}
	claims := mustClaims(c)
	if id == claims.UserID {
		fail(c, http.StatusBadRequest, 40001, "不能删除当前登录账号")
		return
	}
	if err := h.store.DeleteAdmin(id); err != nil {
		fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	ok(c, nil)
}

func isAdmin(c *gin.Context) bool {
	claims := mustClaims(c)
	return claims != nil && claims.Role == "admin"
}
