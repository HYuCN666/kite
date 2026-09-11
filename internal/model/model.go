package model

import "time"

// Inbound 对应一个 Xray 入站节点。
type Inbound struct {
	ID             int64     `json:"id"`
	Tag            string    `json:"tag"`
	Remark         string    `json:"remark"`
	Protocol       string    `json:"protocol"`
	Port           int       `json:"port"`
	Listen         string    `json:"listen"`
	Transport      string    `json:"transport"`
	StreamSettings string    `json:"stream_settings"`
	TLSEnabled     bool      `json:"tls_enabled"`
	TLSCert        string    `json:"tls_cert"`
	TLSKey         string    `json:"tls_key"`
	TLSServerName  string    `json:"tls_server_name"`
	EnableSniffing bool      `json:"enable_sniffing"`
	Enabled        bool      `json:"enabled"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// User 对应一个订阅用户。
type User struct {
	ID                 int64      `json:"id"`
	InboundID          int64      `json:"inbound_id"`
	Email              string     `json:"email"`
	UUID               string     `json:"uuid"`
	Remark             string     `json:"remark"`
	QuotaBytes         int64      `json:"quota_bytes"`
	UsedUplink         int64      `json:"used_uplink"`
	UsedDownlink       int64      `json:"used_downlink"`
	SpeedLimitUplink   int64      `json:"speed_limit_uplink"`
	SpeedLimitDownlink int64      `json:"speed_limit_downlink"`
	ExpireAt           *time.Time `json:"expire_at"`
	Enabled            bool       `json:"enabled"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// Admin 对应面板管理员。
type Admin struct {
	ID            int64      `json:"id"`
	Username      string     `json:"username"`
	PasswordHash  string     `json:"-"`
	LastLoginAt   *time.Time `json:"last_login_at"`
	FailedAttempt int        `json:"-"`
	LockedUntil   *time.Time `json:"-"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// Subscription 对应用户的订阅令牌。
type Subscription struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
}

// TrafficSnapshot 表示某个采集周期的流量增量。
type TrafficSnapshot struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	InboundID  int64     `json:"inbound_id"`
	Uplink     int64     `json:"uplink"`
	Downlink   int64     `json:"downlink"`
	RecordedAt time.Time `json:"recorded_at"`
}
