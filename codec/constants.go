package codec

var (
	NativeJSONCodec        = NewNativeJsonCodec()
	NativeJSONEachRowCodec = NewNativeJsonEachRowCodec(NativeJSONCodec)
)
