package subscription

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/HYuCN666/volans/internal/model"
)

// Node 表示一个可订阅的代理节点。
type Node struct {
	Name     string
	Protocol string
	Host     string
	Port     int
	UUID     string
	Network  string
	TLS      bool
	SNI      string
	WSPath   string
}

// BuildNode 从入站与用户组装节点信息。
func BuildNode(in model.Inbound, u model.User, host string) (*Node, error) {
	n := &Node{
		Name:     u.Remark,
		Protocol: in.Protocol,
		Host:     host,
		Port:     in.Port,
		UUID:     u.UUID,
		Network:  in.Transport,
		TLS:      in.TLSEnabled,
		SNI:      in.TLSServerName,
	}
	if n.Name == "" {
		n.Name = in.Remark
	}
	if n.Name == "" {
		n.Name = fmt.Sprintf("%s-%d", in.Protocol, in.Port)
	}
	if in.Transport == "ws" {
		ss, err := in.ParseStreamSettings()
		if err != nil {
			return nil, err
		}
		n.WSPath = ss.Path
		if n.WSPath == "" {
			n.WSPath = "/"
		}
	}
	return n, nil
}

// ShareLink 生成 vless:// 分享链接。
func (n *Node) ShareLink() string {
	q := url.Values{}
	q.Set("encryption", "none")
	if n.TLS {
		q.Set("security", "tls")
		if n.SNI != "" {
			q.Set("sni", n.SNI)
		}
	}
	if n.Network != "" && n.Network != "tcp" {
		q.Set("type", n.Network)
	}
	if n.Network == "ws" {
		q.Set("path", n.WSPath)
		q.Set("host", n.Host)
	}
	fragment := url.QueryEscape(n.Name)
	return fmt.Sprintf("vless://%s@%s:%d?%s#%s", n.UUID, n.Host, n.Port, q.Encode(), fragment)
}

// RenderV2RayN 生成 base64 订阅内容。
func RenderV2RayN(nodes []*Node) string {
	lines := make([]string, 0, len(nodes))
	for _, n := range nodes {
		lines = append(lines, n.ShareLink())
	}
	return base64.StdEncoding.EncodeToString([]byte(strings.Join(lines, "\n")))
}

// RenderClash 生成 Clash Meta YAML 订阅。
func RenderClash(nodes []*Node) (string, error) {
	type wsOpts struct {
		Path string `yaml:"path"`
		Host string `yaml:"host,omitempty"`
	}
	type proxy struct {
		Name       string  `yaml:"name"`
		Type       string  `yaml:"type"`
		Server     string  `yaml:"server"`
		Port       int     `yaml:"port"`
		UUID       string  `yaml:"uuid"`
		Network    string  `yaml:"network,omitempty"`
		TLS        bool    `yaml:"tls,omitempty"`
		Servername string  `yaml:"servername,omitempty"`
		WSOpts     *wsOpts `yaml:"ws-opts,omitempty"`
	}

	proxies := make([]proxy, 0, len(nodes))
	names := make([]string, 0, len(nodes))
	for _, n := range nodes {
		p := proxy{
			Name:    n.Name,
			Type:    n.Protocol,
			Server:  n.Host,
			Port:    n.Port,
			UUID:    n.UUID,
			Network: n.Network,
			TLS:     n.TLS,
		}
		if n.TLS {
			p.Servername = n.SNI
		}
		if n.Network == "ws" {
			p.WSOpts = &wsOpts{Path: n.WSPath, Host: n.Host}
		}
		proxies = append(proxies, p)
		names = append(names, n.Name)
	}

	doc := map[string]any{
		"proxies": proxies,
		"proxy-groups": []map[string]any{
			{"name": "PROXY", "type": "select", "proxies": names},
		},
		"rules": []string{"MATCH,PROXY"},
	}

	out, err := yaml.Marshal(doc)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
