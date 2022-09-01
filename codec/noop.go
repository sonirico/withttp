package codec

type (
	Encoder interface {
		Encode(any) ([]byte, error)
	}

	Decoder interface {
		Decode([]byte, any) error
	}

	Codec interface {
		Encoder
		Decoder
	}

	NoopCodec struct{}
)

func (c NoopCodec) Encode(_ any) (bts []byte, err error) { return }
func (c NoopCodec) Decode(_ []byte, _ any) (err error)   { return }
