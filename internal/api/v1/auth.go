package v1

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login 处理登录请求。
func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid request")
		return
	}

	admin, err := h.store.GetAdminByUsername(req.Username)
	if err != nil {
		fail(c, http.StatusInternalServerError, 50000, "internal error")
		return
	}
	if admin == nil {
		fail(c, http.StatusUnauthorized, 40100, "用户名或密码错误")
		return
	}

	if admin.LockedUntil != nil && admin.LockedUntil.After(time.Now()) {
		fail(c, http.StatusForbidden, 40300, "账号已锁定，请稍后再试")
		return
	}

	if !h.auth.CheckPassword(admin.PasswordHash, req.Password) {
		attempts, _, _ := h.store.RegisterLoginFailure(admin.ID)
		fail(c, http.StatusUnauthorized, 40100, "用户名或密码错误")
		_ = attempts
		return
	}

	if err := h.store.RegisterLoginSuccess(admin.ID); err != nil {
		fail(c, http.StatusInternalServerError, 50000, "internal error")
		return
	}

	token, err := h.auth.Sign(admin.ID, admin.Username)
	if err != nil {
		fail(c, http.StatusInternalServerError, 50000, "internal error")
		return
	}

	ok(c, gin.H{"token": token})
}

// Me 返回当前登录用户信息。
func (h *Handler) Me(c *gin.Context) {
	claims := mustClaims(c)
	ok(c, gin.H{"uid": claims.UserID, "username": claims.Username})
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// ChangePassword 修改当前管理员密码。
func (h *Handler) ChangePassword(c *gin.Context) {
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid request")
		return
	}
	if len(req.NewPassword) < 8 {
		fail(c, http.StatusBadRequest, 40001, "新密码至少 8 位")
		return
	}

	claims := mustClaims(c)
	admin, err := h.store.GetAdminByUsername(claims.Username)
	if err != nil || admin == nil {
		fail(c, http.StatusInternalServerError, 50000, "internal error")
		return
	}
	if !h.auth.CheckPassword(admin.PasswordHash, req.OldPassword) {
		fail(c, http.StatusUnauthorized, 40100, "旧密码错误")
		return
	}

	hash, err := h.auth.HashPassword(req.NewPassword)
	if err != nil {
		fail(c, http.StatusInternalServerError, 50000, "internal error")
		return
	}
	if err := h.store.UpdateAdminPassword(admin.ID, hash); err != nil {
		fail(c, http.StatusInternalServerError, 50000, "internal error")
		return
	}
	ok(c, nil)
}
