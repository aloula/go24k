package utils

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	linuxOS      = "linux"
	resolution4K = "3840x2160"
)

// VideoInfo contains technical details about a video file
type VideoInfo struct {
	FileSizeMB   float64
	DurationSec  float64
	VideoBitrate string
	AudioBitrate string
	Framerate    string
	Resolution   string
}

// getFileSize gets the file size in MB
func getFileSize(filename string) float64 {
	if fileInfo, err := os.Stat(filename); err == nil {
		return float64(fileInfo.Size()) / (1024 * 1024)
	}
	return 0
}

// runFFProbe executes ffprobe and returns the JSON output
func runFFProbe(filename string) (string, error) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filename)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("ffprobe failed: %v", err)
	}
	return string(output), nil
}

// getVideoDetails extracts technical information from the generated video file
func getVideoDetails(filename string) (*VideoInfo, error) {
	info := &VideoInfo{}
	info.FileSizeMB = getFileSize(filename)

	outputStr, err := runFFProbe(filename)
	if err != nil {
		// Set defaults if ffprobe fails
		info.Framerate = "30 fps"
		info.Resolution = resolution4K
		info.AudioBitrate = "No audio"
		return info, err
	}

	info.DurationSec = extractDuration(outputStr)
	info.VideoBitrate, info.Framerate, info.Resolution = extractVideoInfo(outputStr)
	info.AudioBitrate = extractAudioInfo(outputStr)

	// Set defaults if not found
	if info.Framerate == "" {
		info.Framerate = "30 fps"
	}
	if info.Resolution == "" {
		info.Resolution = resolution4K
	}
	if info.AudioBitrate == "" {
		info.AudioBitrate = "No audio"
	}

	return info, nil
}

// extractDuration parses duration from ffprobe JSON output
func extractDuration(outputStr string) float64 {
	if !strings.Contains(outputStr, `"duration"`) {
		return 0
	}

	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, `"duration"`) && strings.Contains(line, `"format"`) {
			parts := strings.Split(line, `"`)
			for i, part := range parts {
				if part == "duration" && i+2 < len(parts) {
					if duration, err := strconv.ParseFloat(parts[i+2], 64); err == nil {
						return duration
					}
				}
			}
		}
	}
	return 0
}

// extractVideoInfo parses video stream information from ffprobe output
func extractVideoInfo(outputStr string) (bitrate, framerate, resolution string) {
	lines := strings.Split(outputStr, "\n")
	var inVideoStream bool

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, `"codec_type": "video"`) {
			inVideoStream = true
			continue
		}
		if strings.Contains(line, `"codec_type": "audio"`) {
			inVideoStream = false
		}

		if inVideoStream {
			if strings.Contains(line, `"bit_rate"`) && bitrate == "" {
				parts := strings.Split(line, `"`)
				for i, part := range parts {
					if part == "bit_rate" && i+2 < len(parts) {
						if br, err := strconv.Atoi(parts[i+2]); err == nil {
							bitrate = fmt.Sprintf("%.1f Mbps", float64(br)/1000000)
						}
						break
					}
				}
			}
			if strings.Contains(line, `"r_frame_rate"`) && framerate == "" {
				parts := strings.Split(line, `"`)
				for i, part := range parts {
					if part == "r_frame_rate" && i+2 < len(parts) {
						frameRate := parts[i+2]
						if strings.Contains(frameRate, "/") {
							rateParts := strings.Split(frameRate, "/")
							if len(rateParts) == 2 {
								if num, err1 := strconv.ParseFloat(rateParts[0], 64); err1 == nil {
									if den, err2 := strconv.ParseFloat(rateParts[1], 64); err2 == nil && den != 0 {
										framerate = fmt.Sprintf("%.0f fps", num/den)
									}
								}
							}
						}
						break
					}
				}
			}
			if strings.Contains(line, `"width"`) && strings.Contains(line, `"height"`) && resolution == "" {
				resolution = resolution4K // We know our output resolution
			}
		}
	}
	return bitrate, framerate, resolution
}

