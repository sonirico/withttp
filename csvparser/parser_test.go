package csvparser

import (
	"reflect"
	"testing"

	"github.com/pkg/errors"
)

func TestParser_RawColumn(t *testing.T) {
	type (
		args struct {
			payload []byte
			sep     byte
		}

		duck struct {
			Name     string
			Siblings int
		}

		want struct {
			expected duck
			err      error
		}

		testCase struct {
			name string
			args args
			want want
		}
	)

	var (
		duckNameSetter = func(d *duck, name string) {
			d.Name = name
		}
		duckSiblingsSetter = func(d *duck, siblings int) {
			d.Siblings = siblings
		}
	)

	tests := []testCase{
		{
			name: "simple csv string line should parse",
			args: args{
				payload: []byte("a duck knight in shinny armor,2"),
				sep:     SeparatorComma,
			},
			want: want{
				expected: duck{
					Name:     "a duck knight in shinny armor",
					Siblings: 2,
				},
			},
		},
		{
			name: "simple csv string line with trailing separator should parse",
			args: args{
				payload: []byte("a duck knight in shinny armor,2,"),
				sep:     SeparatorComma,
			},
			want: want{
				expected: duck{
					Name:     "a duck knight in shinny armor",
					Siblings: 2,
				},
			},
		},
		{
			name: "simple csv string line with trailing separator and spaces should parse",
			args: args{
				payload: []byte("a duck knight in shinny armor,2, 	"),
				sep:     SeparatorComma,
			},
			want: want{
				expected: duck{
					Name:     "a duck knight in shinny armor",
					Siblings: 2,
				},
			},
		},
		{
			name: "simple csv string line with trailing spaces at the start and spaces should parse",
			args: args{
				payload: []byte(" a duck knight in shinny armor,2, 	"),
				sep:     SeparatorComma,
			},
			want: want{
				expected: duck{
					Name:     "a duck knight in shinny armor",
					Siblings: 2,
				},
			},
		},
		{
			name: "blank column should render emptiness",
			args: args{
				payload: []byte(",2"),
				sep:     SeparatorComma,
			},
			want: want{
				expected: duck{
					Name:     "",
					Siblings: 2,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			parser := New[duck](
				test.args.sep,
				StringCol[duck](QuoteNone, nil, duckNameSetter),
				IntCol[duck](QuoteNone, nil, duckSiblingsSetter),
			)

			rubberDuck := duck{}

			if err := parser.Parse(test.args.payload, &rubberDuck); !errors.Is(test.want.err, err) {
				t.Errorf("unexpected error, want %v, have %v",
					test.want.err, err)
			}

			if !reflect.DeepEqual(test.want.expected, rubberDuck) {
				t.Errorf("unexpected duck\nwant %v\nhave %v",
					test.want.expected, rubberDuck)
			}

		})
	}
}
