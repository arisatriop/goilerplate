package utils

import "strings"

// DefaultServiceName is used when a caller does not name itself.
const DefaultServiceName = "system"

// maxServiceNameLength bounds what an unverified header can put into every log line of a request.
const maxServiceNameLength = 40

// ServiceName sanitises a caller-supplied service name, from the X-Service-Name header over HTTP
// or the same metadata key over gRPC.
//
// The name is an unverified claim. Nothing about a shared secret or an access token says which
// service is behind it, so this is for attribution in logs and never for authorization. Because
// it lands in every log line the request produces, it is bounded and stripped to a plain
// character set rather than taken as given.
func ServiceName(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return DefaultServiceName
	}

	cleaned := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			return r
		case r == '-', r == '_', r == '.':
			return r
		default:
			return -1
		}
	}, raw)

	if cleaned == "" {
		return DefaultServiceName
	}
	if len(cleaned) > maxServiceNameLength {
		cleaned = cleaned[:maxServiceNameLength]
	}

	return cleaned
}