// extractAudioInfo parses audio bitrate from ffprobe output
func extractAudioInfo(outputStr string) string {
	if !strings.Contains(outputStr, `"codec_type": "audio"`) {
		return ""
	}

	lines := strings.Split(outputStr, "\n")
	var inAudioStream bool

	for _, line := range lines {
		if strings.Contains(line, `"codec_type": "audio"`) {
			inAudioStream = true
			continue
		}
		if strings.Contains(line, `"codec_type": "video"`) {
			inAudioStream = false
		}

		if inAudioStream && strings.Contains(line, `"bit_rate"`) {
			parts := strings.Split(line, `"`)
			for i, part := range parts {
				if part == "bit_rate" && i+2 < len(parts) {
					if bitrate, err := strconv.Atoi(parts[i+2]); err == nil {
						return fmt.Sprintf("%d kbps", bitrate/1000)
					}
				}
			}
			break
		}
	}
	return ""
}

// isWSL detects if we're running in Windows Subsystem for Linux
func isWSL() bool {
	if runtime.GOOS != linuxOS {
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
	// First check if encoder is listed
	cmd := exec.Command("ffmpeg", "-encoders")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	if !strings.Contains(string(output), "h264_nvenc") {
		return false
	}

	// Test if NVENC actually works (avoid false positives in WSL/ARM systems)
	// Some systems report NVENC support but can't actually use it
	testCmd := exec.Command("ffmpeg", "-f", "lavfi", "-i", "testsrc=duration=0.1:size=320x240:rate=1",
		"-c:v", "h264_nvenc", "-f", "null", "-")
	err = testCmd.Run()
	return err == nil
}

func checkQSVAvailable() bool {
	// First check if encoder is listed
	cmd := exec.Command("ffmpeg", "-encoders")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	if !strings.Contains(string(output), "h264_qsv") {
		return false
	}

	// Test if QSV actually works
	testCmd := exec.Command("ffmpeg", "-f", "lavfi", "-i", "testsrc=duration=0.1:size=320x240:rate=1",
		"-c:v", "h264_qsv", "-f", "null", "-")
	err = testCmd.Run()
	return err == nil
}

func checkAMFAvailable() bool {
	// First check if encoder is listed
	cmd := exec.Command("ffmpeg", "-encoders")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	if !strings.Contains(string(output), "h264_amf") {
		return false
	}

	// Test if AMF actually works
	testCmd := exec.Command("ffmpeg", "-f", "lavfi", "-i", "testsrc=duration=0.1:size=320x240:rate=1",
		"-c:v", "h264_amf", "-f", "null", "-")
	err = testCmd.Run()
	return err == nil
}

func checkMediaFoundationAvailable() bool {
	// First check if encoder is listed
	cmd := exec.Command("ffmpeg", "-encoders")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	if !strings.Contains(string(output), "h264_mf") {
		return false
	}

	// Test if Media Foundation actually works
	testCmd := exec.Command("ffmpeg", "-f", "lavfi", "-i", "testsrc=duration=0.1:size=320x240:rate=1",
		"-c:v", "h264_mf", "-f", "null", "-")
	err = testCmd.Run()
	return err == nil
}

func checkVAAPIAvailable() bool {
	// First check if encoder is listed
	cmd := exec.Command("ffmpeg", "-encoders")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	if !strings.Contains(string(output), "h264_vaapi") {
		return false
	}

	// Test if VAAPI actually works
	testCmd := exec.Command("ffmpeg", "-f", "lavfi", "-i", "testsrc=duration=0.1:size=320x240:rate=1",
		"-c:v", "h264_vaapi", "-f", "null", "-")
	err = testCmd.Run()
	return err == nil
}

func checkVideoToolboxAvailable() bool {
	cmd := exec.Command("ffmpeg", "-encoders")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "h264_videotoolbox")
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
	hasVideoToolbox := checkVideoToolboxAvailable()
	hasQSV := checkQSVAvailable()
	hasAMF := checkAMFAvailable()
	hasMediaFoundation := checkMediaFoundationAvailable()
	hasVAAPI := checkVAAPIAvailable()

	// Base settings
	settings := []string{
		"-pix_fmt", "yuv420p",
		"-movflags", "+faststart",
		"-r", "30",
		"-s", resolution4K,
	}

	// Priority order: NVENC > VideoToolbox (macOS) > Media Foundation (Windows) > QSV > AMF > VAAPI > CPU
	if hasNVENC {
		// NVIDIA GPU acceleration
		fmt.Printf("Hardware: NVIDIA NVENC detected - using GPU acceleration\n")
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
	} else if hasVideoToolbox {
		// Apple VideoToolbox (macOS native hardware acceleration)
		fmt.Printf("Hardware: VideoToolbox detected - using Apple hardware acceleration\n")
		settings = append(settings,
			"-c:v", "h264_videotoolbox",
			"-profile:v", "high",
			"-level", "5.1",
			"-q:v", "21", // Quality-based encoding similar to CRF
			"-realtime", "false", // Better quality encoding
			"-frames:v", "0", // Unlimited frames
			"-b:v", "10M", // Target bitrate for 4K
			"-maxrate", "15M",
			"-bufsize", "30M",
		)
	} else if hasMediaFoundation {
		// Windows Media Foundation (Snapdragon X, Intel QuickSync, AMD)
		// Tested on Snapdragon X Plus: ~5 seconds faster encoding (25.7s vs ~30s CPU)
		// Optimized bitrate settings to match NVENC performance (15 Mbps target)
		fmt.Printf("Hardware: Media Foundation detected - using Windows hardware acceleration\n")
		settings = append(settings,
			"-c:v", "h264_mf",
			"-quality", "quality", // Use quality mode
			"-rate_control", "quality", // Quality-based rate control
			"-scenario", "display_remoting", // Optimized for high-quality encoding
			"-profile:v", "high",
			"-level", "5.1",
			"-b:v", "12M", // Increased target bitrate (was 8M)
			"-maxrate", "18M", // Increased max bitrate to exceed NVENC (was 12M)
			"-bufsize", "36M", // Doubled buffer size for smoother encoding (was 16M)
		)
	} else if hasQSV {
		// Intel Quick Sync Video
		fmt.Printf("Hardware: Intel QSV detected - using Intel hardware acceleration\n")
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
		fmt.Printf("Hardware: AMD AMF detected - using AMD hardware acceleration\n")
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
		fmt.Printf("Hardware: VAAPI detected - using Linux hardware acceleration\n")
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
		fmt.Printf("CPU: Using libx264 software encoding\n")
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

	if runtime.GOOS == linuxOS {
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
	hasVideoToolbox := checkVideoToolboxAvailable()
	hasQSV := checkQSVAvailable()
	hasAMF := checkAMFAvailable()
	hasMediaFoundation := checkMediaFoundationAvailable()
	hasVAAPI := checkVAAPIAvailable()

	fmt.Printf("\nHardware Acceleration Detection:\n")

	// Show what's available
	if hasNVENC {
		fmt.Printf("  NVIDIA NVENC: Available\n")
	}
	if hasVideoToolbox {
		fmt.Printf("  Apple VideoToolbox: Available\n")
	}
	if hasMediaFoundation {
		fmt.Printf("  Windows Media Foundation: Available (Snapdragon X, Intel, AMD)\n")
	}
	if hasQSV {
		fmt.Printf("  Intel Quick Sync (QSV): Available\n")
	}
	if hasAMF {
		fmt.Printf("  AMD AMF: Available\n")
	}
	if hasVAAPI {
		fmt.Printf("  Linux VAAPI: Available\n")
	}

	// Show selected encoder
	fmt.Printf("\nSelected Encoder:\n")
	if hasNVENC {
		fmt.Printf("  Using: NVIDIA NVENC (highest priority)\n")
		fmt.Printf("  Performance: ~5-10x faster than CPU\n")
	} else if hasVideoToolbox {
		fmt.Printf("  Using: Apple VideoToolbox\n")
		fmt.Printf("  Optimized for: Apple Silicon (M1/M2/M3) hardware encoding\n")
		fmt.Printf("  Performance: ~3-8x faster than CPU\n")
	} else if hasMediaFoundation {
		fmt.Printf("  Using: Windows Media Foundation\n")
		fmt.Printf("  Optimized for: Snapdragon X Plus hardware encoding\n")
		fmt.Printf("  Performance: ~3-5x faster than CPU\n")
	} else if hasQSV {
		fmt.Printf("  Using: Intel Quick Sync Video\n")
		fmt.Printf("  Performance: ~2-4x faster than CPU\n")
	} else if hasAMF {
		fmt.Printf("  Using: AMD Advanced Media Framework\n")
		fmt.Printf("  Performance: ~2-4x faster than CPU\n")
	} else if hasVAAPI {
		fmt.Printf("  Using: Linux VAAPI\n")
		fmt.Printf("  Performance: ~2-3x faster than CPU\n")
	} else {
		fmt.Printf("  Using: CPU libx264 (software encoding)\n")
		fmt.Printf("  Performance: Standard CPU-based encoding\n")
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

// formatSeconds ensures FFmpeg receives consistent decimal timing values.
func formatSeconds(seconds float64) string {
	if seconds < 0 {
		seconds = 0
	}
	return strconv.FormatFloat(seconds, 'f', 3, 64)
}

// processImageFilter creates the video filter for a single image
func processImageFilter(file string, index int, duration, fadeDuration float64, applyKenBurns, exifOverlay bool, fontSize int) string {
	var videoFilter string

	if applyKenBurns {
		// Apply Ken Burns effect.
		effect := getKenBurnsEffect(duration)
		if index == 0 {
			// For the first image, apply the effect followed by a fade-in.
			videoFilter = fmt.Sprintf("[0:v]%s,fade=t=in:st=0:d=%s", effect, formatSeconds(fadeDuration))
		} else {
			videoFilter = fmt.Sprintf("[%d:v]%s", index, effect)
		}
	} else {
		// Static: no zoom/pan effect.
		if index == 0 {
			videoFilter = fmt.Sprintf("[0:v]fade=t=in:st=0:d=%s", formatSeconds(fadeDuration))
		} else {
			videoFilter = fmt.Sprintf("[%d:v]copy", index)
		}
	}

	// Add EXIF overlay if requested
	if exifOverlay {
		originalFile := GetOriginalFilename(file)
		if originalFile != "" {
			if cameraInfo, err := ExtractCameraInfo(originalFile); err == nil && cameraInfo != nil {
				drawtextFilter := FormatCameraInfoOverlay(cameraInfo, fontSize, index)
				if drawtextFilter != "" {
					videoFilter += drawtextFilter
				}
			}
		}
	}

	return videoFilter
}

// buildCrossfadeFilters creates the crossfade transition filters
func buildCrossfadeFilters(numImages int, duration, fadeDuration float64) string {
	var filterComplex string

	// Generate crossfade transitions.
	for i := 0; i < numImages-1; i++ {
		next := i + 1
		offset := float64(i+1) * (duration - fadeDuration)
		if i == 0 {
			filterComplex += fmt.Sprintf("[v%d][v%d]xfade=transition=fade:duration=%s:offset=%s[x%d]; ", i, next, formatSeconds(fadeDuration), formatSeconds(offset), next)
		} else {
			filterComplex += fmt.Sprintf("[x%d][v%d]xfade=transition=fade:duration=%s:offset=%s[x%d]; ", i, next, formatSeconds(fadeDuration), formatSeconds(offset), next)
		}
	}

	return filterComplex
}

// buildFinalFilters creates the fade-out and trim filters
func buildFinalFilters(numImages int, duration, fadeDuration float64) (string, float64) {
	finalLength := (float64(numImages) * duration) - (float64(numImages-1) * fadeDuration)
	// After trim and setpts, fade-out should start at (finalLength - fadeDuration)
	fadeOutStart := finalLength - fadeDuration

	var filterComplex string
	// For a single image, use [v0] as there's no crossfade; otherwise use the last crossfade output [x(n-1)]
	var inputLabel string
	if numImages == 1 {
		inputLabel = "v0"
	} else {
		inputLabel = fmt.Sprintf("x%d", numImages-1)
	}
	// Apply trim first, then fade-out on the trimmed video
	filterComplex += fmt.Sprintf("[%s]trim=duration=%s,setpts=PTS-STARTPTS[xt]; ", inputLabel, formatSeconds(finalLength))
	filterComplex += fmt.Sprintf("[xt]fade=t=out:st=%s:d=%s[xfout]; ", formatSeconds(fadeOutStart), formatSeconds(fadeDuration))

	return filterComplex, finalLength
}

// AudioConfig contains audio processing configuration
type AudioConfig struct {
	Inputs      []string
	MapArgs     []string
	AudioFilter string
	HasAudio    bool
}

// findMusicFiles returns the list of mp3 files without logging
func findMusicFiles() ([]string, error) {
	musicFiles, err := filepath.Glob("*.mp3")
	if err != nil {
		return nil, fmt.Errorf("failed to list mp3 files: %v", err)
	}
	// Sort music files alphabetically for consistent ordering
	sort.Strings(musicFiles)
	return musicFiles, nil
}

// createAudioConcatFile creates a concat demuxer file for multiple audio files
func createAudioConcatFile(musicFiles []string) (string, error) {
	if len(musicFiles) == 0 {
		return "", fmt.Errorf("no music files provided")
	}

	// If only one file, return it directly
	if len(musicFiles) == 1 {
		return musicFiles[0], nil
	}

	// Create concat demuxer file for multiple audio files
	concatFile := "audio_concat.txt"
	var content strings.Builder

	for _, file := range musicFiles {
		// Escape single quotes in filenames
		escapedFile := strings.ReplaceAll(file, "'", "'\\''")
		content.WriteString(fmt.Sprintf("file '%s'\n", escapedFile))
	}

	if err := os.WriteFile(concatFile, []byte(content.String()), 0644); err != nil {
		return "", fmt.Errorf("failed to create concat file: %v", err)
	}

	return concatFile, nil
}

// getTotalAudioDurationSeconds returns the total audio length in seconds for multiple files
func getTotalAudioDurationSeconds(musicFiles []string) (float64, error) {
	if len(musicFiles) == 0 {
		return 0, fmt.Errorf("no music files provided")
	}

	totalDuration := 0.0

	for _, file := range musicFiles {
		dur, err := getAudioDurationSeconds(file)
		if err != nil {
			return 0, err
		}
		totalDuration += dur
	}

	return totalDuration, nil
}

// getAudioBitrateStr returns the audio bitrate of a file as an ffmpeg-compatible
// string (e.g. "192k"). Returns a fallback of "192k" on any error.
func getAudioBitrateStr(filename string) string {
	cmd := exec.Command("ffprobe", "-v", "error", "-select_streams", "a:0",
		"-show_entries", "stream=bit_rate", "-of", "default=noprint_wrappers=1:nokey=1", filename)
	out, err := cmd.Output()
	if err != nil {
		return "192k"
	}
	bitrateStr := strings.TrimSpace(string(out))
	if bitrateStr == "" || bitrateStr == "N/A" {
		// Fall back to format-level bit_rate (common for MP3)
		cmd2 := exec.Command("ffprobe", "-v", "error",
			"-show_entries", "format=bit_rate", "-of", "default=noprint_wrappers=1:nokey=1", filename)
		out2, err2 := cmd2.Output()
		if err2 != nil {
			return "192k"
		}
		bitrateStr = strings.TrimSpace(string(out2))
	}
	bps, err := strconv.Atoi(bitrateStr)
	if err != nil || bps <= 0 {
		return "192k"
	}
	return fmt.Sprintf("%dk", bps/1000)
}

// getAudioDurationSeconds returns audio length in seconds using ffprobe
func getAudioDurationSeconds(filename string) (float64, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", filename)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe duration failed: %v", err)
	}
	durStr := strings.TrimSpace(string(output))
	if durStr == "" {
		return 0, fmt.Errorf("empty duration from ffprobe")
	}
	dur, err := strconv.ParseFloat(durStr, 64)
	if err != nil {
		return 0, fmt.Errorf("parse duration: %w", err)
	}
	return dur, nil
}

// adjustDurationsToMusic scales image and transition durations to fit audio length
func adjustDurationsToMusic(duration, fadeDuration float64, numImages int, audioSeconds float64) (float64, float64, bool) {
	if audioSeconds <= 0 || numImages < 2 {
		return duration, fadeDuration, false
	}

	baseTotal := (float64(numImages) * duration) - (float64(numImages-1) * fadeDuration)
	if baseTotal <= 0 {
		return duration, fadeDuration, false
	}

	scale := audioSeconds / baseTotal
	newDuration := duration * scale
	newFade := fadeDuration * scale

	if newDuration < 1 {
		newDuration = 1
	}
	if newFade < 1 {
		newFade = 1
	}
	if newFade >= newDuration {
		newFade = math.Max(1, newDuration*0.5)
		if newFade >= newDuration {
			newDuration = newFade + 0.1
		}
	}

	// Fine-tune to reduce residual difference
	newTotal := (float64(numImages) * newDuration) - (float64(numImages-1) * newFade)
	diff := audioSeconds - newTotal

	// Adjust duration to absorb any residual mismatch after clamping
	if math.Abs(diff) >= 0.01 {
		newDuration += diff / float64(numImages)
		if newDuration < 1 {
			newDuration = 1
		}
	}

	// Safety check
	if newFade < 1 {
		newFade = 1
	}
	if newFade >= newDuration {
		newFade = math.Max(1, newDuration*0.5)
	}

	return newDuration, newFade, true
}

// setupAudioProcessing handles audio input and processing
func setupAudioProcessing(inputs []string, index int, startFadeOut, fadeDuration float64, musicFiles []string) AudioConfig {
	config := AudioConfig{
		Inputs:   inputs,
		HasAudio: len(musicFiles) > 0,
	}

	if config.HasAudio {
		if len(musicFiles) > 1 {
			// Multiple audio files: use concat demuxer
			fmt.Printf("Audio files found: %d MP3 files\n", len(musicFiles))
			for _, file := range musicFiles {
				fmt.Printf("  - %s\n", file)
			}

			// Create concat file
			concatFile, err := createAudioConcatFile(musicFiles)
			if err != nil {
				fmt.Printf("Warning: Failed to create audio concat: %v\n", err)
				fmt.Printf("Using single audio file: %s\n", musicFiles[0])
				config.Inputs = append(config.Inputs, "-i", musicFiles[0])
			} else {
				// Use concat demuxer to merge all audio files
				config.Inputs = append(config.Inputs, "-f", "concat", "-safe", "0", "-i", concatFile)
			}
		} else {
			// Single audio file
			fmt.Printf("Audio file found: %s\n", musicFiles[0])
			config.Inputs = append(config.Inputs, "-i", musicFiles[0])
		}

		// Apply audio normalization and fades
		// loudnorm normalizes volume to a consistent level across all tracks
		// Then apply fade in/out using the same duration as video fades
		config.AudioFilter = fmt.Sprintf("[%d:a]loudnorm=I=-16:TP=-1.5:LRA=11,afade=t=in:st=0:d=%s,afade=t=out:st=%s:d=%s[musicout]; ", index, formatSeconds(fadeDuration), formatSeconds(startFadeOut-fadeDuration), formatSeconds(fadeDuration))

		// Map video and audio
		config.MapArgs = []string{"-map", "[xfout]", "-map", "[musicout]", "-shortest", "video.mp4"}
	} else {
		fmt.Printf("No MP3 file found - generating video without audio\n")
		// Map only video
		config.MapArgs = []string{"-map", "[xfout]", "video.mp4"}
	}

	return config
}

// runFFmpegCommand executes the ffmpeg command with progress indication
func runFFmpegCommand(args []string, hasAudio bool) error {
	cmd := exec.Command("ffmpeg", args...)

	// Redirect FFmpeg logs to /dev/null.
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open /dev/null: %v", err)
	}
	cmd.Stdout = devNull
	cmd.Stderr = devNull

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ffmpeg start failed: %v", err)
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
				fmt.Printf("\r%s %s...", spinnerChars[i%len(spinnerChars)], message)
				i++
				time.Sleep(200 * time.Millisecond)
			}
		}
	}()

	if err := cmd.Wait(); err != nil {
		close(done)
		return fmt.Errorf("ffmpeg command failed: %v", err)
	}
	close(done)
	return nil
}

