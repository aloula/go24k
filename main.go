package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go24k/utils"
)

func main() {
	// Set up command-line flags.
	duration := flag.Int("d", 5, "Duration per image in seconds")
	transition := flag.Int("t", 1, "Transition (fade) duration in seconds")
	fps := flag.Int("fps", 30, "Output framerate override: 30 or 60")
	fitAudio := flag.Bool("fit-audio", false, "Auto-fit image and transition durations to fill the music length")
	includeVideos := flag.Bool("include-videos", false, "Include supported video files (mp4, mov, mkv, avi, webm, m4v) together with pictures")
	keepVideoAudio := flag.Bool("keep-video-audio", false, "Keep input video audio and blend it with MP3 background audio")
	orderMode := flag.String("order", "metadata", "Timeline order: metadata, filename, or random")
	orderByFilename := flag.Bool("order-by-filename", false, "Order timeline by filename instead of metadata time")
	randomOrder := flag.Bool("random-order", false, "Order timeline randomly")
	fullHD := flag.Bool("fullhd", false, "Generate Full HD (1920x1080) video instead of 4K UHD (3840x2160)")
	effectsMode := flag.String("effects", "disabled", "Image motion effects: disabled, low, medium, or high")
	debug := flag.Bool("debug", false, "Show environment detection and optimization info")
	exifOverlay := flag.Bool("exif-overlay", false, "Add camera info overlay to video (bottom center)")
	overlayFontSize := flag.Int("overlay-font-size", 36, "Font size for EXIF overlay (default: 36)")
	version := flag.Bool("version", false, "Show version information")
	versionShort := flag.Bool("v", false, "Show version information (short)")
	help := flag.Bool("help", false, "Show this help message")
	gui := flag.Bool("gui", false, "Launch desktop GUI")

	// Custom usage function
	flag.Usage = func() {
		fmt.Printf("%s\n\n", utils.GetVersionInfo())
		fmt.Printf("USAGE:\n")
		fmt.Printf("  %s [OPTIONS]\n\n", "go24k")
		fmt.Printf("OPTIONS:\n")
		fmt.Printf("  -d int                                Duration per image in seconds (default 5)\n")
		fmt.Printf("  -t int                                Transition (fade) duration in seconds (default 1)\n")
		fmt.Printf("  -fps int                              Output framerate override: 30 or 60\n")
		fmt.Printf("  -effects string                       Image motion effects: disabled, low, medium, or high (default disabled)\n")
		fmt.Printf("  -fit-audio                            Auto-fit image and transition durations to fill the music length\n")
		fmt.Printf("  -include-videos                       Include supported video files (mp4, mov, mkv, avi, webm, m4v) together with pictures\n")
		fmt.Printf("  -keep-video-audio                     Keep input video audio and blend it with MP3 background audio\n")
		fmt.Printf("  -order string                         Timeline order: metadata, filename, or random (default metadata)\n")
		fmt.Printf("  -fullhd                               Generate Full HD (1920x1080) video instead of 4K UHD (3840x2160)\n")
		fmt.Printf("  -exif-overlay                         Add camera info overlay to video (bottom center)\n")
		fmt.Printf("  -overlay-font-size int                Font size for EXIF overlay (default 36)\n")
		fmt.Printf("  -gui                                  Launch desktop GUI\n")
		fmt.Printf("  -debug                                Show environment detection and optimization info\n")
		fmt.Printf("  -version                              Show version information\n")
		fmt.Printf("  -v                                    Show version information (short)\n")
		fmt.Printf("  -help                                 Show this help message\n")
		fmt.Printf("\nEXAMPLES:\n")
		fmt.Printf("  go24k                                      # Create 4K video with default settings (effects disabled)\n")
		fmt.Printf("  go24k                                      # Auto FPS: 30 when effects are disabled, 60 when enabled\n")
		fmt.Printf("  go24k -d 8 -t 2                            # 8s per image, 2s transitions\n")
		fmt.Printf("  go24k -fps 60                              # Smoother motion at 60 fps\n")
		fmt.Printf("  go24k -effects disabled                    # Disable image motion effects\n")
		fmt.Printf("  go24k -effects low                         # Pan + zoom with low intensity\n")
		fmt.Printf("  go24k -effects medium                      # Pan + zoom with medium intensity\n")
		fmt.Printf("  go24k -effects high                        # Pan + zoom with high intensity\n")
		fmt.Printf("  go24k -exif-overlay                        # Add camera info overlay\n")
		fmt.Printf("  go24k -exif-overlay -overlay-font-size 48  # Large font overlay\n")
		fmt.Printf("  go24k -fit-audio                         # Auto-fit duration to music length\n")
		fmt.Printf("  go24k -include-videos                    # Mix videos (including MOV) with pictures in the timeline\n")
		fmt.Printf("  go24k -order random                      # Random timeline order\n")
		fmt.Printf("  go24k -order filename                    # Filename timeline order\n")
		fmt.Printf("  go24k -include-videos -keep-video-audio  # Keep clip audio and blend it with MP3 audio\n")
		fmt.Printf("  go24k -fullhd                              # Generate Full HD (1920x1080) video\n")
		fmt.Printf("  go24k -gui                                 # Open desktop GUI\n")
		fmt.Printf("  go24k -debug                               # Show hardware detection info\n")
		fmt.Printf("\nFor more information: https://github.com/aloula/go24k\n")
	}

	flag.Parse()

	fpsSpecified := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "fps" {
			fpsSpecified = true
		}
	})

	// Show help if requested
	if *help {
		flag.Usage()
		return
	}

	// Show version info if requested
	if *version || *versionShort {
		if *version {
			fmt.Println(utils.GetFullVersionInfo())
		} else {
			fmt.Println(utils.GetVersionInfo())
		}
		return
	}

	if *gui || shouldAutoLaunchGUI() {
		launchGUI()
		return
	}

	// Show version on startup (brief)
	fmt.Printf("🎬 %s\n", utils.GetVersionInfo())

	// Show debug info if requested
	if *debug {
		utils.ShowEnvironmentInfo()
		return
	}

	startTime := time.Now()

	// Convert images (e.g. scale, add background, overlay, etc.)
	if err := utils.ConvertImages(*fullHD); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	resolvedEffectsMode := strings.ToLower(strings.TrimSpace(*effectsMode))
	switch resolvedEffectsMode {
	case "disabled", "low", "medium", "high":
	default:
		fmt.Printf("Error: invalid -effects value %q. Use disabled, low, medium, or high\n", *effectsMode)
		return
	}

	applyKenBurns := resolvedEffectsMode != "disabled"
	kenBurnsMode := resolvedEffectsMode
	if !applyKenBurns {
		kenBurnsMode = "high"
	}

	resolvedOrderMode := strings.ToLower(strings.TrimSpace(*orderMode))
	switch resolvedOrderMode {
	case "", "metadata":
		resolvedOrderMode = "metadata"
	case "filename", "random":
	default:
		fmt.Printf("Error: invalid -order value %q. Use metadata, filename, or random\n", *orderMode)
		return
	}

	// Backward-compatible aliases; explicit legacy flags override -order.
	if *orderByFilename {
		resolvedOrderMode = "filename"
	}
	if *randomOrder {
		resolvedOrderMode = "random"
	}

	resolvedOrderByFilename := resolvedOrderMode == "filename"
	resolvedRandomOrder := resolvedOrderMode == "random"

	targetFPS := *fps
	if !fpsSpecified {
		if applyKenBurns {
			targetFPS = 60
		} else {
			targetFPS = 30
		}
	}
	// Pass the duration and transition values from the flags.
	utils.GenerateVideo(*duration, *transition, applyKenBurns, *exifOverlay, *overlayFontSize, *fitAudio, *includeVideos, *keepVideoAudio, *fullHD, targetFPS, resolvedOrderByFilename, resolvedRandomOrder, kenBurnsMode)

	elapsedTime := time.Since(startTime).Seconds()
	fmt.Printf("Total time: %.1f sec.\n", elapsedTime)
}

func shouldAutoLaunchGUI() bool {
	if os.Getenv("GO24K_INTERNAL_CLI") == "1" {
		return false
	}

	if len(os.Args) > 1 {
		return false
	}

	exePath, err := os.Executable()
	if err != nil {
		return false
	}

	baseName := strings.ToLower(filepath.Base(exePath))
	return strings.Contains(baseName, "go24k-gui")
}
