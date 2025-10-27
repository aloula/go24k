package main

import (
	"flag"
	"fmt"
	"time"

	"go_video_tools/utils"
)

func main() {
	// Set up command-line flags.
	convertOnly := flag.Bool("convert-only", false, "Convert images only, without generating the video")
	static := flag.Bool("static", false, "Do NOT apply Ken Burns effect; use static images with transitions")
	duration := flag.Int("d", 5, "Duration per image in seconds")
	transition := flag.Int("t", 1, "Transition (fade) duration in seconds")

	// GIF-related flags
	createGif := flag.Bool("gif", false, "Create animated GIF instead of video")
	optimizedGif := flag.Bool("gif-optimized", false, "Create optimized animated GIF with palette (smaller file size)")
	gifFps := flag.Int("gif-fps", 15, "Frames per second for GIF (higher = smoother animation)")
	gifScale := flag.Float64("gif-scale", 1.0, "Scale factor for GIF output (1.0 = full converted size)")
	gifTotalTime := flag.Int("gif-total-time", 0, "Total duration of GIF in seconds (overrides per-image duration)")

	flag.Parse()

	startTime := time.Now()

	// Convert images (e.g. scale, add background, overlay, etc.)
	utils.ConvertImages()

	// Generate output only if convert-only is not set.
	if !*convertOnly {
		// Calculate duration per image based on total time or use individual duration
		gifDuration := *duration

		if (*createGif || *optimizedGif) && *gifTotalTime == 0 && *duration == 5 {
			// Use faster default for GIFs (1 second per image instead of 5)
			gifDuration = 1
			fmt.Println("Using optimized duration for GIF: 1 second per image")
		}

		if *optimizedGif {
			// Create optimized GIF with palette
			if *gifTotalTime > 0 {
				utils.GenerateOptimizedGifWithTotalTime(*gifTotalTime, *transition, *gifFps, *gifScale)
			} else {
				utils.GenerateOptimizedGif(gifDuration, *transition, *gifFps, *gifScale)
			}
		} else if *createGif {
			// Create regular animated GIF
			if *gifTotalTime > 0 {
				utils.GenerateGifWithTotalTime(*gifTotalTime, *transition, *gifFps, *gifScale)
			} else {
				utils.GenerateGif(gifDuration, *transition, *gifFps, *gifScale)
			}
		} else {
			// Generate video (default behavior)
			// If -static is provided, applyKenBurns will be false.
			applyKenBurns := !*static
			// Pass the duration and transition values from the flags.
			utils.GenerateVideo(*duration, *transition, applyKenBurns)
		}
	}

	// Report processing time.
	elapsedTime := time.Since(startTime).Seconds()
	fmt.Printf("\nProcessing time: %.1f seconds\n", elapsedTime)
	fmt.Println("Done!")
}
