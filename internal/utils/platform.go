package utils

import "runtime"

// Platform constants matching Go's GOOS values
const (
	PlatformDarwin  = "darwin"
	PlatformLinux   = "linux"
	PlatformWindows = "windows"
)

// AllPlatforms returns all supported platforms
var AllPlatforms = []string{PlatformDarwin, PlatformLinux, PlatformWindows}

// DesktopPlatforms returns platforms that support desktop applications
var DesktopPlatforms = []string{PlatformDarwin, PlatformLinux, PlatformWindows}

// CurrentPlatform returns the current operating system platform
func CurrentPlatform() string {
	return runtime.GOOS
}

// IsPlatformSupported checks if the current platform is in the supported platforms list
func IsPlatformSupported(supportedPlatforms []string) bool {
	current := CurrentPlatform()
	for _, platform := range supportedPlatforms {
		if platform == current {
			return true
		}
	}
	return false
}

// IsPlatformOneOf checks if the current platform matches any of the given platforms
func IsPlatformOneOf(platforms ...string) bool {
	return IsPlatformSupported(platforms)
}

// GetPlatformDisplayName returns a user-friendly name for the platform
func GetPlatformDisplayName(platform string) string {
	switch platform {
	case PlatformDarwin:
		return "macOS"
	case PlatformLinux:
		return "Linux"
	case PlatformWindows:
		return "Windows"
	default:
		return platform
	}
}

// FilterPlatformSpecific filters a map based on current platform
// The map should have platform names as keys
func FilterPlatformSpecific(platformMap map[string]interface{}) interface{} {
	current := CurrentPlatform()
	if value, exists := platformMap[current]; exists {
		return value
	}
	return nil
}