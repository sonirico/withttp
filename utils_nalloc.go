package withttp

import (
	"reflect"
	"unsafe"
)

func S2B(s string) (bts []byte) {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&bts))
	bh.Data = sh.Data
	bh.Cap = sh.Len
	bh.Len = sh.Len
	return
}

func B2S(data []byte) string {
	return *(*string)(unsafe.Pointer(&data))
}
