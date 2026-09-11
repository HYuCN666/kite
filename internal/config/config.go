package config

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"os"
	"path/filepath"
)

// Config 保存面板运行时的可配置项。
type Config struct {
	Bind     string
	Port     int
	DataDir  string
	XrayPath string
	WebDir   string
	Secret   string
	TLSCert  string
	TLSKey   string
}

// Load 从命令行参数与环境变量解析配置。
func Load() (*Config, error) {
	var bind string
	var port int
	var dataDir string
	var xrayPath string
	var webDir string
	var secret string
	var tlsCert string
	var tlsKey string

	flag.StringVar(&bind, "bind", "127.0.0.1", "panel listen address")
	flag.IntVar(&port, "port", 8080, "panel listen port")
	flag.StringVar(&dataDir, "data", defaultDataDir(), "data directory")
	flag.StringVar(&xrayPath, "xray", "/usr/local/bin/xray", "xray binary path")
	flag.StringVar(&webDir, "web", "web/dist", "web dist directory")
	flag.StringVar(&secret, "secret", os.Getenv("KITE_SECRET"), "jwt signing secret")
	flag.StringVar(&tlsCert, "tls-cert", os.Getenv("KITE_TLS_CERT"), "panel tls certificate path")
	flag.StringVar(&tlsKey, "tls-key", os.Getenv("KITE_TLS_KEY"), "panel tls key path")
	flag.Parse()

	if secret == "" {
		generated, err := loadOrCreateSecret(dataDir)
		if err != nil {
			return nil, err
		}
		secret = generated
	}

	return &Config{
		Bind:     bind,
		Port:     port,
		DataDir:  dataDir,
		XrayPath: xrayPath,
		WebDir:   webDir,
		Secret:   secret,
		TLSCert:  tlsCert,
		TLSKey:   tlsKey,
	}, nil
}

// loadOrCreateSecret 读取持久化密钥，不存在则生成并写入 dataDir/.secret。
func loadOrCreateSecret(dataDir string) (string, error) {
	path := filepath.Join(dataDir, ".secret")
	if b, err := os.ReadFile(path); err == nil && len(b) >= 32 {
		return string(b), nil
	}

	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return "", err
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	secret := hex.EncodeToString(raw)

	if err := os.WriteFile(path, []byte(secret), 0o600); err != nil {
		return "", err
	}
	return secret, nil
}

func defaultDataDir() string {
	if d := os.Getenv("KITE_DATA"); d != "" {
		return d
	}
	return filepath.Join(".", "data")
}
