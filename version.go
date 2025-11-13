package main

import (
	"fmt"
	"runtime"
)

// Version information - update these for each release
const (
	Version     = "1.0"
	BuildDate   = "2025-11-13"
	Description = "4K Video Creator with Hardware Acceleration & EXIF Overlay"
)

// GetVersionInfo returns formatted version information
func GetVersionInfo() string {
	return fmt.Sprintf("Go24K v%s", Version)
}

// GetFullVersionInfo returns detailed version information
func GetFullVersionInfo() string {
	return fmt.Sprintf(`Go24K v%s
%s

Built: %s
Runtime: %s %s/%s
Go Version: %s

A powerful 4K video creator with hardware acceleration support.
Converts JPEG images to stunning 4K videos with Ken Burns effects,
crossfade transitions, and automatic EXIF overlay functionality.

Copyright (c) 2025 - https://github.com/aloula/go24k`,
		Version,
		Description,
		BuildDate,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
		runtime.Version(),
	)
}

// GetBuildInfo returns build information for executables
func GetBuildInfo() string {
	return fmt.Sprintf("Go24K-%s-%s-%s", Version, runtime.GOOS, runtime.GOARCH)
}
