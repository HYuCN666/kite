package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetSettings 返回全部面板设置。
func (h *Handler) GetSettings(c *gin.Context) {
	m, err := h.store.ListSettings()
	if err != nil {
		fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	ok(c, m)
}

// UpdateSettings 更新面板设置（键值对）。
func (h *Handler) UpdateSettings(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid request")
		return
	}
	for k, v := range req {
		if err := h.store.SetSetting(k, v); err != nil {
			fail(c, http.StatusInternalServerError, 50000, err.Error())
			return
		}
	}
	ok(c, nil)
}
