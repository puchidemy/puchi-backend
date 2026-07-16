package biz

import "strings"

// DeviceInfo holds parsed information about a client device.
type DeviceInfo struct {
	Name string
	Type string // "mobile", "desktop", "tablet", "unknown"
	OS   string
}

// ParseDeviceInfo parses a User-Agent string and returns structured device info.
func ParseDeviceInfo(userAgent string) DeviceInfo {
	ua := strings.ToLower(userAgent)
	info := DeviceInfo{Name: userAgent, Type: "unknown", OS: "unknown"}

	switch {
	case strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad"):
		info.Type = "mobile"
		info.OS = "iOS"
	case strings.Contains(ua, "android"):
		info.Type = "mobile"
		info.OS = "Android"
	case strings.Contains(ua, "windows"):
		info.OS = "Windows"
		if strings.Contains(ua, "phone") {
			info.Type = "mobile"
		} else {
			info.Type = "desktop"
		}
	case strings.Contains(ua, "macintosh") || strings.Contains(ua, "mac os"):
		info.OS = "macOS"
		info.Type = "desktop"
	case strings.Contains(ua, "linux"):
		info.OS = "Linux"
		info.Type = "desktop"
	case strings.Contains(ua, "crkey") || strings.Contains(ua, "crkeyos"):
		info.OS = "ChromeOS"
		info.Type = "desktop"
	}

	// Determine device name from browser
	switch {
	case strings.Contains(ua, "firefox"):
		info.Name = "Firefox"
	case strings.Contains(ua, "chrome"):
		info.Name = "Chrome"
	case strings.Contains(ua, "safari"):
		info.Name = "Safari"
	case strings.Contains(ua, "edge"):
		info.Name = "Edge"
	default:
		info.Name = "Unknown Browser"
	}

	return info
}