// displayVideoInfo shows the final video information
func displayVideoInfo(finalLength float64) {
	fmt.Printf("\n=== Video generated successfully! ===\n")
	fmt.Printf("File: video.mp4\n")

	// Get detailed video information
	if videoInfo, err := getVideoDetails("video.mp4"); err == nil {
		fmt.Printf("Resolution: %s (4K UHD)\n", videoInfo.Resolution)
		fmt.Printf("Duration: %.2f sec. (%.1fs actual)\n", finalLength, videoInfo.DurationSec)
		fmt.Printf("File Size: %.1f MB\n", videoInfo.FileSizeMB)
		fmt.Printf("Video Bitrate: %s\n", videoInfo.VideoBitrate)
		fmt.Printf("Audio Bitrate: %s\n", videoInfo.AudioBitrate)
		fmt.Printf("Framerate: %s\n", videoInfo.Framerate)
	} else {
		// Fallback to basic information if ffprobe fails
		fmt.Printf("Resolution: 4K UHD (%s)\n", resolution4K)
		fmt.Printf("Duration: %.2f sec.\n", finalLength)
		if fileInfo, err := os.Stat("video.mp4"); err == nil {
			sizeMB := float64(fileInfo.Size()) / (1024 * 1024)
			fmt.Printf("File Size: %.1f MB\n", sizeMB)
		}
	}
}

