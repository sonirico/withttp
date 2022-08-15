package withttp

import "github.com/pkg/errors"

var (
	ErrAssertion            = errors.New("assertion was unmet")
	ErrUnexpectedStatusCode = errors.Wrap(ErrAssertion, "unexpected status code")
)
