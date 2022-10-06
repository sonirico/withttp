package csvparser

import (
	"strings"
	"testing"

	"github.com/pkg/errors"
)

func TestStringType_Parse(t *testing.T) {
	type (
		args struct {
			payload []byte
			quote   byte
			sep     byte
		}

		want struct {
			expected []byte
			read     int
			err      error
		}

		testCase struct {
			name string
			args args
			want want
		}
	)

	tests := []testCase{
		{
			name: "unquoted simple string",
			args: args{
				quote:   QuoteNone,
				sep:     SeparatorComma,
				payload: []byte("fmartingr,danirod_,3"),
			},
			want: want{
				expected: []byte("fmartingr"),
				read:     9,
				err:      nil,
			},
		},
		{
			name: "double quote simple string",
			args: args{
				quote:   QuoteDouble,
				sep:     SeparatorComma,
				payload: []byte("\"fmartingr\",danirod_,3"),
			},
			want: want{
				expected: []byte("fmartingr"),
				read:     9,
				err:      nil,
			},
		},
		{
			name: "simple quote simple string",
			args: args{
				quote:   QuoteSimple,
				sep:     SeparatorComma,
				payload: []byte("'fmartingr',danirod_,3"),
			},
			want: want{
				expected: []byte("fmartingr"),
				read:     9,
				err:      nil,
			},
		},
		{
			name: "non quote non-ascii string",
			args: args{
				quote:   QuoteNone,
				sep:     SeparatorComma,
				payload: []byte("你好吗,danirod_,3"),
			},
			want: want{
				expected: []byte("你好吗"),
				read:     9,
				err:      nil,
			},
		},
		{
			name: "double quote non-ascii string",
			args: args{
				quote:   QuoteDouble,
				sep:     SeparatorComma,
				payload: []byte("\"你好吗\",danirod_,3"),
			},
			want: want{
				expected: []byte("你好吗"),
				read:     9,
				err:      nil,
			},
		},
		{
			name: "double quote non-ascii string with escaped char same as quote",
			args: args{
				quote:   QuoteDouble,
				sep:     SeparatorComma,
				payload: []byte("\"你\\\"好吗\",danirod_,3"),
			},
			want: want{
				expected: []byte("你\\\"好吗"),
				read:     11,
				err:      nil,
			},
		},
		{
			name: "double quote non-ascii string with escaped char same as quote and other char same as separator",
			args: args{
				quote:   QuoteDouble,
				sep:     SeparatorComma,
				payload: []byte("\"你\\\"好,吗\",danirod_,3"),
			},
			want: want{
				expected: []byte("你\\\"好,吗"),
				read:     12,
				err:      nil,
			},
		},
		{
			name: "simple quoted json",
			args: args{
				quote:   QuoteSimple,
				sep:     SeparatorComma,
				payload: []byte(`'{"name":"Pato","age":3}',danirod_,3`),
			},
			want: want{
				expected: []byte(`{"name":"Pato","age":3}`),
				read:     23,
				err:      nil,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			str := StrType(test.args.quote, test.args.sep)
			actual, read, actualErr := str.Parse(test.args.payload)
			if !errors.Is(test.want.err, actualErr) {
				t.Fatalf("unexpected error. want %v, have %v",
					test.want.err, actualErr)
			}

			if test.want.read != read {
				t.Fatalf("unexpected bytes read, want %d have %d",
					test.want.read, read)
			}

			if strings.Compare(string(test.want.expected), actual) != 0 {
				t.Fatalf("unexpected result. want %v, have %v",
					string(test.want.expected), actual)
			}
		})
	}
}