// GenerateVideo creates a video from already 4K images with crossfade transitions,
// audio fades, and optionally a Ken Burns effect applied to each image.
// If applyKenBurns is false, the images remain static.
// If exifOverlay is true, camera info will be displayed in the footer with specified fontSize.
func GenerateVideo(duration, fadeDuration int, applyKenBurns, exifOverlay bool, fontSize int, fitAudio bool) {
	durationSec := float64(duration)
	fadeSec := float64(fadeDuration)
	// Find all converted .jpg files (4K resolution).
	files, err := filepath.Glob("converted/*.jpg")
	if err != nil {
		log.Fatalf("Failed to list converted .jpg files: %v", err)
	}

	// Check if we have enough images to create a video
	if len(files) == 0 {
		log.Fatalf("No converted images found in 'converted/' directory.\nPlease convert your images first using the image conversion feature.")
	}

	if len(files) < 2 {
		log.Fatalf("Not enough images found. Need at least 2 images to create a video with transitions.\nFound: %d image(s) in 'converted/' directory.", len(files))
	}

	fmt.Printf("Generating video from %d images...\n", len(files))

	// Detect music files once
	musicFiles, err := findMusicFiles()
	if err != nil {
		log.Fatalf("%v", err)
	}

	// Auto-fit durations to music length if requested and audio exists
	if fitAudio {
		if len(musicFiles) == 0 {
			fmt.Printf("fit-audio requested but no MP3 found; using provided durations.\n")
		} else if len(musicFiles) > 1 {
			// Multiple audio files: get total duration
			if audioSeconds, err := getTotalAudioDurationSeconds(musicFiles); err == nil && audioSeconds > 0 {
				// Check minimum required audio length (5s per image, 1s transitions)
				minLength := float64(len(files)*5 - (len(files) - 1))
				if audioSeconds < minLength {
					log.Fatalf("Audio duration (%.1fs) is too short for %d images.\nMinimum required: %.1fs (5s per image × %d - 1s transitions × %d)\nPlease use fewer images or add more audio files.", audioSeconds, len(files), minLength, len(files), len(files)-1)
				}
				oldDuration, oldFade := durationSec, fadeSec
				durationSec, fadeSec, _ = adjustDurationsToMusic(durationSec, fadeSec, len(files), audioSeconds)
				fmt.Printf("Auto-fit to music (%.1fs total): duration %.2fs → %.2fs, transition %.2fs → %.2fs\n", audioSeconds, oldDuration, durationSec, oldFade, fadeSec)
			} else {
				fmt.Printf("fit-audio requested but could not read music duration; using provided durations.\n")
			}
		} else {
			// Single audio file
			if audioSeconds, err := getAudioDurationSeconds(musicFiles[0]); err == nil && audioSeconds > 0 {
				// Check minimum required audio length (5s per image, 1s transitions)
				minLength := float64(len(files)*5 - (len(files) - 1))
				if audioSeconds < minLength {
					log.Fatalf("Audio duration (%.1fs) is too short for %d images.\nMinimum required: %.1fs (5s per image × %d - 1s transitions × %d)\nPlease use fewer images or add more audio files.", audioSeconds, len(files), minLength, len(files), len(files)-1)
				}
				oldDuration, oldFade := durationSec, fadeSec
				durationSec, fadeSec, _ = adjustDurationsToMusic(durationSec, fadeSec, len(files), audioSeconds)
				fmt.Printf("Auto-fit to music (%.1fs): duration %.2fs → %.2fs, transition %.2fs → %.2fs\n", audioSeconds, oldDuration, durationSec, oldFade, fadeSec)
			} else {
				fmt.Printf("fit-audio requested but could not read music duration; using provided durations.\n")
			}
		}
	}

	// Build inputs and process each image
	inputs := []string{}
	filterComplex := ""

	for index, file := range files {
		inputs = append(inputs, "-loop", "1", "-t", formatSeconds(durationSec), "-i", file)
		videoFilter := processImageFilter(file, index, durationSec, fadeSec, applyKenBurns, exifOverlay, fontSize)
		filterComplex += fmt.Sprintf("%s[v%d]; ", videoFilter, index)
	}

	// Add crossfade transitions
	filterComplex += buildCrossfadeFilters(len(files), durationSec, fadeSec)

	// Add final filters and get duration
	finalFilters, finalLength := buildFinalFilters(len(files), durationSec, fadeSec)
	filterComplex += finalFilters

	// Setup audio processing
	totalDuration := (float64(len(files)) * durationSec) - (float64(len(files)-1) * fadeSec)
	startFadeOut := totalDuration - fadeSec
	audioConfig := setupAudioProcessing(inputs, len(files), startFadeOut, fadeSec, musicFiles)

	// Add audio filter to filter complex if audio is present
	if audioConfig.HasAudio {
		filterComplex += audioConfig.AudioFilter
	}

	// Write filter complex to a file to avoid Windows command line length limits
	filterComplexFile := "filter_complex.txt"
	if err := os.WriteFile(filterComplexFile, []byte(filterComplex), 0644); err != nil {
		log.Fatalf("Failed to write filter complex file: %v", err)
	}
	defer os.Remove(filterComplexFile)

	// Build complete FFmpeg command
	args := []string{"-y"}
	args = append(args, audioConfig.Inputs...)
	args = append(args, "-filter_complex_script", filterComplexFile)
	args = append(args, audioConfig.MapArgs...)
	args = append(args, getOptimalVideoSettings()...)

	// Add audio encoding settings if audio is present, preserving input bitrate
	if audioConfig.HasAudio {
		audioBitrate := getAudioBitrateStr(musicFiles[0])
		args = append(args, "-c:a", "aac", "-b:a", audioBitrate)
	}

	args = append(args, "-t", formatSeconds(finalLength))

	// Execute FFmpeg command
	if err := runFFmpegCommand(args, audioConfig.HasAudio); err != nil {
		log.Fatalf("Video generation failed: %v", err)
	}

	// Clean up concat file if it was created
	if len(musicFiles) > 1 {
		os.Remove("audio_concat.txt")
	}

	// Display final information
	displayVideoInfo(finalLength)
}

