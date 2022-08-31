package codec

const (
	LN = byte('\n')
)

type (
	NativeJsonEachRowCodec struct {
		NativeJsonCodec
	}
)

func (c NativeJsonEachRowCodec) Encode(t any) (bts []byte, err error) {
	bts, err = c.NativeJsonCodec.Encode(t)
	if err != nil {
		return
	}

	bts = append(bts, LN)
	return
}

func NewNativeJsonEachRowCodec(inner NativeJsonCodec) NativeJsonEachRowCodec {
	return NativeJsonEachRowCodec{NativeJsonCodec: inner}
}
