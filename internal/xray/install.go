package xray

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
)

const latestURL = "https://api.github.com/repos/XTLS/Xray-core/releases/latest"

// Version 返回已安装内核的版本号。
func (m *Manager) Version() (string, error) {
	out, err := exec.Command(m.binPath, "version").Output()
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile(`(?i)xray\s+([0-9]+\.[0-9]+\.[0-9]+)`)
	if m := re.FindStringSubmatch(string(out)); m != nil {
		return m[1], nil
	}
	return "", fmt.Errorf("cannot parse version from %q", string(out))
}

// Install 下载并安装最新版 Xray 二进制。
func (m *Manager) Install() error {
	if err := os.MkdirAll(filepath.Dir(m.binPath), 0o755); err != nil {
		return err
	}

	asset, err := assetName()
	if err != nil {
		return err
	}
	url := fmt.Sprintf("https://github.com/XTLS/Xray-core/releases/latest/download/%s", asset)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	tmp, err := os.CreateTemp("", "xray-*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := io.Copy(tmp, resp.Body); err != nil {
		tmp.Close()
		return err
	}
	tmp.Close()

	return extractBinary(tmp.Name(), m.binPath)
}

func extractBinary(zipPath, binPath string) error {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer zr.Close()

	for _, f := range zr.File {
		if f.Name != "xray" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		out, err := os.OpenFile(binPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = io.Copy(out, rc)
		return err
	}
	return fmt.Errorf("xray binary not found in archive")
}

func assetName() (string, error) {
	switch {
	case runtime.GOOS == "linux" && runtime.GOARCH == "amd64":
		return "Xray-linux-64.zip", nil
	case runtime.GOOS == "linux" && runtime.GOARCH == "arm64":
		return "Xray-linux-arm64-v8a.zip", nil
	case runtime.GOOS == "windows" && runtime.GOARCH == "amd64":
		return "Xray-windows-64.zip", nil
	case runtime.GOOS == "darwin" && runtime.GOARCH == "amd64":
		return "Xray-macos-64.zip", nil
	case runtime.GOOS == "darwin" && runtime.GOARCH == "arm64":
		return "Xray-macos-arm64.zip", nil
	default:
		return "", fmt.Errorf("unsupported platform: %s/%s", runtime.GOOS, runtime.GOARCH)
	}
}
