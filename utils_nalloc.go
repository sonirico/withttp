package withttp

import (
	"unsafe"
)

func S2B(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func B2S(data []byte) string {
	return unsafe.String(unsafe.SliceData(data), len(data))
}
