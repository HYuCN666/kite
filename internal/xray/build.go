package xray

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/HYuCN666/volans/internal/model"
	"github.com/HYuCN666/volans/internal/xray/config"
)

// BuildOptions 控制配置生成的路径与 API 端口。
type BuildOptions struct {
	APIListen string
	APIPort   int
	AccessLog string
	ErrorLog  string
}

// BuildConfig 根据入站与用户数据组装 Xray 配置。
func BuildConfig(inbounds []model.Inbound, users []model.User, opts BuildOptions) (*config.Config, error) {
	levels, levelOf := buildSpeedLevels(users)

	cfg := &config.Config{
		Log: &config.LogConfig{
			Access:   opts.AccessLog,
			Error:    opts.ErrorLog,
			Loglevel: "warning",
		},
		API: &config.APIConfig{
			Tag:      "api",
			Services: []string{"StatsService"},
		},
		Stats: &config.StatsConfig{},
		Policy: &config.PolicyConfig{
			Levels: levels,
			System: &config.PolicySystem{
				StatsInboundUplink:    true,
				StatsInboundDownlink:  true,
				StatsOutboundUplink:   true,
				StatsOutboundDownlink: true,
			},
		},
		Outbounds: []config.Outbound{
			{Tag: "direct", Protocol: "freedom"},
		},
		Routing: &config.RoutingConfig{
			Rules: []config.RoutingRule{
				{Type: "field", InboundTag: []string{"api"}, OutboundTag: "api"},
			},
		},
	}

	cfg.Inbounds = append(cfg.Inbounds, apiInbound(opts.APIListen, opts.APIPort))

	usersByInbound := groupUsers(users)
	for _, in := range inbounds {
		if !in.Enabled {
			continue
		}
		ib, err := buildInbound(in, usersByInbound[in.ID], levelOf)
		if err != nil {
			return nil, err
		}
		cfg.Inbounds = append(cfg.Inbounds, ib)
	}

	return cfg, nil
}

func apiInbound(listen string, port int) config.Inbound {
	settings, _ := json.Marshal(map[string]string{"address": "127.0.0.1"})
	return config.Inbound{
		Tag:      "api",
		Listen:   listen,
		Port:     port,
		Protocol: "dokodemo-door",
		Settings: settings,
	}
}

func buildInbound(in model.Inbound, users []model.User, levelOf func(model.User) int) (config.Inbound, error) {
	ib := config.Inbound{
		Tag:            in.Tag,
		Listen:         in.Listen,
		Port:           in.Port,
		Protocol:       in.Protocol,
		StreamSettings: buildStreamSettings(in),
	}
	if in.EnableSniffing {
		ib.Sniffing = &config.Sniffing{Enabled: true, DestOverride: []string{"http", "tls"}}
	}

	switch in.Protocol {
	case "vless":
		clients := make([]config.VLESSUser, 0, len(users))
		for _, u := range users {
			if !u.Enabled {
				continue
			}
			clients = append(clients, config.VLESSUser{
				ID:    u.UUID,
				Email: u.Email,
				Level: levelOf(u),
			})
		}
		settings, err := json.Marshal(config.VLESSInboundSettings{
			Clients:    clients,
			Decryption: "none",
		})
		if err != nil {
			return ib, err
		}
		ib.Settings = settings
	default:
		return ib, fmt.Errorf("unsupported protocol: %s", in.Protocol)
	}

	return ib, nil
}

func buildStreamSettings(in model.Inbound) *config.StreamSettings {
	ss := &config.StreamSettings{Network: in.Transport}

	if in.Transport == "ws" {
		parsed, _ := in.ParseStreamSettings()
		if parsed != nil {
			ws := &config.WSSettings{Path: parsed.Path, Headers: parsed.Headers}
			if ws.Path == "" {
				ws.Path = "/"
			}
			ss.WSSettings = ws
		}
	}

	if in.TLSEnabled {
		ss.Security = "tls"
		ss.TLSSettings = &config.TLSSettings{
			ServerName: in.TLSServerName,
			Certificates: []config.Certificate{
				{CertificateFile: in.TLSCert, KeyFile: in.TLSKey},
			},
		}
	}

	return ss
}

func groupUsers(users []model.User) map[int64][]model.User {
	m := make(map[int64][]model.User)
	for _, u := range users {
		m[u.InboundID] = append(m[u.InboundID], u)
	}
	return m
}

// buildSpeedLevels 为不同的限速组合分配 policy level，并返回 level 查询函数。
func buildSpeedLevels(users []model.User) (map[string]*config.PolicyLevel, func(model.User) int) {
	levels := map[string]*config.PolicyLevel{
		"0": {StatsUserUplink: true, StatsUserDownlink: true},
	}
	index := map[string]int{}
	next := 0

	levelOf := func(u model.User) int {
		if u.SpeedLimitUplink == 0 && u.SpeedLimitDownlink == 0 {
			return 0
		}
		key := fmt.Sprintf("%d:%d", u.SpeedLimitUplink, u.SpeedLimitDownlink)
		if lvl, ok := index[key]; ok {
			return lvl
		}
		next++
		index[key] = next
		levels[strconv.Itoa(next)] = &config.PolicyLevel{
			UplinkOnly:        int(u.SpeedLimitUplink),
			DownlinkOnly:      int(u.SpeedLimitDownlink),
			StatsUserUplink:   true,
			StatsUserDownlink: true,
		}
		return next
	}

	// 预填充 level，保证 levelOf 在任意调用顺序下结果一致。
	for _, u := range users {
		levelOf(u)
	}

	return levels, levelOf
}
