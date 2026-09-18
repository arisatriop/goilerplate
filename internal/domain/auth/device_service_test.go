package auth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const chromeOnMac = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0 Safari/537.36"

// The client IP is whatever the delivery layer resolved; the domain does not second-guess it.
// Trusting a forwarding header is a transport decision, made in bootstrap's trusted-proxy list.
func TestDeviceService_ExtractDeviceInfo_UsesResolvedClientIP(t *testing.T) {
	service := NewDeviceService()

	info := service.ExtractDeviceInfo(DeviceRequest{ClientIP: "203.0.113.7", UserAgent: chromeOnMac})

	assert.Equal(t, "203.0.113.7", info.IPAddress)
}

func TestDeviceService_ExtractDeviceInfo_TypeAndName(t *testing.T) {
	tests := []struct {
		userAgent string
		wantType  string
		wantName  string
	}{
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) Mobile/15E148", DeviceTypeMobile, "iPhone"},
		{"Mozilla/5.0 (Linux; Android 14; Pixel 8) Mobile", DeviceTypeMobile, "Android Phone"},
		{"Mozilla/5.0 (iPad; CPU OS 17_0 like Mac OS X)", DeviceTypeTablet, "iPad"},
		{"MyApp Electron/30.0 (Windows NT 10.0)", DeviceTypeDesktop, "Windows PC"},
		{chromeOnMac, DeviceTypeWeb, "Chrome Browser"},
		{"", DeviceTypeWeb, "Web Browser"},
	}

	service := NewDeviceService()
	for _, tt := range tests {
		info := service.ExtractDeviceInfo(DeviceRequest{UserAgent: tt.userAgent, ClientIP: "10.0.0.9"})
		assert.Equal(t, tt.wantType, info.DeviceType, tt.userAgent)
		assert.Equal(t, tt.wantName, info.DeviceName, tt.userAgent)
		assert.Equal(t, tt.userAgent, info.UserAgent)
	}
}

func TestDeviceService_ExtractDeviceInfo_Fingerprint(t *testing.T) {
	// Arrange
	service := NewDeviceService()
	base := DeviceRequest{UserAgent: chromeOnMac, AcceptLanguage: "en-US", AcceptEncoding: "gzip", ClientIP: "10.0.0.9"}
	otherIP := base
	otherIP.ClientIP = "10.0.0.1"
	otherLanguage := base
	otherLanguage.AcceptLanguage = "id-ID"

	// Act
	first := service.ExtractDeviceInfo(base).DeviceID
	again := service.ExtractDeviceInfo(base).DeviceID

	// Assert
	assert.True(t, strings.HasPrefix(first, "fp_"))
	assert.Len(t, first, 16)
	assert.Equal(t, first, again, "fingerprint is deterministic")
	assert.NotEqual(t, first, service.ExtractDeviceInfo(otherIP).DeviceID, "the resolved client IP is part of the fingerprint")
	assert.NotEqual(t, first, service.ExtractDeviceInfo(otherLanguage).DeviceID)
}
