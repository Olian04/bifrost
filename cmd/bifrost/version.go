package main

import (
	"runtime/debug"
)

var Version = "unknown"

func getVersion() string {
	// If Version was set via -ldflags, use it (for CI/CD builds)
	if Version != "unknown" {
		return Version
	}

	// Try to get version from build info (works with go install @version)
	if info, ok := debug.ReadBuildInfo(); ok {
		// Check build settings for VCS tag first (most reliable for git tags)
		for _, setting := range info.Settings {
			if setting.Key == "vcs.tag" && setting.Value != "" {
				return setting.Value
			}
		}

		if info.Main.Version != "" {
			return info.Main.Version
		}

		// Also check if this module appears in dependencies (for some build scenarios)
		for _, dep := range info.Deps {
			if dep.Path == "github.com/lolocompany/bifrost" {
				if dep.Version != "" && dep.Version != "(devel)" {
					return dep.Version
				}
			}
		}
	}

	return "unknown"
}
