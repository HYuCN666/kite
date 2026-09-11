package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// GenerateTOTPSecret 生成 base32 编码的 TOTP 密钥。
func (m *Manager) GenerateTOTPSecret() string {
	b := make([]byte, 20)
	_, _ = rand.Read(b)
	return base32.StdEncoding.EncodeToString(b)
}

// ValidateTOTP 校验 TOTP 动态口令，skew 为允许的时间步偏移。
func (m *Manager) ValidateTOTP(secret, code string, skew int) bool {
	key, err := base32.StdEncoding.DecodeString(strings.ToUpper(secret))
	if err != nil || len(key) == 0 {
		return false
	}
	n, err := strconv.Atoi(code)
	if err != nil {
		return false
	}
	now := time.Now().Unix() / 30
	for i := int64(-skew); i <= int64(skew); i++ {
		if totpAt(key, now+i) == n {
			return true
		}
	}
	return false
}

// TOTPURL 生成 otpauth:// 地址，供二维码展示。
func (m *Manager) TOTPURL(secret, issuer, account string) string {
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s", issuer, account, secret, issuer)
}

func totpAt(key []byte, counter int64) int {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(counter))
	mac := hmac.New(sha1.New, key)
	_, _ = mac.Write(b[:])
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	code := (binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff) % 1000000
	return int(code)
}
