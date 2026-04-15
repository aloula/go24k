package utils

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// isWSL detects if we're running in Windows Subsystem for Linux.
func isWSL() bool {
	if runtime.GOOS != linuxOS {
		return false
	}

	if data, err := os.ReadFile("/proc/version"); err == nil {
		version := strings.ToLower(string(data))
		return strings.Contains(version, "microsoft") || strings.Contains(version, "wsl")
	}

	return os.Getenv("WSL_DISTRO_NAME") != ""
}

func checkNVENCAvailable() bool {
	cmd := newExecCommand("ffmpeg", "-encoders")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	if !strings.Contains(string(output), "h264_nvenc") {
		return false
	}

	testCmd := newExecCommand("ffmpeg", "-f", "lavfi", "-i", "testsrc=duration=0.1:size=320x240:rate=1",
		"-c:v", "h264_nvenc", "-f", "null", "-")
	return testCmd.Run() == nil
}

func checkQSVAvailable() bool {
	cmd := newExecCommand("ffmpeg", "-encoders")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	if !strings.Contains(string(output), "h264_qsv") {
		return false
	}

	testCmd := newExecCommand("ffmpeg", "-f", "lavfi", "-i", "testsrc=duration=0.1:size=320x240:rate=1",
		"-c:v", "h264_qsv", "-f", "null", "-")
	return testCmd.Run() == nil
}

func checkAMFAvailable() bool {
	cmd := newExecCommand("ffmpeg", "-encoders")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	if !strings.Contains(string(output), "h264_amf") {
		return false
	}

	testCmd := newExecCommand("ffmpeg", "-f", "lavfi", "-i", "testsrc=duration=0.1:size=320x240:rate=1",
		"-c:v", "h264_amf", "-f", "null", "-")
	return testCmd.Run() == nil
}

func checkMediaFoundationAvailable() bool {
	cmd := newExecCommand("ffmpeg", "-encoders")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	if !strings.Contains(string(output), "h264_mf") {
		return false
	}

	testCmd := newExecCommand("ffmpeg", "-f", "lavfi", "-i", "testsrc=duration=0.1:size=320x240:rate=1",
		"-c:v", "h264_mf", "-f", "null", "-")
	return testCmd.Run() == nil
}

func checkVAAPIAvailable() bool {
	cmd := newExecCommand("ffmpeg", "-encoders")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	if !strings.Contains(string(output), "h264_vaapi") {
		return false
	}

	testCmd := newExecCommand("ffmpeg", "-f", "lavfi", "-i", "testsrc=duration=0.1:size=320x240:rate=1",
		"-c:v", "h264_vaapi", "-f", "null", "-")
	return testCmd.Run() == nil
}

func checkVideoToolboxAvailable() bool {
	cmd := newExecCommand("ffmpeg", "-encoders")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "h264_videotoolbox")
}

// HardwareEncoder represents different hardware encoding options.
type HardwareEncoder struct {
	Name        string
	Codec       string
	Description string
	Platform    string
}

// getOptimalVideoSettings returns optimized FFmpeg settings based on environment and hardware.
func getOptimalVideoSettings() []string {
	hasNVENC := checkNVENCAvailable()
	hasVideoToolbox := checkVideoToolboxAvailable()
	hasQSV := checkQSVAvailable()
	hasAMF := checkAMFAvailable()
	hasMediaFoundation := checkMediaFoundationAvailable()
	hasVAAPI := checkVAAPIAvailable()

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

	if hasNVENC {
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
		fmt.Printf("Hardware: VideoToolbox detected - using Apple hardware acceleration\n")
		settings = append(settings,
			"-c:v", "h264_videotoolbox",
			"-profile:v", "high",
			"-level", h264Level,
			"-q:v", "21",
			"-realtime", "false",
			"-frames:v", "0",
			"-b:v", "10M",
			"-maxrate", "15M",
			"-bufsize", "30M",
		)
	} else if hasMediaFoundation {
		fmt.Printf("Hardware: Media Foundation detected - using Windows hardware acceleration\n")
		settings = append(settings,
			"-c:v", "h264_mf",
			"-quality", "quality",
			"-rate_control", "quality",
			"-scenario", "display_remoting",
			"-profile:v", "high",
			"-level", h264Level,
			"-b:v", "12M",
			"-maxrate", "18M",
			"-bufsize", "36M",
		)
	} else if hasQSV {
		fmt.Printf("Hardware: Intel QSV detected - using Intel hardware acceleration\n")
		settings = append(settings,
			"-c:v", "h264_qsv",
			"-preset", "slower",
			"-profile:v", "high",
			"-level", h264Level,
			"-global_quality", "21",
			"-look_ahead", "1",
			"-maxrate", "12M",
			"-bufsize", "24M",
		)
	} else if hasAMF {
		fmt.Printf("Hardware: AMD AMF detected - using AMD hardware acceleration\n")
		settings = append(settings,
			"-c:v", "h264_amf",
			"-quality", "quality",
			"-rc", "cqp",
			"-qp_i", "21", "-qp_p", "21", "-qp_b", "21",
			"-profile:v", "high",
			"-level", h264Level,
			"-maxrate", "12M",
			"-bufsize", "24M",
		)
	} else if hasVAAPI {
		fmt.Printf("Hardware: VAAPI detected - using Linux hardware acceleration\n")
		settings = append(settings,
			"-c:v", "h264_vaapi",
			"-profile:v", "high",
			"-level", h264Level,
			"-crf", "21",
			"-maxrate", "10M",
			"-bufsize", "20M",
		)
	} else {
		fmt.Printf("CPU: Using libx264 software encoding\n")
		settings = append(settings,
			"-c:v", "libx264",
			"-preset", "slow",
			"-profile:v", "high",
			"-level", h264Level,
			"-crf", "21",
		)
	}

	return settings
}

// ShowEnvironmentInfo displays environment detection and optimization details.
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

	hasNVENC := checkNVENCAvailable()
	hasVideoToolbox := checkVideoToolboxAvailable()
	hasQSV := checkQSVAvailable()
	hasAMF := checkAMFAvailable()
	hasMediaFoundation := checkMediaFoundationAvailable()
	hasVAAPI := checkVAAPIAvailable()

	fmt.Printf("\nHardware Acceleration Detection:\n")
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

	settings := getOptimalVideoSettings()
	fmt.Printf("\nOptimized FFmpeg Settings:\n")
	for i := 0; i < len(settings); i += 2 {
		if i+1 < len(settings) {
			fmt.Printf("  %s: %s\n", settings[i], settings[i+1])
		}
	}

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
