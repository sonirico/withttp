package withttp

import (
	"github.com/pkg/errors"
	"github.com/sonirico/withttp/codec"
	"strings"
)

type (
	ContentType string
)

var (
	ContentTypeJSON = "application/json"
)

var (
	ErrUnknownContentType = errors.New("unknown content type")
)

func (c ContentType) String() string {
	return string(c)
}

func (c ContentType) IsJSON() bool {
	lower := strings.ToLower(strings.TrimSpace(c.String()))
	hasApp := strings.Contains(lower, "application")
	hasJSon := strings.Contains(lower, "json")
	return hasApp && hasJSon
}

func (c ContentType) Codec() (codec.Codec, error) {
	switch {
	case c.IsJSON():
		return codec.NativeJSONCodec, nil
	default:
		return nil, errors.Wrapf(ErrUnknownContentType, "got: '%s'", c.String())
	}
}
