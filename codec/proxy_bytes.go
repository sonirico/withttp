package codec

import "github.com/pkg/errors"

type (
	ProxyBytesCodec struct{}
)

func (e ProxyBytesCodec) Encode(x any) ([]byte, error) {
	bts, ok := x.([]byte)
	if !ok {
		return nil, errors.Wrapf(ErrTypeAssertion, "want '[]byte', have %T", x)
	}

	return bts, nil
}

func (e ProxyBytesCodec) Decode(_ []byte, _ any) error {
	panic("not implemented")
}
