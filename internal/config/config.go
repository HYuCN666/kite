package config

import (
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
}

// Load 从命令行参数与环境变量解析配置。
func Load() (*Config, error) {
	var bind string
	var port int
	var dataDir string
	var xrayPath string
	var webDir string
	var secret string

	flag.StringVar(&bind, "bind", "127.0.0.1", "panel listen address")
	flag.IntVar(&port, "port", 8080, "panel listen port")
	flag.StringVar(&dataDir, "data", defaultDataDir(), "data directory")
	flag.StringVar(&xrayPath, "xray", "/usr/local/bin/xray", "xray binary path")
	flag.StringVar(&webDir, "web", "web/dist", "web dist directory")
	flag.StringVar(&secret, "secret", os.Getenv("KITE_SECRET"), "jwt signing secret")
	flag.Parse()

	if secret == "" {
		secret = "insecure-dev-secret"
	}

	return &Config{
		Bind:     bind,
		Port:     port,
		DataDir:  dataDir,
		XrayPath: xrayPath,
		WebDir:   webDir,
		Secret:   secret,
	}, nil
}

func defaultDataDir() string {
	if d := os.Getenv("KITE_DATA"); d != "" {
		return d
	}
	return filepath.Join(".", "data")
}
