package withttp

import (
	"github.com/pkg/errors"
	"github.com/sonirico/withttp/codec"
)

var (
	ContentTypeJSON        string = "application/json"
	ContentTypeJSONEachRow string = "application/jsoneachrow"
)

var (
	ErrUnknownContentType = errors.New("unknown content type")
)

func ContentTypeCodec(c string) (codec.Codec, error) {
	switch c {
	case ContentTypeJSON:
		return codec.NativeJSONCodec, nil
	case ContentTypeJSONEachRow:
		return codec.NativeJSONEachRowCodec, nil
	default:
		return nil, errors.Wrapf(ErrUnknownContentType, "got: '%s'", c)
	}
}
