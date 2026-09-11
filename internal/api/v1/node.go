package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Status 返回内核状态。
func (h *Handler) Status(c *gin.Context) {
	installed := h.xray.Installed()
	version := ""
	if installed {
		if v, err := h.xray.Version(); err == nil {
			version = v
		}
	}
	ok(c, gin.H{
		"installed": installed,
		"version":   version,
		"running":   h.xray.Running(),
	})
}

type controlRequest struct {
	Action string `json:"action"`
}

// Control 控制内核进程（start/stop/restart）。
func (h *Handler) Control(c *gin.Context) {
	var req controlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid request")
		return
	}

	var err error
	switch req.Action {
	case "start":
		err = h.xray.Start()
	case "stop":
		err = h.xray.Stop()
	case "restart":
		err = h.xray.Restart()
	default:
		fail(c, http.StatusBadRequest, 40001, "unknown action")
		return
	}
	if err != nil {
		fail(c, http.StatusInternalServerError, 50010, err.Error())
		return
	}
	ok(c, nil)
}

// Install 下载并安装内核。
func (h *Handler) Install(c *gin.Context) {
	if err := h.xray.Install(); err != nil {
		fail(c, http.StatusInternalServerError, 50010, err.Error())
		return
	}
	if err := h.rebuild(); err != nil {
		fail(c, http.StatusInternalServerError, 50010, err.Error())
		return
	}
	ok(c, nil)
}
