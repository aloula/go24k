package utils

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// isWSL detects if we're running in Windows Subsystem for Linux
func isWSL() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	// Check /proc/version for WSL signature
	if data, err := os.ReadFile("/proc/version"); err == nil {
		version := strings.ToLower(string(data))
		return strings.Contains(version, "microsoft") || strings.Contains(version, "wsl")
	}

	// Fallback: check for WSL environment variable
	if wslDistro := os.Getenv("WSL_DISTRO_NAME"); wslDistro != "" {
		return true
	}

	return false
}

// getOptimalVideoSettings returns optimized FFmpeg settings based on environment
func getOptimalVideoSettings() []string {
	// Base settings optimized for quality and compatibility
	settings := []string{
		"-c:v", "libx264",
		"-preset", "slow", // Better quality than "fast"
		"-profile:v", "high", // Higher profile for better compression
		"-level", "4.0", // Higher level for 4K support
		"-pix_fmt", "yuv420p",
		"-movflags", "+faststart",
		"-r", "30",
		"-s", "3840x2160",
	}

	// Detect environment and apply consistent quality settings
	isWindowsNative := runtime.GOOS == "windows"
	isWSLEnv := isWSL()

	if isWindowsNative {
		// Windows native: use slightly higher CRF for consistency
		fmt.Printf("Environment: Windows native - using optimized settings\n")
		settings = append(settings, "-crf", "21") // Slightly better quality
	} else if isWSLEnv {
		// WSL: match Windows quality explicitly
		fmt.Printf("Environment: WSL detected - using Windows-compatible settings\n")
		settings = append(settings, "-crf", "21") // Match Windows quality
	} else {
		// Linux native: standard high quality
		fmt.Printf("Environment: Linux native - using high quality settings\n")
		settings = append(settings, "-crf", "20") // Highest quality for Linux
	}

	return settings
}

// ShowEnvironmentInfo displays environment detection and optimization details
func ShowEnvironmentInfo() {
	fmt.Printf("=== Go24K Environment Detection ===\n\n")

	fmt.Printf("Operating System: %s\n", runtime.GOOS)
	fmt.Printf("Architecture: %s\n", runtime.GOARCH)

	if runtime.GOOS == "linux" {
		if isWSL() {
			fmt.Printf("Environment: WSL (Windows Subsystem for Linux)\n")
		} else {
			fmt.Printf("Environment: Native Linux\n")
		}
	} else {
		fmt.Printf("Environment: Native %s\n", strings.ToUpper(runtime.GOOS[:1])+runtime.GOOS[1:])
	}

	// Show the settings that would be used
	settings := getOptimalVideoSettings()
	fmt.Printf("\nOptimized FFmpeg Settings:\n")
	for i := 0; i < len(settings); i += 2 {
		if i+1 < len(settings) {
			fmt.Printf("  %s: %s\n", settings[i], settings[i+1])
		}
	}

	// Show quality explanation
	fmt.Printf("\nQuality Optimization:\n")
	if runtime.GOOS == "windows" {
		fmt.Printf("  • Windows native: CRF 21 (high quality)\n")
	} else if isWSL() {
		fmt.Printf("  • WSL environment: CRF 21 (Windows-compatible quality)\n")
	} else {
		fmt.Printf("  • Linux native: CRF 20 (highest quality)\n")
	}

	fmt.Printf("\nCRF Scale: Lower values = higher quality/larger files\n")
	fmt.Printf("  • CRF 18-20: Visually lossless\n")
	fmt.Printf("  • CRF 21-23: High quality\n")
	fmt.Printf("  • CRF 24-28: Medium quality\n")
}

