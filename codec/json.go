package codec

import "encoding/json"

type (
	NativeJsonCodec struct{}
)

func (c NativeJsonCodec) Encode(t any) ([]byte, error) {
	return json.Marshal(t)
}

func (c NativeJsonCodec) Decode(data []byte, item any) (err error) {
	err = json.Unmarshal(data, item)
	return
}

func NewNativeJsonCodec() NativeJsonCodec {
	return NativeJsonCodec{}
}
