// Package auth provides authentication domain services and entities.
package auth

import (
	"crypto/md5"
	"fmt"
	"net"
	"strings"
)

// DeviceInfo represents device information
type DeviceInfo struct {
	DeviceID   string
	DeviceType string
	DeviceName string
	IPAddress  string
	UserAgent  string
}

// DeviceRequest carries the request attributes used to identify a device.
// Delivery layers (HTTP, gRPC) build it from their own request types.
type DeviceRequest struct {
	UserAgent      string
	AcceptLanguage string
	AcceptEncoding string
	ForwardedFor   string // X-Forwarded-For header
	RealIP         string // X-Real-IP header
	RemoteIP       string // peer address as seen by the server
}

// DeviceService handles device detection and fingerprinting
type DeviceService interface {
	ExtractDeviceInfo(req DeviceRequest) *DeviceInfo
}

type deviceService struct{}

// NewDeviceService creates a new device service
func NewDeviceService() DeviceService {
	return &deviceService{}
}

// ExtractDeviceInfo extracts and generates device information from the request attributes
func (s *deviceService) ExtractDeviceInfo(req DeviceRequest) *DeviceInfo {
	ipAddress := s.getClientIP(req)
	deviceType := s.detectDeviceType(req.UserAgent)

	return &DeviceInfo{
		DeviceID:   s.generateDeviceFingerprint(req, ipAddress),
		DeviceType: deviceType,
		DeviceName: s.generateDeviceName(deviceType, req.UserAgent),
		IPAddress:  ipAddress,
		UserAgent:  req.UserAgent,
	}
}

// getClientIP extracts the real client IP from forwarding headers, falling back to the peer address
func (s *deviceService) getClientIP(req DeviceRequest) string {
	if req.ForwardedFor != "" {
		ips := strings.Split(req.ForwardedFor, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	if req.RealIP != "" {
		if net.ParseIP(req.RealIP) != nil {
			return req.RealIP
		}
	}

	return req.RemoteIP
}

// detectDeviceType determines the device type from user agent
func (s *deviceService) detectDeviceType(userAgent string) string {
	userAgent = strings.ToLower(userAgent)

	if strings.Contains(userAgent, "mobile") || strings.Contains(userAgent, "android") || strings.Contains(userAgent, "iphone") {
		return DeviceTypeMobile
	}
	if strings.Contains(userAgent, "tablet") || strings.Contains(userAgent, "ipad") {
		return DeviceTypeTablet
	}
	if strings.Contains(userAgent, "electron") || strings.Contains(userAgent, "desktop") {
		return DeviceTypeDesktop
	}

	return DeviceTypeWeb
}

// generateDeviceFingerprint creates a unique device identifier
func (s *deviceService) generateDeviceFingerprint(req DeviceRequest, ip string) string {
	fingerprint := fmt.Sprintf("%s|%s|%s|%s", req.UserAgent, req.AcceptLanguage, req.AcceptEncoding, ip)
	hash := md5.Sum([]byte(fingerprint))
	return fmt.Sprintf("fp_%x", hash)[:16]
}

// generateDeviceName creates a human-readable device name
func (s *deviceService) generateDeviceName(deviceType, userAgent string) string {
	userAgent = strings.ToLower(userAgent)

	switch deviceType {
	case DeviceTypeMobile:
		if strings.Contains(userAgent, "iphone") {
			return "iPhone"
		}
		if strings.Contains(userAgent, "android") {
			return "Android Phone"
		}
		return "Mobile Device"

	case DeviceTypeTablet:
		if strings.Contains(userAgent, "ipad") {
			return "iPad"
		}
		return "Tablet"

	case DeviceTypeDesktop:
		if strings.Contains(userAgent, "windows") {
			return "Windows PC"
		}
		if strings.Contains(userAgent, "macintosh") || strings.Contains(userAgent, "mac os") {
			return "Mac"
		}
		if strings.Contains(userAgent, "linux") {
			return "Linux PC"
		}
		return "Desktop"

	case DeviceTypeWeb:
		if strings.Contains(userAgent, "chrome") {
			return "Chrome Browser"
		}
		if strings.Contains(userAgent, "firefox") {
			return "Firefox Browser"
		}
		if strings.Contains(userAgent, "safari") && !strings.Contains(userAgent, "chrome") {
			return "Safari Browser"
		}
		if strings.Contains(userAgent, "edge") {
			return "Edge Browser"
		}
		return "Web Browser"

	default:
		return "Unknown Device"
	}
}
