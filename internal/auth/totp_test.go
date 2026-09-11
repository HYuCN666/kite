package auth

import "testing"

// RFC 6238 附录 B 测试向量（SHA1，6 位）。
func TestTOTPAt(t *testing.T) {
	key := []byte("12345678901234567890")
	cases := []struct {
		counter int64
		code    int
	}{
		{1, 287082},
		{37037036, 81804},
		{37037037, 50471},
		{41152263, 5924},
		{66666666, 279037},
		{666666666, 353130},
	}
	for _, c := range cases {
		if got := totpAt(key, c.counter); got != c.code {
			t.Errorf("counter %d: got %d, want %d", c.counter, got, c.code)
		}
	}
}
