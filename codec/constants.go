package codec

import "github.com/pkg/errors"

var (
	NativeJSONCodec        = NewNativeJsonCodec()
	NativeJSONEachRowCodec = NewNativeJsonEachRowCodec(NativeJSONCodec)
	ProxyBytesEncoder      = ProxyBytesCodec{}
)

var (
	ErrTypeAssertion = errors.New("unexpected type")
)
