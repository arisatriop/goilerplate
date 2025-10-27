// Package auth provides authentication domain services and entities.
package auth

import (
	"crypto/md5"
	"fmt"
	"net"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// DeviceInfo represents device information
type DeviceInfo struct {
	DeviceID   string
	DeviceType string
	DeviceName string
	IPAddress  string
	UserAgent  string
	Location   string
}

// DeviceService handles device detection and fingerprinting
type DeviceService interface {
	ExtractDeviceInfo(ctx *fiber.Ctx) *DeviceInfo
}

type deviceService struct{}

// NewDeviceService creates a new device service
func NewDeviceService() DeviceService {
	return &deviceService{}
}

// ExtractDeviceInfo extracts and generates device information from the request context
func (s *deviceService) ExtractDeviceInfo(ctx *fiber.Ctx) *DeviceInfo {
	userAgent := ctx.Get("User-Agent")
	ipAddress := s.getClientIP(ctx)
	deviceType := s.detectDeviceType(userAgent)

	return &DeviceInfo{
		DeviceID:   s.generateDeviceFingerprint(ctx),
		DeviceType: deviceType,
		DeviceName: s.generateDeviceName(deviceType, userAgent),
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		Location:   "", // TODO: Implement geolocation if needed
	}
}

// getClientIP extracts the real client IP from various headers
func (s *deviceService) getClientIP(ctx *fiber.Ctx) string {
	if xForwardedFor := ctx.Get("X-Forwarded-For"); xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	if xRealIP := ctx.Get("X-Real-IP"); xRealIP != "" {
		if net.ParseIP(xRealIP) != nil {
			return xRealIP
		}
	}

	return ctx.IP()
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
func (s *deviceService) generateDeviceFingerprint(ctx *fiber.Ctx) string {
	userAgent := ctx.Get("User-Agent")
	acceptLanguage := ctx.Get("Accept-Language")
	acceptEncoding := ctx.Get("Accept-Encoding")
	ip := s.getClientIP(ctx)

	fingerprint := fmt.Sprintf("%s|%s|%s|%s", userAgent, acceptLanguage, acceptEncoding, ip)
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
