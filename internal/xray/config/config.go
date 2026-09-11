package config

import "encoding/json"

// Config 为 Xray 运行时配置的顶层结构。
type Config struct {
	Log       *LogConfig       `json:"log,omitempty"`
	Inbounds  []Inbound        `json:"inbounds,omitempty"`
	Outbounds []Outbound       `json:"outbounds,omitempty"`
	Stats     *StatsConfig     `json:"stats,omitempty"`
	Policy    *PolicyConfig    `json:"policy,omitempty"`
	API       *APIConfig       `json:"api,omitempty"`
	Routing   *RoutingConfig   `json:"routing,omitempty"`
}

// LogConfig 为日志配置。
type LogConfig struct {
	Access   string `json:"access,omitempty"`
	Error    string `json:"error,omitempty"`
	Loglevel string `json:"loglevel,omitempty"`
}

// Inbound 为入站配置。
type Inbound struct {
	Tag            string          `json:"tag,omitempty"`
	Listen         string          `json:"listen,omitempty"`
	Port           int             `json:"port,omitempty"`
	Protocol       string          `json:"protocol"`
	Settings       json.RawMessage `json:"settings,omitempty"`
	StreamSettings *StreamSettings `json:"streamSettings,omitempty"`
	Sniffing       *Sniffing       `json:"sniffing,omitempty"`
}

// VLESSUser 为 VLESS 入站用户。
type VLESSUser struct {
	ID         string `json:"id"`
	Email      string `json:"email,omitempty"`
	Encryption string `json:"encryption,omitempty"`
	Level      int    `json:"level,omitempty"`
}

// VLESSInboundSettings 为 VLESS 入站设置。
type VLESSInboundSettings struct {
	Clients    []VLESSUser `json:"clients"`
	Decryption string      `json:"decryption,omitempty"`
}

// StreamSettings 为传输层设置。
type StreamSettings struct {
	Network  string          `json:"network,omitempty"`
	Security string          `json:"security,omitempty"`
	TLSSettings *TLSSettings `json:"tlsSettings,omitempty"`
	WSSettings  *WSSettings  `json:"wsSettings,omitempty"`
}

// TLSSettings 为 TLS 设置。
type TLSSettings struct {
	ServerName    string        `json:"serverName,omitempty"`
	Certificates  []Certificate `json:"certificates,omitempty"`
	ALPN          []string      `json:"alpn,omitempty"`
}

// Certificate 为 TLS 证书。
type Certificate struct {
	CertificateFile string `json:"certificateFile,omitempty"`
	KeyFile         string `json:"keyFile,omitempty"`
}

// WSSettings 为 WebSocket 设置。
type WSSettings struct {
	Path    string            `json:"path,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// Sniffing 为流量嗅探设置。
type Sniffing struct {
	Enabled      bool     `json:"enabled,omitempty"`
	DestOverride []string `json:"destOverride,omitempty"`
}

// Outbound 为出站配置。
type Outbound struct {
	Tag      string `json:"tag,omitempty"`
	Protocol string `json:"protocol"`
}

// StatsConfig 开启统计。
type StatsConfig struct{}

// PolicyConfig 为策略配置。
type PolicyConfig struct {
	Levels map[string]*PolicyLevel `json:"levels,omitempty"`
	System *PolicySystem           `json:"system,omitempty"`
}

// PolicyLevel 为某一级别的策略。
type PolicyLevel struct {
	Handshake         int  `json:"handshake,omitempty"`
	ConnIdle          int  `json:"connIdle,omitempty"`
	UplinkOnly        int  `json:"uplinkOnly,omitempty"`
	DownlinkOnly      int  `json:"downlinkOnly,omitempty"`
	StatsUserUplink   bool `json:"statsUserUplink,omitempty"`
	StatsUserDownlink bool `json:"statsUserDownlink,omitempty"`
}

// PolicySystem 为系统级策略。
type PolicySystem struct {
	StatsInboundUplink   bool `json:"statsInboundUplink,omitempty"`
	StatsInboundDownlink bool `json:"statsInboundDownlink,omitempty"`
}

// APIConfig 为 Xray API 配置。
type APIConfig struct {
	Tag      string   `json:"tag"`
	Services []string `json:"services,omitempty"`
}

// RoutingConfig 为路由配置。
type RoutingConfig struct {
	Rules []RoutingRule `json:"rules,omitempty"`
}

// RoutingRule 为单条路由规则。
type RoutingRule struct {
	Type        string   `json:"type,omitempty"`
	InboundTag  []string `json:"inboundTag,omitempty"`
	OutboundTag string   `json:"outboundTag,omitempty"`
}
