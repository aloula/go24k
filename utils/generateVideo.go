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

// Hardware encoder detection functions
func checkNVENCAvailable() bool {
	cmd := exec.Command("ffmpeg", "-encoders")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "h264_nvenc")
}

func checkQSVAvailable() bool {
	cmd := exec.Command("ffmpeg", "-encoders")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "h264_qsv")
}

func checkAMFAvailable() bool {
	cmd := exec.Command("ffmpeg", "-encoders")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "h264_amf")
}

func checkMediaFoundationAvailable() bool {
	cmd := exec.Command("ffmpeg", "-encoders")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "h264_mf")
}

func checkVAAPIAvailable() bool {
	cmd := exec.Command("ffmpeg", "-encoders")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "h264_vaapi")
}

// HardwareEncoder represents different hardware encoding options
type HardwareEncoder struct {
	Name        string
	Codec       string
	Description string
	Platform    string
}

// getOptimalVideoSettings returns optimized FFmpeg settings based on environment and hardware
func getOptimalVideoSettings() []string {
	// Check hardware acceleration availability in priority order
	hasNVENC := checkNVENCAvailable()
	hasQSV := checkQSVAvailable()
	hasAMF := checkAMFAvailable()
	hasMediaFoundation := checkMediaFoundationAvailable()
	hasVAAPI := checkVAAPIAvailable()

	// Base settings
	settings := []string{
		"-pix_fmt", "yuv420p",
		"-movflags", "+faststart",
		"-r", "30",
		"-s", "3840x2160",
	}

	// Priority order: NVENC > Media Foundation (for Snapdragon) > QSV > AMF > VAAPI > CPU
	if hasNVENC {
		// NVIDIA GPU acceleration
		fmt.Printf("🚀 Hardware: NVIDIA NVENC detected - using GPU acceleration\n")
		settings = append(settings,
			"-c:v", "h264_nvenc",
			"-preset", "slow",
			"-profile:v", "high",
			"-level", "5.1",
			"-rc:v", "vbr",
			"-cq:v", "21",
			"-b:v", "0",
			"-maxrate", "15M",
			"-bufsize", "30M",
		)
	} else if hasMediaFoundation {
		// Windows Media Foundation (Snapdragon X, Intel QuickSync, AMD)
		// Tested on Snapdragon X Plus: ~5 seconds faster encoding (25.7s vs ~30s CPU)
		fmt.Printf("🧠 Hardware: Media Foundation detected - using Windows hardware acceleration\n")
		settings = append(settings,
			"-c:v", "h264_mf",
			"-quality", "quality", // Use quality mode
			"-rate_control", "quality", // Quality-based rate control
			"-scenario", "display_remoting", // Optimized for high-quality encoding
			"-profile:v", "high",
			"-level", "5.1",
			"-b:v", "8M", // Target bitrate
			"-maxrate", "12M",
			"-bufsize", "16M",
		)
	} else if hasQSV {
		// Intel Quick Sync Video
		fmt.Printf("⚡ Hardware: Intel QSV detected - using Intel hardware acceleration\n")
		settings = append(settings,
			"-c:v", "h264_qsv",
			"-preset", "slower", // QSV preset for quality
			"-profile:v", "high",
			"-level", "5.1",
			"-global_quality", "21", // Similar to CRF
			"-look_ahead", "1",
			"-maxrate", "12M",
			"-bufsize", "24M",
		)
	} else if hasAMF {
		// AMD Advanced Media Framework
		fmt.Printf("🔥 Hardware: AMD AMF detected - using AMD hardware acceleration\n")
		settings = append(settings,
			"-c:v", "h264_amf",
			"-quality", "quality", // Quality mode
			"-rc", "cqp", // Constant quantization parameter
			"-qp_i", "21", "-qp_p", "21", "-qp_b", "21", // Quality settings
			"-profile:v", "high",
			"-level", "5.1",
			"-maxrate", "12M",
			"-bufsize", "24M",
		)
	} else if hasVAAPI {
		// Linux VAAPI (Intel/AMD integrated graphics)
		fmt.Printf("🐧 Hardware: VAAPI detected - using Linux hardware acceleration\n")
		settings = append(settings,
			"-c:v", "h264_vaapi",
			"-profile:v", "high",
			"-level", "5.1",
			"-crf", "21", // Constant rate factor
			"-maxrate", "10M",
			"-bufsize", "20M",
		)
	} else {
		// Fallback to CPU encoding
		fmt.Printf("💻 CPU: Using libx264 software encoding\n")
		settings = append(settings,
			"-c:v", "libx264",
			"-preset", "slow",
			"-profile:v", "high",
			"-level", "5.1",
			"-crf", "21", // Constant rate factor
		)
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

	// Check all hardware acceleration types
	hasNVENC := checkNVENCAvailable()
	hasQSV := checkQSVAvailable()
	hasAMF := checkAMFAvailable()
	hasMediaFoundation := checkMediaFoundationAvailable()
	hasVAAPI := checkVAAPIAvailable()

	fmt.Printf("\nHardware Acceleration Detection:\n")

	// Show what's available
	if hasNVENC {
		fmt.Printf("  🚀 NVIDIA NVENC: Available\n")
	}
	if hasMediaFoundation {
		fmt.Printf("  🧠 Windows Media Foundation: Available (Snapdragon X, Intel, AMD)\n")
	}
	if hasQSV {
		fmt.Printf("  ⚡ Intel Quick Sync (QSV): Available\n")
	}
	if hasAMF {
		fmt.Printf("  🔥 AMD AMF: Available\n")
	}
	if hasVAAPI {
		fmt.Printf("  🐧 Linux VAAPI: Available\n")
	}

	// Show selected encoder
	fmt.Printf("\nSelected Encoder:\n")
	if hasNVENC {
		fmt.Printf("  🎯 Using: NVIDIA NVENC (highest priority)\n")
		fmt.Printf("  ⚡ Performance: ~5-10x faster than CPU\n")
	} else if hasMediaFoundation {
		fmt.Printf("  🎯 Using: Windows Media Foundation\n")
		fmt.Printf("  🧠 Optimized for: Snapdragon X Plus hardware encoding\n")
		fmt.Printf("  ⚡ Performance: ~3-5x faster than CPU\n")
	} else if hasQSV {
		fmt.Printf("  🎯 Using: Intel Quick Sync Video\n")
		fmt.Printf("  ⚡ Performance: ~2-4x faster than CPU\n")
	} else if hasAMF {
		fmt.Printf("  🎯 Using: AMD Advanced Media Framework\n")
		fmt.Printf("  ⚡ Performance: ~2-4x faster than CPU\n")
	} else if hasVAAPI {
		fmt.Printf("  🎯 Using: Linux VAAPI\n")
		fmt.Printf("  ⚡ Performance: ~2-3x faster than CPU\n")
	} else {
		fmt.Printf("  💻 Using: CPU libx264 (software encoding)\n")
		fmt.Printf("  ⏱️  Performance: Standard CPU-based encoding\n")
	}

	// Show the settings that would be used
	settings := getOptimalVideoSettings()
	fmt.Printf("\nOptimized FFmpeg Settings:\n")
	for i := 0; i < len(settings); i += 2 {
		if i+1 < len(settings) {
			fmt.Printf("  %s: %s\n", settings[i], settings[i+1])
		}
	}

	// Show quality explanation based on selected encoder
	fmt.Printf("\nEncoding Strategy:\n")
	if hasNVENC {
		fmt.Printf("  • NVIDIA NVENC: CQ 21 (constant quality)\n")
		fmt.Printf("  • Bitrate: Variable (up to 15 Mbps for 4K)\n")
		fmt.Printf("  • Speed: 5-10x faster than CPU\n")
	} else if hasMediaFoundation {
		fmt.Printf("  • Media Foundation: Quality mode optimized for Snapdragon X\n")
		fmt.Printf("  • Bitrate: 8 Mbps target (up to 12 Mbps max)\n")
		fmt.Printf("  • Speed: 3-5x faster than CPU (hardware acceleration)\n")
	} else if hasQSV {
		fmt.Printf("  • Intel QSV: Global quality 21 with look-ahead\n")
		fmt.Printf("  • Bitrate: Variable (up to 12 Mbps for 4K)\n")
		fmt.Printf("  • Speed: 2-4x faster than CPU\n")
	} else if hasAMF {
		fmt.Printf("  • AMD AMF: Constant QP mode (21 for all frame types)\n")
		fmt.Printf("  • Bitrate: Variable (up to 12 Mbps for 4K)\n")
		fmt.Printf("  • Speed: 2-4x faster than CPU\n")
	} else if hasVAAPI {
		fmt.Printf("  • Linux VAAPI: CRF 21 with hardware acceleration\n")
		fmt.Printf("  • Bitrate: Variable (up to 10 Mbps for 4K)\n")
		fmt.Printf("  • Speed: 2-3x faster than CPU\n")
	} else {
		fmt.Printf("  • CPU libx264: CRF 21 (software encoding)\n")
		fmt.Printf("  • Quality: High (software optimized)\n")
		fmt.Printf("  • Speed: Standard CPU performance\n")
	}

	fmt.Printf("\nQuality Reference:\n")
	fmt.Printf("  • Value 18-20: Visually lossless quality\n")
	fmt.Printf("  • Value 21-23: High quality (recommended)\n")
	fmt.Printf("  • Value 24-28: Medium quality\n")
	fmt.Printf("  • Hardware encoders use equivalent quality settings\n")
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
