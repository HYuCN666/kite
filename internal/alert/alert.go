package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Config 为告警通知配置。
type Config struct {
	Webhook       string
	TelegramToken string
	TelegramChat  string
}

// Notify 发送告警通知（Webhook / Telegram）。
func Notify(cfg Config, title, message string) {
	text := fmt.Sprintf("[Volans] %s\n%s", title, message)
	if cfg.Webhook != "" {
		body, _ := json.Marshal(map[string]string{"title": title, "message": message, "text": text})
		go post(cfg.Webhook, body)
	}
	if cfg.TelegramToken != "" && cfg.TelegramChat != "" {
		url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", cfg.TelegramToken)
		body, _ := json.Marshal(map[string]string{"chat_id": cfg.TelegramChat, "text": text})
		go post(url, body)
	}
}

func post(url string, body []byte) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return
	}
	_ = resp.Body.Close()
}
