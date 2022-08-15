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
)
