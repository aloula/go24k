package utils

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	linuxOS            = "linux"
	resolution4K       = "3840x2160"
	resolutionFullHD   = "1920x1080"
	outputVideoLegacy  = "video.mp4"
	outputVideoUHD     = "video_uhd.mp4"
	outputVideoFHD     = "video_fhd.mp4"
	kenBurnsModeLow    = "low"
	kenBurnsModeMedium = "medium"
	kenBurnsModeHigh   = "high"
)

// activeResolution holds the target output resolution for the current run.
// Set at the start of GenerateVideo() based on the fullHD flag.
var activeResolution = resolution4K

// activeFPS holds the target output framerate for the current run.
// Set at the start of GenerateVideo().
var activeFPS = 30

// activeKenBurnsMode holds the motion profile for Ken Burns when enabled.
// Set at the start of GenerateVideo().
var activeKenBurnsMode = kenBurnsModeHigh

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
	cmd := newExecCommand("ffprobe",
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
	if duration, err := getMediaDurationSeconds(filename); err == nil {
		info.DurationSec = duration
	}

	outputStr, err := runFFProbe(filename)
	if err != nil {
		// Set defaults if ffprobe fails
		info.Framerate = fmt.Sprintf("%d fps", activeFPS)
		info.Resolution = activeResolution
		info.AudioBitrate = "No audio"
		return info, err
	}

	info.VideoBitrate, info.Framerate, info.Resolution = extractVideoInfo(outputStr)
	info.AudioBitrate = extractAudioInfo(outputStr)

	// Set defaults if not found
	if info.Framerate == "" {
		info.Framerate = fmt.Sprintf("%d fps", activeFPS)
	}
	if info.Resolution == "" {
		info.Resolution = activeResolution
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
				resolution = activeResolution // We know our output resolution
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
	cmd := newExecCommand("ffmpeg", "-encoders")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	if !strings.Contains(string(output), "h264_nvenc") {
		return false
	}

	// Test if NVENC actually works (avoid false positives in WSL/ARM systems)
	// Some systems report NVENC support but can't actually use it
	testCmd := newExecCommand("ffmpeg", "-f", "lavfi", "-i", "testsrc=duration=0.1:size=320x240:rate=1",
		"-c:v", "h264_nvenc", "-f", "null", "-")
	err = testCmd.Run()
	return err == nil
}

func checkQSVAvailable() bool {
	// First check if encoder is listed
	cmd := newExecCommand("ffmpeg", "-encoders")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	if !strings.Contains(string(output), "h264_qsv") {
		return false
	}

	// Test if QSV actually works
	testCmd := newExecCommand("ffmpeg", "-f", "lavfi", "-i", "testsrc=duration=0.1:size=320x240:rate=1",
		"-c:v", "h264_qsv", "-f", "null", "-")
	err = testCmd.Run()
	return err == nil
}

func checkAMFAvailable() bool {
	// First check if encoder is listed
	cmd := newExecCommand("ffmpeg", "-encoders")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	if !strings.Contains(string(output), "h264_amf") {
		return false
	}

	// Test if AMF actually works
	testCmd := newExecCommand("ffmpeg", "-f", "lavfi", "-i", "testsrc=duration=0.1:size=320x240:rate=1",
		"-c:v", "h264_amf", "-f", "null", "-")
	err = testCmd.Run()
	return err == nil
}

func checkMediaFoundationAvailable() bool {
	// First check if encoder is listed
	cmd := newExecCommand("ffmpeg", "-encoders")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	if !strings.Contains(string(output), "h264_mf") {
		return false
	}

	// Test if Media Foundation actually works
	testCmd := newExecCommand("ffmpeg", "-f", "lavfi", "-i", "testsrc=duration=0.1:size=320x240:rate=1",
		"-c:v", "h264_mf", "-f", "null", "-")
	err = testCmd.Run()
	return err == nil
}

func checkVAAPIAvailable() bool {
	// First check if encoder is listed
	cmd := newExecCommand("ffmpeg", "-encoders")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	if !strings.Contains(string(output), "h264_vaapi") {
		return false
	}

	// Test if VAAPI actually works
	testCmd := newExecCommand("ffmpeg", "-f", "lavfi", "-i", "testsrc=duration=0.1:size=320x240:rate=1",
		"-c:v", "h264_vaapi", "-f", "null", "-")
	err = testCmd.Run()
	return err == nil
}

func checkVideoToolboxAvailable() bool {
	cmd := newExecCommand("ffmpeg", "-encoders")
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
		"-r", strconv.Itoa(activeFPS),
		"-s", activeResolution,
	}

	h264Level := "5.1"
	if activeResolution == resolution4K && activeFPS >= 60 {
		h264Level = "5.2"
	}

	// Priority order: NVENC > VideoToolbox (macOS) > Media Foundation (Windows) > QSV > AMF > VAAPI > CPU
	if hasNVENC {
		// NVIDIA GPU acceleration
		fmt.Printf("Hardware: NVIDIA NVENC detected - using GPU acceleration\n")
		settings = append(settings,
			"-c:v", "h264_nvenc",
			"-preset", "slow",
			"-profile:v", "high",
			"-level", h264Level,
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
			"-level", h264Level,
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
			"-level", h264Level,
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
			"-level", h264Level,
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
			"-level", h264Level,
			"-maxrate", "12M",
			"-bufsize", "24M",
		)
	} else if hasVAAPI {
		// Linux VAAPI (Intel/AMD integrated graphics)
		fmt.Printf("Hardware: VAAPI detected - using Linux hardware acceleration\n")
		settings = append(settings,
			"-c:v", "h264_vaapi",
			"-profile:v", "high",
			"-level", h264Level,
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
			"-level", h264Level,
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

// supersampledResolution returns a 2× upscaled version of activeResolution.
// The zoompan filter operates at this larger canvas size to eliminate the
// sub-pixel rounding jitter that otherwise appears at lower resolutions.
func supersampledResolution() string {
	parts := strings.SplitN(activeResolution, "x", 2)
	if len(parts) != 2 {
		return activeResolution
	}
	w, err1 := strconv.Atoi(parts[0])
	h, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return activeResolution
	}
	return fmt.Sprintf("%dx%d", w*2, h*2)
}

// processImageFilter creates the video filter for a single image
func processImageFilter(file string, index int, duration, fadeDuration float64, applyKenBurns, exifOverlay bool, fontSize int) string {
	var videoFilter string

	if applyKenBurns {
		// Scale up to 2× before zoompan to prevent sub-pixel jitter, then scale back down.
		superRes := supersampledResolution()
		superResScale := strings.Replace(superRes, "x", ":", 1)
		activeResScale := strings.Replace(activeResolution, "x", ":", 1)
		effect := getKenBurnsEffect(duration)
		if index == 0 {
			// For the first image, apply the effect followed by a fade-in.
			videoFilter = fmt.Sprintf("[0:v]scale=%s,%s,scale=%s,fade=t=in:st=0:d=%s,fps=%d,settb=AVTB,setsar=1,format=yuv420p", superResScale, effect, activeResScale, formatSeconds(fadeDuration), activeFPS)
		} else {
			videoFilter = fmt.Sprintf("[%d:v]scale=%s,%s,scale=%s,fps=%d,settb=AVTB,setsar=1,format=yuv420p", index, superResScale, effect, activeResScale, activeFPS)
		}
	} else {
		// Static: no zoom/pan effect.
		if index == 0 {
			videoFilter = fmt.Sprintf("[0:v]fps=%d,settb=AVTB,setsar=1,format=yuv420p,fade=t=in:st=0:d=%s", activeFPS, formatSeconds(fadeDuration))
		} else {
			videoFilter = fmt.Sprintf("[%d:v]fps=%d,settb=AVTB,setsar=1,format=yuv420p", index, activeFPS)
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

// buildCrossfadeFilters creates crossfade transitions for variable media segment lengths.
func buildCrossfadeFilters(segmentDurations []float64, fadeDuration float64) string {
	var filterComplex string
	numItems := len(segmentDurations)
	if numItems < 2 {
		return filterComplex
	}

	cumulative := 0.0

	// Generate crossfade transitions.
	for i := 0; i < numItems-1; i++ {
		cumulative += segmentDurations[i]
		next := i + 1
		offset := cumulative - (float64(i+1) * fadeDuration)
		if i == 0 {
			filterComplex += fmt.Sprintf("[v%d][v%d]xfade=transition=fade:duration=%s:offset=%s[x%d]; ", i, next, formatSeconds(fadeDuration), formatSeconds(offset), next)
		} else {
			filterComplex += fmt.Sprintf("[x%d][v%d]xfade=transition=fade:duration=%s:offset=%s[x%d]; ", i, next, formatSeconds(fadeDuration), formatSeconds(offset), next)
		}
	}

	return filterComplex
}

func calculateFinalLength(segmentDurations []float64, fadeDuration float64) float64 {
	total := 0.0
	for _, d := range segmentDurations {
		total += d
	}
	if len(segmentDurations) < 2 {
		return total
	}
	return total - (float64(len(segmentDurations)-1) * fadeDuration)
}

// buildFinalFilters creates the fade-out and trim filters
func buildFinalFilters(segmentDurations []float64, fadeDuration float64) (string, float64) {
	numItems := len(segmentDurations)
	finalLength := calculateFinalLength(segmentDurations, fadeDuration)
	// After trim and setpts, fade-out should start at (finalLength - fadeDuration)
	fadeOutStart := finalLength - fadeDuration

	var filterComplex string
	// For a single image, use [v0] as there's no crossfade; otherwise use the last crossfade output [x(n-1)]
	var inputLabel string
	if numItems == 1 {
		inputLabel = "v0"
	} else {
		inputLabel = fmt.Sprintf("x%d", numItems-1)
	}
	// Apply trim first, then fade-out on the trimmed video
	filterComplex += fmt.Sprintf("[%s]trim=duration=%s,setpts=PTS-STARTPTS[xt]; ", inputLabel, formatSeconds(finalLength))
	filterComplex += fmt.Sprintf("[xt]fade=t=out:st=%s:d=%s[xfout]; ", formatSeconds(fadeOutStart), formatSeconds(fadeDuration))

	return filterComplex, finalLength
}

// AudioConfig contains audio processing configuration
type AudioConfig struct {
	Inputs             []string
	MapArgs            []string
	AudioFilter        string
	HasAudio           bool
	AudioBitrateSource string
}

// MediaInput represents an item (image or video) to be included in the timeline.
type MediaInput struct {
	Path            string
	IsImage         bool
	HasAudio        bool
	SegmentDuration float64
	CapturedAt      time.Time
	HasCapturedAt   bool
	SortName        string
}

// findVideoFiles returns video files in the current directory based on selected options.
func findVideoFiles(includeVideos, includeMOV bool) ([]string, error) {
	if !includeVideos && !includeMOV {
		return []string{}, nil
	}

	allowedExtensions := map[string]struct{}{}
	if includeVideos {
		for _, ext := range []string{".mp4", ".mov", ".mkv", ".avi", ".webm", ".m4v"} {
			allowedExtensions[ext] = struct{}{}
		}
	} else if includeMOV {
		allowedExtensions[".mov"] = struct{}{}
	}

	generatedOutputs := generatedOutputVideoNames()
	var files []string

	directoryEntries, err := os.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("failed to read current directory: %v", err)
	}

	for _, entry := range directoryEntries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if _, isGeneratedOutput := generatedOutputs[strings.ToLower(filepath.Base(name))]; isGeneratedOutput {
			continue
		}

		ext := strings.ToLower(filepath.Ext(name))
		if _, isAllowed := allowedExtensions[ext]; !isAllowed {
			continue
		}

		files = append(files, name)
	}

	sort.Strings(files)
	return files, nil
}

func generatedOutputVideoNames() map[string]struct{} {
	return map[string]struct{}{
		strings.ToLower(outputVideoLegacy): {},
		strings.ToLower(outputVideoUHD):    {},
		strings.ToLower(outputVideoFHD):    {},
	}
}

func outputVideoFilename() string {
	if activeResolution == resolutionFullHD {
		return outputVideoFHD
	}
	return outputVideoUHD
}

// collectMediaInputs builds a sorted timeline from converted images and optional videos.
// Default ordering is capture metadata time, with filename as deterministic fallback.
// If orderByFilename is true, ordering uses filenames only.
func collectMediaInputs(imageDuration float64, includeVideos, includeMOV, orderByFilename bool) ([]MediaInput, error) {
	imageFiles, err := filepath.Glob("converted/*.jpg")
	if err != nil {
		return nil, fmt.Errorf("failed to list converted images: %v", err)
	}
	sort.Strings(imageFiles)

	var media []MediaInput
	for _, file := range imageFiles {
		capturedAt, hasCapturedAt := extractImageTimestampFromConvertedName(file)
		if !hasCapturedAt {
			capturedAt, hasCapturedAt = extractCaptureTimeFromFilename(file)
		}
		sortName := mediaSortName(file)
		if orderByFilename {
			sortName = resolveImageSortName(file)
		}
		media = append(media, MediaInput{
			Path:            file,
			IsImage:         true,
			SegmentDuration: imageDuration,
			CapturedAt:      capturedAt,
			HasCapturedAt:   hasCapturedAt,
			SortName:        sortName,
		})
	}

	if includeVideos || includeMOV {
		videoFiles, err := findVideoFiles(includeVideos, includeMOV)
		if err != nil {
			return nil, err
		}

		for _, file := range videoFiles {
			duration, err := getMediaDurationSeconds(file)
			if err != nil {
				return nil, fmt.Errorf("failed to read video duration for %s: %v", file, err)
			}
			capturedAt, hasCapturedAt := getVideoCaptureTime(file)
			if !hasCapturedAt {
				capturedAt, hasCapturedAt = extractCaptureTimeFromFilename(file)
			}
			hasAudio, err := hasAudioStream(file)
			if err != nil {
				return nil, fmt.Errorf("failed to inspect audio stream for %s: %v", file, err)
			}
			if duration <= 0 {
				return nil, fmt.Errorf("video %s has invalid duration %.2f", file, duration)
			}

			videoSortName := mediaSortName(file)
			media = append(media, MediaInput{
				Path:            file,
				IsImage:         false,
				HasAudio:        hasAudio,
				SegmentDuration: duration,
				CapturedAt:      capturedAt,
				HasCapturedAt:   hasCapturedAt,
				SortName:        videoSortName,
			})
		}
	}

	sort.Slice(media, func(i, j int) bool {
		if orderByFilename {
			return media[i].SortName < media[j].SortName
		}

		if media[i].HasCapturedAt && media[j].HasCapturedAt {
			if !media[i].CapturedAt.Equal(media[j].CapturedAt) {
				return media[i].CapturedAt.Before(media[j].CapturedAt)
			}
		}
		return media[i].SortName < media[j].SortName
	})

	if len(media) == 0 {
		if includeVideos || includeMOV {
			return nil, fmt.Errorf("no converted images or supported videos found")
		}
		return nil, fmt.Errorf("no converted images found in 'converted/' directory.\nPlease convert your images first using the image conversion feature")
	}

	if len(media) < 2 {
		return nil, fmt.Errorf("not enough media found. Need at least 2 images/videos to create a video with transitions.\nFound: %d item(s)", len(media))
	}

	return media, nil
}

func mediaSortName(path string) string {
	name := strings.TrimSpace(path)
	if name == "" {
		return ""
	}
	return strings.ToLower(filepath.Base(name))
}

// resolveImageSortName returns the preferred sort key for converted images.
// In filename-order mode we sort by original source filename when possible.
func resolveImageSortName(convertedPath string) string {
	original := GetOriginalFilename(convertedPath)
	if original != "" {
		return mediaSortName(original)
	}
	return mediaSortName(convertedPath)
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
	cmd := newExecCommand("ffprobe", "-v", "error", "-select_streams", "a:0",
		"-show_entries", "stream=bit_rate", "-of", "default=noprint_wrappers=1:nokey=1", filename)
	out, err := cmd.Output()
	if err != nil {
		return "192k"
	}
	bitrateStr := strings.TrimSpace(string(out))
	if bitrateStr == "" || bitrateStr == "N/A" {
		// Fall back to format-level bit_rate (common for MP3)
		cmd2 := newExecCommand("ffprobe", "-v", "error",
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
func getMediaDurationSeconds(filename string) (float64, error) {
	cmd := newExecCommand("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", filename)
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

// getAudioDurationSeconds returns audio length in seconds using ffprobe
func getAudioDurationSeconds(filename string) (float64, error) {
	return getMediaDurationSeconds(filename)
}

func hasAudioStream(filename string) (bool, error) {
	cmd := newExecCommand("ffprobe", "-v", "error", "-select_streams", "a", "-show_entries", "stream=index", "-of", "csv=p=0", filename)
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("ffprobe audio stream failed: %v", err)
	}
	return strings.TrimSpace(string(output)) != "", nil
}

func extractImageTimestampFromConvertedName(path string) (time.Time, bool) {
	base := filepath.Base(path)
	base = trimConvertedImageResolutionSuffix(base)

	if len(base) >= len("20060102_150405") {
		candidate := base[:len("20060102_150405")]
		if ts, err := time.Parse("20060102_150405", candidate); err == nil {
			return ts, true
		}
	}

	return time.Time{}, false
}

func extractCaptureTimeFromFilename(path string) (time.Time, bool) {
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	if name == "" {
		return time.Time{}, false
	}

	if ts, ok := parseTimestampCandidate(name); ok {
		return ts, true
	}

	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(\d{8}[_-]\d{6})`),
		regexp.MustCompile(`(\d{14})`),
		regexp.MustCompile(`(\d{4}[-_]\d{2}[-_]\d{2}[T _-]\d{2}[:._-]\d{2}[:._-]\d{2})`),
	}

	for _, pattern := range patterns {
		match := pattern.FindString(name)
		if match == "" {
			continue
		}
		if ts, ok := parseTimestampCandidate(match); ok {
			return ts, true
		}
	}

	return time.Time{}, false
}

func parseTimestampCandidate(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}

	if ts, err := time.Parse("20060102_150405", value); err == nil {
		return ts, true
	}
	if ts, err := time.Parse("20060102-150405", value); err == nil {
		return ts, true
	}
	if ts, err := time.Parse("20060102150405", value); err == nil {
		return ts, true
	}

	normalized := strings.NewReplacer("_", "-", ".", ":").Replace(value)
	layouts := []string{
		"2006-01-02-15-04-05",
		"2006-01-02-15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if ts, err := time.Parse(layout, normalized); err == nil {
			return ts, true
		}
	}

	return time.Time{}, false
}

func parseVideoCreationTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("empty timestamp")
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05Z07:00",
	}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported timestamp format: %s", value)
}

func getVideoCaptureTime(filename string) (time.Time, bool) {
	cmd := newExecCommand("ffprobe", "-v", "error",
		"-show_entries", "format_tags=creation_time:stream_tags=creation_time",
		"-of", "default=noprint_wrappers=1:nokey=1", filename)
	output, err := cmd.Output()
	if err != nil {
		return time.Time{}, false
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if ts, parseErr := parseVideoCreationTime(line); parseErr == nil {
			return ts, true
		}
	}

	return time.Time{}, false
}

func buildTimelineOffsets(mediaInputs []MediaInput, fadeDuration float64) []float64 {
	offsets := make([]float64, len(mediaInputs))
	currentOffset := 0.0

	for index := 1; index < len(mediaInputs); index++ {
		currentOffset += mediaInputs[index-1].SegmentDuration - fadeDuration
		offsets[index] = currentOffset
	}

	return offsets
}

func clipAudioFadeDuration(segmentDuration, fadeDuration float64) float64 {
	if segmentDuration <= 0 {
		return 0
	}

	maxFade := segmentDuration / 4
	if maxFade < 0.25 {
		maxFade = 0.25
	}

	if fadeDuration < maxFade {
		return fadeDuration
	}

	return maxFade
}

func joinFilterInputs(labels []string) string {
	var builder strings.Builder
	for _, label := range labels {
		builder.WriteString("[")
		builder.WriteString(label)
		builder.WriteString("]")
	}
	return builder.String()
}

// buildMusicMuteExpression returns an FFmpeg volume expression that silences the
// background music whenever a video clip with audio is playing, with smooth
// fade-out before the clip starts and fade-in after it ends.
func buildMusicMuteExpression(mediaInputs []MediaInput, offsets []float64, fadeDuration float64) string {
	var parts []string

	for i, media := range mediaInputs {
		if media.IsImage || !media.HasAudio {
			continue
		}

		fadeLen := math.Max(clipAudioFadeDuration(media.SegmentDuration, fadeDuration), 0.001)
		clipStart := offsets[i]
		clipEnd := clipStart + media.SegmentDuration
		muteStart := math.Max(clipStart-fadeLen, 0)
		muteEnd := clipEnd + fadeLen

		// Volume envelope: from muteStart fade smoothly to 0 at clipStart,
		// stay silent until clipEnd, then fade back to 1 by muteEnd.
		expr := fmt.Sprintf(
			"if(lt(t,%.3f),1,if(lt(t,%.3f),(%.3f-t)/%.3f,if(lt(t,%.3f),0,if(lt(t,%.3f),(t-%.3f)/%.3f,1))))",
			muteStart,
			clipStart, clipStart, fadeLen,
			clipEnd,
			muteEnd, clipEnd, fadeLen,
		)
		parts = append(parts, expr)
	}

	if len(parts) == 0 {
		return "1"
	}
	if len(parts) == 1 {
		return parts[0]
	}
	// Multiple clips: take the minimum so all mute windows are respected.
	result := parts[0]
	for _, p := range parts[1:] {
		result = "min(" + result + "," + p + ")"
	}
	return result
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
func setupAudioProcessing(inputs []string, mediaInputs []MediaInput, finalLength, fadeDuration float64, musicFiles []string, keepVideoAudio bool) AudioConfig {
	config := AudioConfig{
		Inputs: inputs,
	}

	hasMusic := len(musicFiles) > 0
	musicInputIndex := len(mediaInputs)

	if hasMusic {
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
		config.AudioBitrateSource = musicFiles[0]
	}

	offsets := buildTimelineOffsets(mediaInputs, fadeDuration)
	videoAudioLabels := []string{}

	if keepVideoAudio {
		for index, media := range mediaInputs {
			if media.IsImage || !media.HasAudio {
				continue
			}

			fadeLength := clipAudioFadeDuration(media.SegmentDuration, fadeDuration)
			fadeOutStart := media.SegmentDuration - fadeLength
			delayMs := int(math.Round(offsets[index] * 1000))
			label := fmt.Sprintf("clipaudio%d", len(videoAudioLabels))

			config.AudioFilter += fmt.Sprintf("[%d:a]aformat=sample_fmts=fltp:sample_rates=48000:channel_layouts=stereo,aresample=48000,atrim=duration=%s,asetpts=PTS-STARTPTS,afade=t=in:st=0:d=%s,afade=t=out:st=%s:d=%s,adelay=%d|%d[%s]; ", index, formatSeconds(media.SegmentDuration), formatSeconds(fadeLength), formatSeconds(fadeOutStart), formatSeconds(fadeLength), delayMs, delayMs, label)
			videoAudioLabels = append(videoAudioLabels, label)

			if config.AudioBitrateSource == "" {
				config.AudioBitrateSource = media.Path
			}
		}

		if len(videoAudioLabels) > 0 {
			fmt.Printf("Keeping audio from %d input video(s)\n", len(videoAudioLabels))
		} else {
			fmt.Printf("keep-video-audio requested, but no input videos with audio were found\n")
		}
	}

	var clipAudioBusLabel string
	if len(videoAudioLabels) == 1 {
		clipAudioBusLabel = videoAudioLabels[0]
	} else if len(videoAudioLabels) > 1 {
		clipAudioBusLabel = "clipaudiobus"
		config.AudioFilter += fmt.Sprintf("%samix=inputs=%d:duration=longest:normalize=0:dropout_transition=%s[%s]; ", joinFilterInputs(videoAudioLabels), len(videoAudioLabels), formatSeconds(fadeDuration), clipAudioBusLabel)
	}

	var finalAudioLabel string
	if hasMusic {
		musicFadeOutStart := finalLength - fadeDuration
		if musicFadeOutStart < 0 {
			musicFadeOutStart = 0
		}

		config.AudioFilter += fmt.Sprintf("[%d:a]aformat=sample_fmts=fltp:sample_rates=48000:channel_layouts=stereo,aresample=48000,loudnorm=I=-16:TP=-1.5:LRA=11,atrim=duration=%s,asetpts=PTS-STARTPTS,afade=t=in:st=0:d=%s,afade=t=out:st=%s:d=%s[musicout]; ", musicInputIndex, formatSeconds(finalLength), formatSeconds(fadeDuration), formatSeconds(musicFadeOutStart), formatSeconds(fadeDuration))

		if clipAudioBusLabel != "" {
			// Silence the music precisely during each clip segment with smooth fades.
			muteExpr := buildMusicMuteExpression(mediaInputs, offsets, fadeDuration)
			config.AudioFilter += fmt.Sprintf("[musicout]volume='%s':eval=frame[musicmuted]; ", muteExpr)
			config.AudioFilter += fmt.Sprintf("[musicmuted][%s]amix=inputs=2:duration=first:normalize=0:dropout_transition=%s[mixedaudio]; ", clipAudioBusLabel, formatSeconds(fadeDuration))
			finalAudioLabel = "mixedaudio"
		} else {
			finalAudioLabel = "musicout"
		}
	} else if clipAudioBusLabel != "" {
		finalAudioLabel = clipAudioBusLabel
	}

	config.HasAudio = finalAudioLabel != ""
	if config.HasAudio {
		if !hasMusic && clipAudioBusLabel != "" {
			fmt.Printf("No MP3 file found - using input video audio only\n")
		}
		config.MapArgs = []string{"-map", "[xfout]", "-map", fmt.Sprintf("[%s]", finalAudioLabel), "-shortest"}
	} else {
		fmt.Printf("No MP3 file found - generating video without audio\n")
		config.MapArgs = []string{"-map", "[xfout]"}
	}

	return config
}

// processVideoFilter creates a normalized filter for a video input.
// The source video is centered on the target canvas without stretching to preserve its framing.
func processVideoFilter(index int, fadeDuration float64) string {
	res := activeResolution
	parts := strings.SplitN(res, "x", 2)
	w, h := parts[0], parts[1]
	base := fmt.Sprintf("[%d:v]fps=%d,settb=AVTB,scale='if(gt(iw,%s)+gt(ih,%s),%s,iw)':'if(gt(iw,%s)+gt(ih,%s),%s,ih)':force_original_aspect_ratio=decrease,pad=%s:%s:(ow-iw)/2:(oh-ih)/2:black,setsar=1,format=yuv420p,setpts=PTS-STARTPTS", index, activeFPS, w, h, w, w, h, h, w, h)
	if index == 0 {
		return fmt.Sprintf("%s,fade=t=in:st=0:d=%s", base, formatSeconds(fadeDuration))
	}
	return base
}

// runFFmpegCommand executes the ffmpeg command with progress indication
func runFFmpegCommand(args []string, hasAudio bool) error {
	cmd := newExecCommand("ffmpeg", args...)
	var stderr bytes.Buffer

	// Keep stdout quiet but retain stderr for actionable failures.
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open /dev/null: %v", err)
	}
	defer devNull.Close()
	cmd.Stdout = devNull
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ffmpeg start failed: %v", err)
	}

	showSpinner := os.Getenv("GO24K_INTERNAL_CLI") != "1"
	var done chan struct{}
	if showSpinner {
		done = make(chan struct{})
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
	}

	if err := cmd.Wait(); err != nil {
		if showSpinner {
			close(done)
		}
		stderrOutput := strings.TrimSpace(stderr.String())
		if stderrOutput != "" {
			return fmt.Errorf("ffmpeg command failed: %v\n%s", err, stderrOutput)
		}
		return fmt.Errorf("ffmpeg command failed: %v", err)
	}
	if showSpinner {
		close(done)
	}
	return nil
}

// displayVideoInfo shows the final video information
func displayVideoInfo(outputFilename string, finalLength float64) {
	resLabel := "4K UHD"
	if activeResolution == resolutionFullHD {
		resLabel = "Full HD"
	}
	fmt.Printf("\n=== Video generated successfully! ===\n")
	fmt.Printf("File: %s\n", outputFilename)

	// Get detailed video information
	if videoInfo, err := getVideoDetails(outputFilename); err == nil {
		fmt.Printf("Resolution: %s (%s)\n", videoInfo.Resolution, resLabel)
		fmt.Printf("Duration: %.2f sec. (%.1fs actual)\n", finalLength, videoInfo.DurationSec)
		fmt.Printf("File Size: %.1f MB\n", videoInfo.FileSizeMB)
		fmt.Printf("Video Bitrate: %s\n", videoInfo.VideoBitrate)
		fmt.Printf("Audio Bitrate: %s\n", videoInfo.AudioBitrate)
		fmt.Printf("Framerate: %s\n", videoInfo.Framerate)
	} else {
		// Fallback to basic information if ffprobe fails
		fmt.Printf("Resolution: %s (%s)\n", activeResolution, resLabel)
		fmt.Printf("Duration: %.2f sec.\n", finalLength)
		if fileInfo, err := os.Stat(outputFilename); err == nil {
			sizeMB := float64(fileInfo.Size()) / (1024 * 1024)
			fmt.Printf("File Size: %.1f MB\n", sizeMB)
		}
	}
}

// GenerateVideo creates a video from converted images with crossfade transitions,
// audio fades, and optionally a Ken Burns effect applied to each image.
// If applyKenBurns is false, the images remain static.
// If exifOverlay is true, camera info will be displayed in the footer with specified fontSize.
// If fullHD is true, the output resolution will be Full HD (1920x1080) instead of 4K UHD (3840x2160).
func GenerateVideo(duration, fadeDuration int, applyKenBurns, exifOverlay bool, fontSize int, fitAudio, includeVideos, includeMOV, keepVideoAudio, fullHD bool, fps int, orderByFilename bool, kenBurnsMode string) {
	// Set active resolution based on the fullHD flag.
	if fullHD {
		activeResolution = resolutionFullHD
	} else {
		activeResolution = resolution4K
	}

	if fps != 60 {
		fps = 30
	}
	activeFPS = fps
	activeKenBurnsMode = normalizeKenBurnsMode(kenBurnsMode)
	outputFilename := outputVideoFilename()

	durationSec := float64(duration)
	fadeSec := float64(fadeDuration)

	mediaInputs, err := collectMediaInputs(durationSec, includeVideos, includeMOV, orderByFilename)
	if err != nil {
		log.Fatalf("%v", err)
	}

	if fadeSec <= 0 {
		log.Fatalf("transition duration must be greater than 0")
	}

	imageCount := 0
	videoCount := 0
	for _, media := range mediaInputs {
		if media.IsImage {
			imageCount++
		} else {
			videoCount++
		}
		if media.SegmentDuration <= fadeSec {
			log.Fatalf("media item %s has duration %.2fs which must be greater than transition %.2fs", media.Path, media.SegmentDuration, fadeSec)
		}
	}

	fmt.Printf("Generating video from %d media items (%d images, %d videos)...\n", len(mediaInputs), imageCount, videoCount)
	if applyKenBurns {
		fmt.Printf("Ken Burns mode: %s\n", activeKenBurnsMode)
	} else {
		fmt.Printf("Ken Burns mode: disabled (-static)\n")
	}

	// Detect music files once
	musicFiles, err := findMusicFiles()
	if err != nil {
		log.Fatalf("%v", err)
	}

	// Auto-fit durations to music length if requested and audio exists
	if fitAudio {
		if videoCount > 0 {
			fmt.Printf("fit-audio with mixed images/videos keeps original video lengths and uses provided image/transition durations.\n")
		}
		if len(musicFiles) == 0 {
			fmt.Printf("fit-audio requested but no MP3 found; using provided durations.\n")
		} else if len(musicFiles) > 1 {
			// Multiple audio files: get total duration
			if audioSeconds, err := getTotalAudioDurationSeconds(musicFiles); err == nil && audioSeconds > 0 {
				if videoCount > 0 {
					fmt.Printf("fit-audio skipped in mixed media mode; keeping source video durations.\n")
				} else {
					// Check minimum required audio length (5s per image, 1s transitions)
					minLength := float64(len(mediaInputs)*5 - (len(mediaInputs) - 1))
					if audioSeconds < minLength {
						log.Fatalf("Audio duration (%.1fs) is too short for %d images.\nMinimum required: %.1fs (5s per image × %d - 1s transitions × %d)\nPlease use fewer images or add more audio files.", audioSeconds, len(mediaInputs), minLength, len(mediaInputs), len(mediaInputs)-1)
					}
					oldDuration, oldFade := durationSec, fadeSec
					durationSec, fadeSec, _ = adjustDurationsToMusic(durationSec, fadeSec, len(mediaInputs), audioSeconds)
					fmt.Printf("Auto-fit to music (%.1fs total): duration %.2fs → %.2fs, transition %.2fs → %.2fs\n", audioSeconds, oldDuration, durationSec, oldFade, fadeSec)

					for i := range mediaInputs {
						if mediaInputs[i].IsImage {
							mediaInputs[i].SegmentDuration = durationSec
						}
					}
				}
			} else {
				fmt.Printf("fit-audio requested but could not read music duration; using provided durations.\n")
			}
		} else {
			// Single audio file
			if audioSeconds, err := getAudioDurationSeconds(musicFiles[0]); err == nil && audioSeconds > 0 {
				if videoCount > 0 {
					fmt.Printf("fit-audio skipped in mixed media mode; keeping source video durations.\n")
				} else {
					// Check minimum required audio length (5s per image, 1s transitions)
					minLength := float64(len(mediaInputs)*5 - (len(mediaInputs) - 1))
					if audioSeconds < minLength {
						log.Fatalf("Audio duration (%.1fs) is too short for %d images.\nMinimum required: %.1fs (5s per image × %d - 1s transitions × %d)\nPlease use fewer images or add more audio files.", audioSeconds, len(mediaInputs), minLength, len(mediaInputs), len(mediaInputs)-1)
					}
					oldDuration, oldFade := durationSec, fadeSec
					durationSec, fadeSec, _ = adjustDurationsToMusic(durationSec, fadeSec, len(mediaInputs), audioSeconds)
					fmt.Printf("Auto-fit to music (%.1fs): duration %.2fs → %.2fs, transition %.2fs → %.2fs\n", audioSeconds, oldDuration, durationSec, oldFade, fadeSec)

					for i := range mediaInputs {
						if mediaInputs[i].IsImage {
							mediaInputs[i].SegmentDuration = durationSec
						}
					}
				}
			} else {
				fmt.Printf("fit-audio requested but could not read music duration; using provided durations.\n")
			}
		}
	}

	// Build inputs and process each image
	inputs := []string{}
	filterComplex := ""

	segmentDurations := make([]float64, 0, len(mediaInputs))
	for index, media := range mediaInputs {
		var videoFilter string
		if media.IsImage {
			inputs = append(inputs, "-loop", "1", "-t", formatSeconds(media.SegmentDuration), "-i", media.Path)
			videoFilter = processImageFilter(media.Path, index, media.SegmentDuration, fadeSec, applyKenBurns, exifOverlay, fontSize)
		} else {
			inputs = append(inputs, "-i", media.Path)
			videoFilter = processVideoFilter(index, fadeSec)
		}
		segmentDurations = append(segmentDurations, media.SegmentDuration)
		filterComplex += fmt.Sprintf("%s[v%d]; ", videoFilter, index)
	}

	// Add crossfade transitions
	filterComplex += buildCrossfadeFilters(segmentDurations, fadeSec)

	// Add final filters and get duration
	finalFilters, finalLength := buildFinalFilters(segmentDurations, fadeSec)
	filterComplex += finalFilters

	// Setup audio processing
	totalDuration := finalLength
	audioConfig := setupAudioProcessing(inputs, mediaInputs, totalDuration, fadeSec, musicFiles, keepVideoAudio)

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
		audioBitrateSource := audioConfig.AudioBitrateSource
		if audioBitrateSource == "" && len(musicFiles) > 0 {
			audioBitrateSource = musicFiles[0]
		}
		audioBitrate := "192k"
		if audioBitrateSource != "" {
			audioBitrate = getAudioBitrateStr(audioBitrateSource)
		}
		args = append(args, "-c:a", "aac", "-b:a", audioBitrate)
	}

	args = append(args, "-t", formatSeconds(finalLength))
	args = append(args, outputFilename)

	// Execute FFmpeg command
	if err := runFFmpegCommand(args, audioConfig.HasAudio); err != nil {
		log.Fatalf("Video generation failed: %v", err)
	}

	// Clean up concat file if it was created
	if len(musicFiles) > 1 {
		os.Remove("audio_concat.txt")
	}

	// Display final information
	displayVideoInfo(outputFilename, finalLength)
}

func normalizeKenBurnsMode(mode string) string {
	mode = strings.TrimSpace(strings.ToLower(mode))
	switch mode {
	case kenBurnsModeLow:
		return kenBurnsModeLow
	case kenBurnsModeMedium:
		return kenBurnsModeMedium
	case kenBurnsModeHigh:
		return kenBurnsModeHigh
	// Backward-compatibility aliases from previous releases.
	case "subtle":
		return kenBurnsModeLow
	case "cinematic":
		return kenBurnsModeMedium
	case "dynamic":
		return kenBurnsModeHigh
	default:
		return kenBurnsModeHigh
	}
}

// getKenBurnsEffect generates a Ken Burns effect using a fixed zoompan expression.
// This approach is based on the method described in the Bannerbear blog.
// Updated with softer effects: slower zoom speed, lower max zoom, and reduced movement
func getKenBurnsEffect(duration float64) string {
	totalFrames := int(math.Round(duration * float64(activeFPS)))
	if totalFrames < 1 {
		totalFrames = 1
	}

	mode := normalizeKenBurnsMode(activeKenBurnsMode)

	startZoom := 1.00
	endZoom := 1.07
	offsetX := "126"
	offsetY := "70"

	if mode == kenBurnsModeLow {
		endZoom = 1.04
		offsetX = "98"
		offsetY = "56"
	} else if mode == kenBurnsModeHigh {
		endZoom = 1.10
		offsetX = "154"
		offsetY = "84"
	}

	if activeResolution == resolutionFullHD {
		if mode == kenBurnsModeLow {
			endZoom = 1.03
			offsetX = "70"
			offsetY = "40"
		} else if mode == kenBurnsModeMedium {
			endZoom = 1.05
			offsetX = "84"
			offsetY = "48"
		} else {
			endZoom = 1.07
			offsetX = "98"
			offsetY = "56"
		}
	}

	// zoompan runs at the supersampled canvas; processImageFilter then scales back to activeResolution.
	superRes := supersampledResolution()
	zoomStep := (endZoom - startZoom) / float64(totalFrames)
	if zoomStep < 0.0001 {
		zoomStep = 0.0001
	}
	zoomStepStr := strconv.FormatFloat(zoomStep, 'f', 6, 64)
	endZoomStr := strconv.FormatFloat(endZoom, 'f', 2, 64)

	buildExpr := func(xOffset, yOffset string) string {
		return fmt.Sprintf(
			"zoompan=z='min(zoom+%s,%s)':x='iw/2-(iw/zoom/2)%s':y='ih/2-(ih/zoom/2)%s':d=%d:s=%s",
			zoomStepStr,
			endZoomStr,
			xOffset,
			yOffset,
			totalFrames,
			superRes,
		)
	}

	// Pan+zoom variants across all intensity levels (low/medium/high).
	variants := []string{
		buildExpr("+"+offsetX, ""),
		buildExpr("-"+offsetX, ""),
		buildExpr("", "+"+offsetY),
		buildExpr("", "-"+offsetY),
		buildExpr("+"+offsetX, "+"+offsetY),
		buildExpr("-"+offsetX, "+"+offsetY),
		buildExpr("+"+offsetX, "-"+offsetY),
		buildExpr("-"+offsetX, "-"+offsetY),
	}

	return variants[rand.Intn(len(variants))]
}
