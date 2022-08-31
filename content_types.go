package withttp

import (
	"github.com/pkg/errors"
	"github.com/sonirico/withttp/codec"
)

type (
	ContentType string
)

var (
	ContentTypeJSON        ContentType = "application/json"
	ContentTypeJSONEachRow ContentType = "application/jsoneachrow"
)

var (
	ErrUnknownContentType = errors.New("unknown content type")
)

func (c ContentType) String() string {
	return string(c)
}

func (c ContentType) Codec() (codec.Codec, error) {
	switch c {
	case ContentTypeJSON:
		return codec.NativeJSONCodec, nil
	case ContentTypeJSONEachRow:
		return codec.NativeJSONEachRowCodec, nil
	default:
		return nil, errors.Wrapf(ErrUnknownContentType, "got: '%s'", c.String())
	}
}
