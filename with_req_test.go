package withttp

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
)

func TestCall_StreamingRequestFromSlice(t *testing.T) {
	type (
		payload struct {
			Name string `json:"name"`
		}

		args struct {
			Stream []payload
		}

		want struct {
			expectedErr     error
			ReceivedPayload []byte
		}

		testCase struct {
			name string
			args args
			want want
		}
	)

	tests := []testCase{
		{
			name: "one element in the stream",
			args: args{
				Stream: []payload{
					{
						Name: "I am the first payload",
					},
				},
			},
			want: want{
				ReceivedPayload: []byte(`{"name":"I am the first payload"}`),
			},
		},
		{
			name: "several elements in the stream",
			args: args{
				Stream: []payload{
					{
						Name: "I am the first payload",
					},
					{
						Name: "I am the second payload",
					},
				},
			},
			want: want{
				ReceivedPayload: streamTextJoin("\n", []string{
					`{"name":"I am the first payload"}`,
					`{"name":"I am the second payload"}`,
				}),
			},
		},
	}

	endpoint := NewEndpoint("mock").
		Response(MockedRes(func(res Response) {
			res.SetStatus(http.StatusOK)
			res.SetBody(io.NopCloser(bytes.NewReader(nil)))
		}))

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {

			call := NewCall[any](NewMockHttpClientAdapter()).
				ContentType(ContentTypeJSONEachRow).
				RequestStreamBody(
					RequestStreamBody[any, payload](Slice[payload](test.args.Stream)),
				).
				ExpectedStatusCodes(http.StatusOK)

			err := call.CallEndpoint(context.TODO(), endpoint)

			if !assertError(t, test.want.expectedErr, err) {
				t.FailNow()
			}
			actualReceivedBody := call.Req.Body()

			if !BytesEquals(test.want.ReceivedPayload, actualReceivedBody) {
				t.Errorf("unexpected received payload\nwant '%s'\nhave '%s'",
					string(test.want.ReceivedPayload), string(actualReceivedBody))
			}
		})
	}
}
