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
	fps := flag.Int("fps", 30, "Output framerate override: 30 or 60")
	fitAudio := flag.Bool("fit-audio", false, "Auto-fit image and transition durations to fill the music length")
	includeVideos := flag.Bool("include-videos", false, "Include supported video files together with pictures")
	keepVideoAudio := flag.Bool("keep-video-audio", false, "Keep input video audio and blend it with MP3 background audio")
	orderByFilename := flag.Bool("order-by-filename", false, "Order timeline by filename instead of metadata time")
	fullHD := flag.Bool("fullhd", false, "Generate Full HD (1920x1080) video instead of 4K UHD (3840x2160)")
	kenBurnsMode := flag.String("kenburns-mode", "dynamic", "Ken Burns mode: cinematic or dynamic")
	debug := flag.Bool("debug", false, "Show environment detection and optimization info")
	exifOverlay := flag.Bool("exif-overlay", false, "Add camera info overlay to video (bottom center)")
	overlayFontSize := flag.Int("overlay-font-size", 36, "Font size for EXIF overlay (default: 36)")
	version := flag.Bool("version", false, "Show version information")
	versionShort := flag.Bool("v", false, "Show version information (short)")
	help := flag.Bool("help", false, "Show this help message")

	// Custom usage function
	flag.Usage = func() {
		fmt.Printf("%s\n\n", utils.GetVersionInfo())
		fmt.Printf("USAGE:\n")
		fmt.Printf("  %s [OPTIONS]\n\n", "go24k")
		fmt.Printf("OPTIONS:\n")
		flag.PrintDefaults()
		fmt.Printf("\nEXAMPLES:\n")
		fmt.Printf("  go24k                                      # Create 4K video with default settings\n")
		fmt.Printf("  go24k                                      # Auto FPS: 60 with Ken Burns, 30 with -static\n")
		fmt.Printf("  go24k -d 8 -t 2                            # 8s per image, 2s transitions\n")
		fmt.Printf("  go24k -fps 60                              # Smoother motion at 60 fps\n")
		fmt.Printf("  go24k -static                              # Disable Ken Burns effect\n")
		fmt.Printf("  go24k -kenburns-mode dynamic               # Directional Ken Burns motion\n")
		fmt.Printf("  go24k -exif-overlay                        # Add camera info overlay\n")
		fmt.Printf("  go24k -exif-overlay -overlay-font-size 48  # Large font overlay\n")
		fmt.Printf("  go24k -fit-audio                         # Auto-fit duration to music length\n")
		fmt.Printf("  go24k -include-videos                    # Mix videos with pictures in the timeline\n")
		fmt.Printf("  go24k -include-videos -keep-video-audio  # Keep clip audio and blend it with MP3 audio\n")
		fmt.Printf("  go24k -order-by-filename                 # Ignore metadata and sort timeline by filename\n")
		fmt.Printf("  go24k -fullhd                              # Generate Full HD (1920x1080) video\n")
		fmt.Printf("  go24k -convert-only                        # Only convert images to 4K\n")
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

	// Generate video only if convert-only is not set.
	if !*convertOnly {
		// If -static is provided, applyKenBurns will be false.
		applyKenBurns := !*static
		targetFPS := *fps
		if !fpsSpecified {
			if applyKenBurns {
				targetFPS = 60
			} else {
				targetFPS = 30
			}
		}
		// Pass the duration and transition values from the flags.
		utils.GenerateVideo(*duration, *transition, applyKenBurns, *exifOverlay, *overlayFontSize, *fitAudio, *includeVideos, *keepVideoAudio, *fullHD, targetFPS, *orderByFilename, *kenBurnsMode)
	}

	// Report processing time (only if not convert-only since conversion already shows its time)
	if !*convertOnly {
		elapsedTime := time.Since(startTime).Seconds()
		fmt.Printf("Total time: %.1f sec.\n", elapsedTime)
	}
}
