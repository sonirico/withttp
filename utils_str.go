package withttp

import (
	"bytes"
	"strings"
)

func StrIsset(s string) bool {
	return len(strings.TrimSpace(s)) > 0
}

func BtsIsset(bts []byte) bool {
	return len(bytes.TrimSpace(bts)) > 0
}

func BytesEquals(a, b []byte) bool {
	return bytes.Compare(bytes.TrimSpace(a), bytes.TrimSpace(b)) == 0
}
