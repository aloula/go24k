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
	flag.Parse()

	startTime := time.Now()

	// Convert images (e.g. scale, add background, overlay, etc.)
	utils.ConvertImages()

	// Generate video only if convert-only is not set.
	if !*convertOnly {
		// If -static is provided, applyKenBurns will be false.
		applyKenBurns := !*static
		// Pass the duration and transition values from the flags.
		utils.GenerateVideo(*duration, *transition, applyKenBurns)
	}

	// Report processing time.
	elapsedTime := time.Since(startTime).Seconds()
	fmt.Printf("\nProcessing time: %.1f seconds\n", elapsedTime)
	fmt.Println("Done!")
}
