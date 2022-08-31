package withttp

import (
	"github.com/pkg/errors"
	"strings"
	"testing"
)

func assertError(t *testing.T, expected, actual error) bool {
	t.Helper()

	if actual != nil {
		if expected != nil {
			if !errors.Is(expected, actual) {
				t.Errorf("unexpected error, want %s, have %s",
					expected, actual)
				return false
			}
		} else {
			t.Errorf("unexpected error, want none, have %s", actual)
			return false
		}
	} else {
		if expected != nil {
			t.Errorf("unexpected error, want %s, have none",
				expected)
			return false
		}
	}

	return true
}

func streamTextJoin(sep string, items []string) []byte {
	return []byte(strings.Join(items, sep))
}
