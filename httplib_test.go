package trace

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReplyJSON(t *testing.T) {
	t.Parallel()

	var expectedErrorResponse = `{
		"error": {
			"message": "test error"
		}
	}`

	tests := []struct {
		desc string
		err  error
	}{
		{
			desc: "plain error",
			err:  errors.New("test error"),
		},
		{
			desc: "trace error",
			err:  &TraceErr{Err: errors.New("test error")},
		},
		{
			desc: "trace error with stacktrace",
			err:  &TraceErr{Err: errors.New("test error"), Traces: Traces{{Path: "A", Func: "B", Line: 1}}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			const errCode = 400
			replyJSON(recorder, errCode, tc.err)
			require.JSONEq(t, expectedErrorResponse, recorder.Body.String())
		})
	}
}

func TestUnmarshalError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		desc          string
		inputErr      error
		inputResponse string
		assertErr     func(error) bool
		expectedMsg   string
	}{
		{
			desc:          "unmarshal not found error",
			inputErr:      &NotFoundError{},
			inputResponse: `{"error": {"message": "ABC"}}`,
			assertErr:     IsNotFound,
			expectedMsg:   "ABC",
		},
		{
			desc:          "unmarshal access denied error",
			inputErr:      &AccessDeniedError{},
			inputResponse: `{"error": {"message": "ABC"}}`,
			assertErr:     IsAccessDenied,
			expectedMsg:   "ABC",
		},
		{
			desc:          "unmarshal error without error JSON key",
			inputErr:      &AccessDeniedError{},
			inputResponse: `{"message": "ABC"}`,
			assertErr:     IsAccessDenied,
			expectedMsg:   "ABC",
		},
		{
			desc:          "unmarshal invalid error",
			inputErr:      &AccessDeniedError{},
			inputResponse: `{"error": "message ABC"}`,
			assertErr:     IsAccessDenied,
			expectedMsg:   "{\"error\": \"message ABC\"}\n\taccess denied",
		},
		{
			desc:          "unmarshal invalid error without error JSON key",
			inputErr:      &AccessDeniedError{},
			inputResponse: `["error message ABC"]`,
			assertErr:     IsAccessDenied,
			expectedMsg:   "[\"error message ABC\"]\n\taccess denied",
		},
		{
			desc:          "unmarshal error with non-JSON body",
			inputErr:      &AccessDeniedError{},
			inputResponse: "error message ABC",
			assertErr:     IsAccessDenied,
			expectedMsg:   "error message ABC\n\taccess denied",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			readErr := unmarshalError(tc.inputErr, []byte(tc.inputResponse))
			require.True(t, tc.assertErr(readErr))
			require.EqualError(t, readErr, tc.expectedMsg)
		})
	}
}
