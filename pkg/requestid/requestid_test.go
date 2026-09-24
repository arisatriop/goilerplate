package requestid_test

import (
	"strings"
	"testing"

	"goilerplate/pkg/requestid"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolve_KeepsAWellFormedID(t *testing.T) {
	for _, id := range []string{
		"0190a6f0-0000-7000-8000-000000000001",
		"01ARZ3NDEKTSV4RRFFQ69G5FAV",
		"4bf92f3577b34da6a3ce929d0e0e4736",
		"req_abc.123-XYZ",
		strings.Repeat("a", requestid.MaxLength),
	} {
		assert.Equal(t, id, requestid.Resolve(id), id)
	}
}

// Each of these would otherwise land verbatim in every log line of the request and in the
// response header.
func TestResolve_ReplacesAnUnsafeID(t *testing.T) {
	tests := map[string]string{
		"empty":             "",
		"too long":          strings.Repeat("a", requestid.MaxLength+1),
		"newline":           "abc\nlevel=ERROR msg=forged",
		"carriage return":   "abc\r\nSet-Cookie: x=y",
		"space":             "abc def",
		"quote":             `abc"}`,
		"non-ascii":         "réquest",
		"control character": "abc\x00",
	}

	for name, incoming := range tests {
		t.Run(name, func(t *testing.T) {
			got := requestid.Resolve(incoming)

			assert.NotEqual(t, incoming, got)
			parsed, err := uuid.Parse(got)
			require.NoError(t, err, "a replacement must be a generated ID")
			assert.Equal(t, uuid.Version(7), parsed.Version())
		})
	}
}
