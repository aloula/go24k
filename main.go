package main

import (
	"flag"
	"fmt"
	"time"

	"go24k/utils"
)

// showWelcomeMessage displays the Go24K ASCII art logo and welcome message
func showWelcomeMessage() {
	fmt.Print(`
╔══════════════════════════════════════════════════════════════════╗
║                                                                  ║
║     ██████   ██████  ██████  ██   ██ ██   ██                    ║
║    ██       ██    ██      ██ ██   ██ ██  ██                     ║
║    ██   ███ ██    ██  █████  ███████ █████                      ║
║    ██    ██ ██    ██ ██      ██   ██ ██  ██                     ║
║     ██████   ██████  ███████ ██   ██ ██   ██                    ║
║                                                                  ║
║    🎬 Professional 4K Video Creator from Images 🎥              ║
║    📸 Ken Burns Effects • Hardware Acceleration • 4K Quality     ║
║                                                                  ║
╚══════════════════════════════════════════════════════════════════╝
`)
	fmt.Println("🚀 Starting Go24K Video Generation...")
	fmt.Println()
}

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

	// Show welcome message with ASCII art logo
	showWelcomeMessage()

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