// GenerateVideo creates a video from already 3840x2160 images with crossfade transitions,
// audio fades, and optionally a Ken Burns effect applied to each image.
// If applyKenBurns is false, the images remain static.
func GenerateVideo(duration, fadeDuration int, applyKenBurns bool) {
	// Find all converted .jpg files (3840x2160).
	files, err := filepath.Glob("converted/*.jpg")
	if err != nil {
		log.Fatalf("Failed to list converted .jpg files: %v", err)
	}

	index := 0
	inputs := []string{}
	filterComplex := ""

	// Process each image file.
	for _, file := range files {
		inputs = append(inputs, "-loop", "1", "-t", fmt.Sprintf("%d", duration), "-i", file)
		if applyKenBurns {
			// Apply Ken Burns effect.
			effect := getKenBurnsEffect(duration)
			if index == 0 {
				// For the first image, apply the effect followed by a fade-in.
				filterComplex += fmt.Sprintf("[0:v]%s,fade=t=in:st=0:d=%d[v%d]; ", effect, fadeDuration, index)
			} else {
				filterComplex += fmt.Sprintf("[%d:v]%s[v%d]; ", index, effect, index)
			}
		} else {
			// Static: no zoom/pan effect.
			if index == 0 {
				filterComplex += fmt.Sprintf("[0:v]fade=t=in:st=0:d=%d[v%d]; ", fadeDuration, index)
			} else {
				filterComplex += fmt.Sprintf("[%d:v]copy[v%d]; ", index, index)
			}
		}
		index++
	}

	totalFiles := len(files)

	// Generate crossfade transitions.
	for i := 0; i < index-1; i++ {
		next := i + 1
		offset := (i + 1) * (duration - fadeDuration)
		if i == 0 {
			filterComplex += fmt.Sprintf("[v%d][v%d]xfade=transition=fade:duration=%d:offset=%d[x%d]; ", i, next, fadeDuration, offset, next)
		} else {
			filterComplex += fmt.Sprintf("[x%d][v%d]xfade=transition=fade:duration=%d:offset=%d[x%d]; ", i, next, fadeDuration, offset, next)
		}
	}

	// Apply fade-out to the final image.
	totalDuration := index*duration - (index-1)*fadeDuration
	startFadeOut := totalDuration - fadeDuration
	filterComplex += fmt.Sprintf("[x%d]fade=t=out:st=%d:d=%d[xf]; ", index-1, startFadeOut, fadeDuration)

	// Force the final video to exactly be ND seconds.
	finalLength := (totalFiles * duration) - ((totalFiles - 1) * fadeDuration)
	filterComplex += fmt.Sprintf("[xf]trim=duration=%d,setpts=PTS-STARTPTS[xfout]; ", finalLength)

	// Check for music input.
	musicFiles, err := filepath.Glob("*.mp3")
	if err != nil {
		log.Fatalf("Failed to list mp3 files: %v", err)
	}

	var mapArgs []string
	hasAudio := len(musicFiles) > 0

	if hasAudio {
		fmt.Printf("Audio file found: %s\n", musicFiles[0])
		inputs = append(inputs, "-i", musicFiles[0])

		// Apply audio fades.
		filterComplex += fmt.Sprintf("[%d:a]afade=t=in:st=0:d=2,afade=t=out:st=%d:d=4[musicout]; ", index, startFadeOut-4)

		// Map video and audio
		mapArgs = []string{"-map", "[xfout]", "-map", "[musicout]", "-shortest", "video.mp4"}
	} else {
		fmt.Printf("No MP3 file found - generating video without audio\n")

		// Map only video
		mapArgs = []string{"-map", "[xfout]", "video.mp4"}
	}

	// Build the complete ffmpeg command.
	args := []string{"-y"}
	args = append(args, inputs...)
	args = append(args, "-filter_complex", filterComplex)
	args = append(args, mapArgs...)

	// Video encoding settings with environment-specific optimization
	args = append(args, getOptimalVideoSettings()...,
	)

	// Audio encoding settings (only if audio is present)
	if hasAudio {
		args = append(args,
			"-c:a", "aac",
			"-b:a", "192k",
		)
	}

	args = append(args, "-t", fmt.Sprintf("%d", finalLength))

	// Remove printing of the FFmpeg command.
	cmd := exec.Command("ffmpeg", args...)

	// Redirect FFmpeg logs to /dev/null.
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		log.Fatalf("Failed to open /dev/null: %v", err)
	}
	cmd.Stdout = devNull
	cmd.Stderr = devNull

	if err := cmd.Start(); err != nil {
		log.Fatalf("ffmpeg start failed: %v", err)
	}

	done := make(chan struct{})
	go func() {
		spinnerChars := []string{"|", "/", "-", "\\"}
		i := 0
		var message string
		if hasAudio {
			message = "Generating video with audio"
		} else {
			message = "Generating video (no audio)"
		}

		for {
			select {
			case <-done:
				fmt.Print("\r")
				return
			default:
				fmt.Printf("\r%s...:   %s", message, spinnerChars[i%len(spinnerChars)])
				i++
				time.Sleep(200 * time.Millisecond)
			}
		}
	}()

	if err := cmd.Wait(); err != nil {
		close(done)
		log.Fatalf("ffmpeg command failed: %v", err)
	}
	close(done)

	// Display success message with video information
	fmt.Printf("\n=== Video generated successfully! ===\n")
	fmt.Printf("File: video.mp4\n")
	fmt.Printf("Resolution: 4K UHD (3840x2160)\n")
	fmt.Printf("Duration: %d seconds\n", finalLength)
	fmt.Printf("Images: %d\n", totalFiles)
	if hasAudio {
		fmt.Printf("Audio: %s\n", filepath.Base(musicFiles[0]))
	} else {
		fmt.Printf("Audio: None (no MP3 file found)\n")
	}

	// Get file size for user feedback
	if fileInfo, err := os.Stat("video.mp4"); err == nil {
		sizeKB := fileInfo.Size() / 1024
		sizeMB := float64(sizeKB) / 1024
		if sizeMB < 1024 {
			fmt.Printf("Size: %.1f MB\n", sizeMB)
		} else {
			fmt.Printf("Size: %.2f GB\n", sizeMB/1024)
		}
	}
}

