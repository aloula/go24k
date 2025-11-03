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
	flag.Parse()

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
		utils.GenerateVideo(*duration, *transition, applyKenBurns)
	}

	// Report processing time (only if not convert-only since conversion already shows its time)
	if !*convertOnly {
		elapsedTime := time.Since(startTime).Seconds()
		fmt.Printf("\nProcessing time: %.1f seconds\n", elapsedTime)
	}
	fmt.Println("Done!")
}
