package model

import "encoding/json"

// StreamSettings 为入站的传输层参数，序列化后存入 inbounds.stream_settings 列。
type StreamSettings struct {
	Path    string            `json:"path,omitempty"`
	Host    string            `json:"host,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// ParseStreamSettings 解析入站的流设置 JSON，空时返回零值。
func (in *Inbound) ParseStreamSettings() (*StreamSettings, error) {
	ss := &StreamSettings{}
	if in.StreamSettings == "" {
		return ss, nil
	}
	if err := json.Unmarshal([]byte(in.StreamSettings), ss); err != nil {
		return nil, err
	}
	return ss, nil
}