// getKenBurnsEffect generates a Ken Burns effect using a fixed zoompan expression.
// This approach is based on the method described in the Bannerbear blog.
func getKenBurnsEffect(duration int) string {
	totalFrames := duration * 30
	offset := totalFrames * 2 // adjust offset as desired

	// Define nine variants based on different focal positions.
	centerExpr := "zoompan=zoom='min(zoom+0.001,1.5)':x='iw/2-(iw/zoom/2)':y='ih/2-(ih/zoom/2)':d=%d:s=3840x2160"
	topLeftExpr := "zoompan=zoom='min(zoom+0.001,1.5)':x='iw/2-(iw/zoom/2)-%d':y='ih/2-(ih/zoom/2)-%d':d=%d:s=3840x2160"
	topRightExpr := "zoompan=zoom='min(zoom+0.001,1.5)':x='iw/2-(iw/zoom/2)+%d':y='ih/2-(ih/zoom/2)-%d':d=%d:s=3840x2160"
	bottomLeftExpr := "zoompan=zoom='min(zoom+0.001,1.5)':x='iw/2-(iw/zoom/2)-%d':y='ih/2-(ih/zoom/2)+%d':d=%d:s=3840x2160"
	bottomRightExpr := "zoompan=zoom='min(zoom+0.001,1.5)':x='iw/2-(iw/zoom/2)+%d':y='ih/2-(ih/zoom/2)+%d':d=%d:s=3840x2160"
	leftExpr := "zoompan=zoom='min(zoom+0.001,1.5)':x='iw/2-(iw/zoom/2)-%d':y='ih/2-(ih/zoom/2)':d=%d:s=3840x2160"
	rightExpr := "zoompan=zoom='min(zoom+0.001,1.5)':x='iw/2-(iw/zoom/2)+%d':y='ih/2-(ih/zoom/2)':d=%d:s=3840x2160"
	topExpr := "zoompan=zoom='min(zoom+0.001,1.5)':x='iw/2-(iw/zoom/2)':y='ih/2-(ih/zoom/2)-%d':d=%d:s=3840x2160"
	bottomExpr := "zoompan=zoom='min(zoom+0.001,1.5)':x='iw/2-(iw/zoom/2)':y='ih/2-(ih/zoom/2)+%d':d=%d:s=3840x2160"

	// Create a slice with formatted expressions.
	var variants []string
	variants = append(variants, fmt.Sprintf(centerExpr, totalFrames))
	variants = append(variants, fmt.Sprintf(topLeftExpr, offset, offset, totalFrames))
	variants = append(variants, fmt.Sprintf(topRightExpr, offset, offset, totalFrames))
	variants = append(variants, fmt.Sprintf(bottomLeftExpr, offset, offset, totalFrames))
	variants = append(variants, fmt.Sprintf(bottomRightExpr, offset, offset, totalFrames))
	variants = append(variants, fmt.Sprintf(leftExpr, offset, totalFrames))
	variants = append(variants, fmt.Sprintf(rightExpr, offset, totalFrames))
	variants = append(variants, fmt.Sprintf(topExpr, offset, totalFrames))
	variants = append(variants, fmt.Sprintf(bottomExpr, offset, totalFrames))

	// Randomly choose one variant.
	expr := variants[rand.Intn(len(variants))]

	//fmt.Println("Ken Burns effect:", expr)
	return expr
}
