package acme

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/crypto/acme/autocert"
)

// Issuer 通过 ACME（Let's Encrypt）为域名签发证书。
type Issuer struct {
	manager *autocert.Manager
	dir     string
}

// New 创建签发器，证书缓存于 cacheDir。
func New(cacheDir string) *Issuer {
	m := &autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache(cacheDir),
		HostPolicy: func(_ context.Context, _ string) error {
			return nil
		},
	}
	return &Issuer{manager: m, dir: cacheDir}
}

// HTTPHandler 返回 ACME HTTP-01 挑战处理器。
func (i *Issuer) HTTPHandler() http.Handler {
	return i.manager.HTTPHandler(nil)
}

// Issue 为域名签发证书，写入 PEM 文件并返回证书/私钥路径。
func (i *Issuer) Issue(domain string) (certPath, keyPath string, err error) {
	cert, err := i.manager.GetCertificate(&tls.ClientHelloInfo{ServerName: domain})
	if err != nil {
		return "", "", err
	}

	certPath = filepath.Join(i.dir, domain+".crt")
	keyPath = filepath.Join(i.dir, domain+".key")

	var certPEM bytes.Buffer
	for _, der := range cert.Certificate {
		_ = pem.Encode(&certPEM, &pem.Block{Type: "CERTIFICATE", Bytes: der})
	}

	keyDER, err := x509.MarshalPKCS8PrivateKey(cert.PrivateKey)
	if err != nil {
		return "", "", err
	}
	var keyPEM bytes.Buffer
	_ = pem.Encode(&keyPEM, &pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})

	if err := os.WriteFile(certPath, certPEM.Bytes(), 0o644); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(keyPath, keyPEM.Bytes(), 0o600); err != nil {
		return "", "", err
	}
	return certPath, keyPath, nil
}
