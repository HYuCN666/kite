package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/HYuCN666/volans/internal/model"
	"github.com/HYuCN666/volans/internal/sshx"
)

type serverRequest struct {
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	AuthType   string `json:"auth_type"`
	Password   string `json:"password"`
	PrivateKey string `json:"private_key"`
	Enabled    bool   `json:"enabled"`
}

func (r *serverRequest) validate() string {
	if r.Name == "" || r.Host == "" {
		return "名称与主机地址必填"
	}
	if r.Port <= 0 || r.Port > 65535 {
		return "SSH 端口无效"
	}
	if r.Username == "" {
		return "SSH 用户必填"
	}
	if r.AuthType != "key" && r.Password == "" {
		return "需填写密码或私钥"
	}
	return ""
}

func (r *serverRequest) toModel(sv *model.Server) {
	sv.Name = r.Name
	sv.Host = r.Host
	sv.Port = r.Port
	sv.Username = r.Username
	sv.AuthType = r.AuthType
	sv.Password = r.Password
	sv.PrivateKey = r.PrivateKey
	sv.Enabled = r.Enabled
}

// ListServers 返回服务器列表。
func (h *Handler) ListServers(c *gin.Context) {
	servers, err := h.store.ListServers()
	if err != nil {
		fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	out := make([]gin.H, 0, len(servers)+1)
	out = append(out, gin.H{"id": 0, "name": "本机 (Local)", "host": "", "enabled": true})
	for _, sv := range servers {
		out = append(out, gin.H{
			"id": sv.ID, "name": sv.Name, "host": sv.Host, "port": sv.Port,
			"username": sv.Username, "auth_type": sv.AuthType, "enabled": sv.Enabled,
		})
	}
	ok(c, out)
}

// CreateServer 创建服务器。
func (h *Handler) CreateServer(c *gin.Context) {
	var req serverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid request")
		return
	}
	if msg := req.validate(); msg != "" {
		fail(c, http.StatusBadRequest, 40001, msg)
		return
	}
	if req.Port == 0 {
		req.Port = 22
	}
	sv := &model.Server{}
	req.toModel(sv)
	id, err := h.store.CreateServer(sv)
	if err != nil {
		fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	ok(c, gin.H{"id": id})
}

// UpdateServer 更新服务器。
func (h *Handler) UpdateServer(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid id")
		return
	}
	var req serverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid request")
		return
	}
	sv, err := h.store.GetServer(id)
	if err != nil || sv == nil {
		fail(c, http.StatusNotFound, 40400, "服务器不存在")
		return
	}
	if req.Name != "" {
		req.toModel(sv)
	} else {
		sv.Enabled = req.Enabled
	}
	if err := h.store.UpdateServer(sv); err != nil {
		fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	ok(c, nil)
}

// DeleteServer 删除服务器。
func (h *Handler) DeleteServer(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid id")
		return
	}
	if err := h.store.DeleteServer(id); err != nil {
		fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}
	ok(c, nil)
}

// TestServer 测试 SSH 连接。
func (h *Handler) TestServer(c *gin.Context) {
	var req serverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid request")
		return
	}
	client := sshx.New(sshx.Config{
		Host: req.Host, Port: req.Port, Username: req.Username,
		AuthType: req.AuthType, Password: req.Password, PrivateKey: req.PrivateKey,
	})
	if _, err := client.Run("echo ok"); err != nil {
		fail(c, http.StatusBadGateway, 50010, "连接失败: "+err.Error())
		return
	}
	client.Close()
	ok(c, gin.H{"connected": true})
}

// ServerStatus 返回远端节点状态。
func (h *Handler) ServerStatus(c *gin.Context) {
	remote, err := h.remoteFor(parseID(c))
	if err != nil {
		fail(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	defer remote.Close()

	version := ""
	if remote.Installed() {
		if v, err := remote.Version(); err == nil {
			version = v
		}
	}
	ok(c, gin.H{"installed": remote.Installed(), "version": version, "running": remote.Running()})
}

// ServerInstall 在远端安装 Xray。
func (h *Handler) ServerInstall(c *gin.Context) {
	remote, err := h.remoteFor(parseID(c))
	if err != nil {
		fail(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	defer remote.Close()
	if err := remote.Install(); err != nil {
		fail(c, http.StatusInternalServerError, 50010, err.Error())
		return
	}
	if err := h.rebuild(); err != nil {
		fail(c, http.StatusInternalServerError, 50010, err.Error())
		return
	}
	ok(c, nil)
}

// ServerControl 控制远端进程。
func (h *Handler) ServerControl(c *gin.Context) {
	var req struct {
		Action string `json:"action"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 40001, "invalid request")
		return
	}
	remote, err := h.remoteFor(parseID(c))
	if err != nil {
		fail(c, http.StatusBadRequest, 40001, err.Error())
		return
	}
	defer remote.Close()

	switch req.Action {
	case "start":
		err = remote.Start()
	case "stop":
		err = remote.Stop()
	case "restart":
		err = remote.Restart()
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

func parseID(c *gin.Context) int64 {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	return id
}
