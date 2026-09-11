package sshx

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

// Config 为 SSH 连接配置。
type Config struct {
	Host       string
	Port       int
	Username   string
	AuthType   string // "password" | "key"
	Password   string
	PrivateKey string
}

// Client 为 SSH 执行客户端。
type Client struct {
	cfg    Config
	client *ssh.Client
}

// New 创建客户端。
func New(cfg Config) *Client {
	return &Client{cfg: cfg}
}

// Connect 建立 SSH 连接。
func (c *Client) Connect() error {
	if c.client != nil {
		return nil
	}

	var auth []ssh.AuthMethod
	switch c.cfg.AuthType {
	case "key":
		signer, err := ssh.ParsePrivateKey([]byte(c.cfg.PrivateKey))
		if err != nil {
			return fmt.Errorf("parse private key: %w", err)
		}
		auth = append(auth, ssh.PublicKeys(signer))
	default:
		auth = append(auth, ssh.Password(c.cfg.Password))
	}

	config := &ssh.ClientConfig{
		User:            c.cfg.Username,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := net.JoinHostPort(c.cfg.Host, fmt.Sprintf("%d", c.cfg.Port))
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return err
	}
	c.client = client
	return nil
}

// Close 关闭连接。
func (c *Client) Close() error {
	if c.client != nil {
		err := c.client.Close()
		c.client = nil
		return err
	}
	return nil
}

// Run 执行命令并返回合并的标准输出/错误。
func (c *Client) Run(cmd string) (string, error) {
	if err := c.Connect(); err != nil {
		return "", err
	}
	session, err := c.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var out bytes.Buffer
	session.Stdout = &out
	session.Stderr = &out
	if err := session.Run(cmd); err != nil {
		return out.String(), err
	}
	return out.String(), nil
}

// WriteFile 通过 SSH 将内容写入远端文件。
func (c *Client) WriteFile(path string, content []byte) error {
	if err := c.Connect(); err != nil {
		return err
	}
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdin = bytes.NewReader(content)
	return session.Run("mkdir -p $(dirname " + path + ") && cat > " + path)
}

// ForwardLocal 建立本地端口到远端地址的转发，返回本地监听地址。
func (c *Client) ForwardLocal(remoteAddr string) (string, error) {
	if err := c.Connect(); err != nil {
		return "", err
	}
	listener, err := c.client.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}

	go func() {
		for {
			local, err := listener.Accept()
			if err != nil {
				return
			}
			go func(local net.Conn) {
				defer local.Close()
				remote, err := c.client.Dial("tcp", remoteAddr)
				if err != nil {
					return
				}
				defer remote.Close()
				go func() { _, _ = io.Copy(remote, local) }()
				_, _ = io.Copy(local, remote)
			}(local)
		}
	}()

	return listener.Addr().String(), nil
}
