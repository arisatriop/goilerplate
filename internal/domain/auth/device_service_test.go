package auth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const chromeOnMac = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0 Safari/537.36"

func TestDeviceService_ExtractDeviceInfo_ClientIP(t *testing.T) {
	tests := []struct {
		name string
		req  DeviceRequest
		want string
	}{
		{"first forwarded address", DeviceRequest{ForwardedFor: "203.0.113.7, 10.0.0.1", RealIP: "198.51.100.2", RemoteIP: "10.0.0.9"}, "203.0.113.7"},
		{"real ip when forwarded is invalid", DeviceRequest{ForwardedFor: "not-an-ip", RealIP: "198.51.100.2", RemoteIP: "10.0.0.9"}, "198.51.100.2"},
		{"remote ip without headers", DeviceRequest{RemoteIP: "10.0.0.9"}, "10.0.0.9"},
		{"invalid real ip falls back to remote", DeviceRequest{RealIP: "bogus", RemoteIP: "10.0.0.9"}, "10.0.0.9"},
	}

	service := NewDeviceService()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, service.ExtractDeviceInfo(tt.req).IPAddress)
		})
	}
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
		info := service.ExtractDeviceInfo(DeviceRequest{UserAgent: tt.userAgent, RemoteIP: "10.0.0.9"})
		assert.Equal(t, tt.wantType, info.DeviceType, tt.userAgent)
		assert.Equal(t, tt.wantName, info.DeviceName, tt.userAgent)
		assert.Equal(t, tt.userAgent, info.UserAgent)
	}
}

func TestDeviceService_ExtractDeviceInfo_Fingerprint(t *testing.T) {
	// Arrange
	service := NewDeviceService()
	base := DeviceRequest{UserAgent: chromeOnMac, AcceptLanguage: "en-US", AcceptEncoding: "gzip", RemoteIP: "10.0.0.9"}
	sameClientViaProxy := base
	sameClientViaProxy.RemoteIP = "10.0.0.1"
	sameClientViaProxy.ForwardedFor = "10.0.0.9"
	otherLanguage := base
	otherLanguage.AcceptLanguage = "id-ID"

	// Act
	first := service.ExtractDeviceInfo(base).DeviceID
	again := service.ExtractDeviceInfo(base).DeviceID

	// Assert
	assert.True(t, strings.HasPrefix(first, "fp_"))
	assert.Len(t, first, 16)
	assert.Equal(t, first, again, "fingerprint is deterministic")
	assert.Equal(t, first, service.ExtractDeviceInfo(sameClientViaProxy).DeviceID, "fingerprint uses the resolved client IP")
	assert.NotEqual(t, first, service.ExtractDeviceInfo(otherLanguage).DeviceID)
}
