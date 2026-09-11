package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// IssueCert 通过 ACME 为域名签发证书。
func (h *Handler) IssueCert(c *gin.Context) {
	var req struct {
		Domain string `json:"domain"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Domain == "" {
		fail(c, http.StatusBadRequest, 40001, "域名必填")
		return
	}
	if h.acme == nil {
		fail(c, http.StatusBadRequest, 40001, "ACME 未启用（需 --acme-http 端口）")
		return
	}
	certPath, keyPath, err := h.acme.Issue(req.Domain)
	if err != nil {
		fail(c, http.StatusInternalServerError, 50000, "签发失败: "+err.Error())
		return
	}
	ok(c, gin.H{"cert_path": certPath, "key_path": keyPath})
}
