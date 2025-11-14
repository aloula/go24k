package main

import (
	"flag"
	"fmt"
	"time"

	"go24k/utils"
)

func main() {
	// Set up command-line flags.
	convertOnly := flag.Bool("convert-only", false, "Convert images only, without generating the video")
	static := flag.Bool("static", false, "Do NOT apply Ken Burns effect; use static images with transitions")
	duration := flag.Int("d", 5, "Duration per image in seconds")
	transition := flag.Int("t", 1, "Transition (fade) duration in seconds")
	debug := flag.Bool("debug", false, "Show environment detection and optimization info")
	exifOverlay := flag.Bool("exif-overlay", false, "Add camera info overlay to video (bottom center)")
	overlayFontSize := flag.Int("overlay-font-size", 36, "Font size for EXIF overlay (default: 36)")
	version := flag.Bool("version", false, "Show version information")
	versionShort := flag.Bool("v", false, "Show version information (short)")
	help := flag.Bool("help", false, "Show this help message")

	// Custom usage function
	flag.Usage = func() {
		fmt.Printf("%s\n\n", GetVersionInfo())
		fmt.Printf("USAGE:\n")
		fmt.Printf("  %s [OPTIONS]\n\n", "go24k")
		fmt.Printf("OPTIONS:\n")
		flag.PrintDefaults()
		fmt.Printf("\nEXAMPLES:\n")
		fmt.Printf("  go24k                                      # Create 4K video with default settings\n")
		fmt.Printf("  go24k -d 8 -t 2                            # 8s per image, 2s transitions\n")
		fmt.Printf("  go24k -static                              # Disable Ken Burns effect\n")
		fmt.Printf("  go24k -exif-overlay                        # Add camera info overlay\n")
		fmt.Printf("  go24k -exif-overlay -overlay-font-size 48  # Large font overlay\n")
		fmt.Printf("  go24k -convert-only                        # Only convert images to 4K\n")
		fmt.Printf("  go24k -debug                               # Show hardware detection info\n")
		fmt.Printf("\nFor more information: https://github.com/aloula/go24k\n")
	}

	flag.Parse()

	// Show help if requested
	if *help {
		flag.Usage()
		return
	}

	// Show version info if requested
	if *version || *versionShort {
		if *version {
			fmt.Println(GetFullVersionInfo())
		} else {
			fmt.Println(GetVersionInfo())
		}
		return
	}

	// Show version on startup (brief)
	fmt.Printf("🎬 %s\n", GetVersionInfo())

	// Show debug info if requested
	if *debug {
		utils.ShowEnvironmentInfo()
		return
	}

	startTime := time.Now()

	// Convert images (e.g. scale, add background, overlay, etc.)
	if err := utils.ConvertImages(); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Generate video only if convert-only is not set.
	if !*convertOnly {
		// If -static is provided, applyKenBurns will be false.
		applyKenBurns := !*static
		// Pass the duration and transition values from the flags.
		utils.GenerateVideo(*duration, *transition, applyKenBurns, *exifOverlay, *overlayFontSize)
	}

	// Report processing time (only if not convert-only since conversion already shows its time)
	if !*convertOnly {
		elapsedTime := time.Since(startTime).Seconds()
		fmt.Printf("Total time: %.1f sec.\n", elapsedTime)
	}
}