// getKenBurnsEffect generates a Ken Burns effect using a fixed zoompan expression.
// This approach is based on the method described in the Bannerbear blog.
// Updated with softer effects: slower zoom speed, lower max zoom, and reduced movement
func getKenBurnsEffect(duration float64) string {
	totalFrames := int(math.Round(duration * 30))
	if totalFrames < 1 {
		totalFrames = 1
	}
	offset := int(float64(totalFrames) * 1.2) // reduced offset for gentler movement

	// Define nine variants based on different focal positions with softer effects
	// Zoom speed reduced from 0.001 to 0.0005, max zoom reduced from 1.5 to 1.3
	centerExpr := "zoompan=zoom='min(zoom+0.0005,1.3)':x='iw/2-(iw/zoom/2)':y='ih/2-(ih/zoom/2)':d=%d:s=" + resolution4K
	topLeftExpr := "zoompan=zoom='min(zoom+0.0005,1.3)':x='iw/2-(iw/zoom/2)-%d':y='ih/2-(ih/zoom/2)-%d':d=%d:s=" + resolution4K
	topRightExpr := "zoompan=zoom='min(zoom+0.0005,1.3)':x='iw/2-(iw/zoom/2)+%d':y='ih/2-(ih/zoom/2)-%d':d=%d:s=" + resolution4K
	bottomLeftExpr := "zoompan=zoom='min(zoom+0.0005,1.3)':x='iw/2-(iw/zoom/2)-%d':y='ih/2-(ih/zoom/2)+%d':d=%d:s=" + resolution4K
	bottomRightExpr := "zoompan=zoom='min(zoom+0.0005,1.3)':x='iw/2-(iw/zoom/2)+%d':y='ih/2-(ih/zoom/2)+%d':d=%d:s=" + resolution4K
	leftExpr := "zoompan=zoom='min(zoom+0.0005,1.3)':x='iw/2-(iw/zoom/2)-%d':y='ih/2-(ih/zoom/2)':d=%d:s=" + resolution4K
	rightExpr := "zoompan=zoom='min(zoom+0.0005,1.3)':x='iw/2-(iw/zoom/2)+%d':y='ih/2-(ih/zoom/2)':d=%d:s=" + resolution4K
	topExpr := "zoompan=zoom='min(zoom+0.0005,1.3)':x='iw/2-(iw/zoom/2)':y='ih/2-(ih/zoom/2)-%d':d=%d:s=" + resolution4K
	bottomExpr := "zoompan=zoom='min(zoom+0.0005,1.3)':x='iw/2-(iw/zoom/2)':y='ih/2-(ih/zoom/2)+%d':d=%d:s=" + resolution4K

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
