package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The name arrives in a header or in gRPC metadata and is never verified, yet it lands in every
// log line the request produces. So it is bounded and stripped rather than taken as given.
func TestServiceName(t *testing.T) {
	long := strings.Repeat("a", 60)

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"empty falls back", "", DefaultServiceName},
		{"whitespace only falls back", "   ", DefaultServiceName},
		{"plain name", "billing-worker", "billing-worker"},
		{"dots and underscores kept", "billing_worker.v2", "billing_worker.v2"},
		{"surrounding space trimmed", "  billing-worker  ", "billing-worker"},
		{"newlines stripped", "billing\nworker", "billingworker"},
		{"quotes and braces stripped", `bill"ing{}`, "billing"},
		{"nothing usable falls back", `{"":}`, DefaultServiceName},
		{"over-long name truncated", long, long[:maxServiceNameLength]},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, ServiceName(tc.raw))
		})
	}
}
