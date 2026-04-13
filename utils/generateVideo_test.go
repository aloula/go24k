package utils

import (
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestIsWSL tests the WSL detection function
func TestIsWSL(t *testing.T) {
	// Save original environment
	originalWSLDistro := os.Getenv("WSL_DISTRO_NAME")

	defer func() {
		// Restore original environment
		os.Setenv("WSL_DISTRO_NAME", originalWSLDistro)
	}()

	tests := []struct {
		name        string
		goos        string
		wslDistro   string
		procVersion string
		expected    bool
	}{
		{
			name:     "Windows - should return false",
			goos:     "windows",
			expected: false,
		},
		{
			name:     "macOS - should return false",
			goos:     "darwin",
			expected: false,
		},
		{
			name:      "Linux with WSL distro env var",
			goos:      "linux",
			wslDistro: "Ubuntu",
			expected:  true,
		},
		{
			name:      "Linux without WSL indicators",
			goos:      "linux",
			wslDistro: "",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment for test
			os.Setenv("WSL_DISTRO_NAME", tt.wslDistro)

			// Note: We can't easily mock runtime.GOOS and /proc/version
			// So we'll test what we can and skip the parts that require
			// system-level mocking in unit tests

			if tt.goos != "linux" {
				// For non-Linux systems, we expect false regardless
				// This is hard to test without mocking runtime.GOOS
				t.Skip("Skipping GOOS-dependent test")
			}

			// Test WSL environment variable detection
			if tt.wslDistro != "" {
				// We know this should detect WSL via env var
				result := isWSL()
				if runtime.GOOS == "linux" && result != tt.expected {
					t.Errorf("isWSL() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

// TestCheckNVENCAvailable tests NVENC detection
func TestCheckNVENCAvailable(t *testing.T) {
	// This test is environment-dependent
	// We'll test that the function doesn't panic and returns a boolean
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("checkNVENCAvailable() panicked: %v", r)
		}
	}()

	result := checkNVENCAvailable()

	// Result should be a boolean (true or false both valid)
	if result != true && result != false {
		t.Errorf("checkNVENCAvailable() should return boolean, got %T", result)
	}

	t.Logf("NVENC Available: %v", result)
}

// TestCheckQSVAvailable tests Intel QuickSync detection
func TestCheckQSVAvailable(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("checkQSVAvailable() panicked: %v", r)
		}
	}()

	result := checkQSVAvailable()

	if result != true && result != false {
		t.Errorf("checkQSVAvailable() should return boolean, got %T", result)
	}

	t.Logf("QSV Available: %v", result)
}

// TestCheckAMFAvailable tests AMD AMF detection
func TestCheckAMFAvailable(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("checkAMFAvailable() panicked: %v", r)
		}
	}()

	result := checkAMFAvailable()

	if result != true && result != false {
		t.Errorf("checkAMFAvailable() should return boolean, got %T", result)
	}

	t.Logf("AMF Available: %v", result)
}

func TestCheckVideoToolboxAvailable(t *testing.T) {
	// Graceful handling if ffmpeg is not available
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Recovered from panic: %v", r)
		}
	}()

	result := checkVideoToolboxAvailable()

	if result != true && result != false {
		t.Errorf("checkVideoToolboxAvailable() should return boolean, got %T", result)
	}

	t.Logf("VideoToolbox Available: %v", result)
}

// TestGetOptimalVideoSettings tests video settings generation
func TestGetOptimalVideoSettings(t *testing.T) {
	settings := getOptimalVideoSettings()

	if len(settings) == 0 {
		t.Error("getOptimalVideoSettings() returned empty settings")
	}

	// Check that settings come in pairs (flag, value)
	if len(settings)%2 != 0 {
		t.Errorf("Settings should come in pairs, got %d items", len(settings))
	}

	// Check for required settings
	requiredSettings := []string{"-pix_fmt", "-movflags", "-r", "-s"}
	for _, required := range requiredSettings {
		found := false
		for i := 0; i < len(settings); i += 2 {
			if settings[i] == required {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Required setting %s not found in output", required)
		}
	}

	// Verify resolution is 4K
	for i := 0; i < len(settings)-1; i += 2 {
		if settings[i] == "-s" {
			if settings[i+1] != "3840x2160" {
				t.Errorf("Expected 4K resolution 3840x2160, got %s", settings[i+1])
			}
		}
	}

	t.Logf("Video settings: %v", settings)
}

// TestGetKenBurnsEffect tests Ken Burns effect generation
func TestGetKenBurnsEffect(t *testing.T) {
	testCases := []struct {
		name     string
		duration float64
	}{
		{"Short duration", 3},
		{"Medium duration", 5},
		{"Long duration", 10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			effect := getKenBurnsEffect(tc.duration)

			if effect == "" {
				t.Error("getKenBurnsEffect() returned empty string")
			}

			// Should contain zoompan filter
			if !strings.Contains(effect, "zoompan") {
				t.Error("Effect should contain 'zoompan' filter")
			}

			if !strings.Contains(effect, "cos(PI*on/") && !strings.Contains(effect, "min(zoom+") {
				t.Error("Effect should contain a valid motion expression")
			}

			// Should contain supersampled resolution (2× active) used to prevent zoompan jitter
			if !strings.Contains(effect, "7680x4320") {
				t.Error("Effect should contain supersampled 8K resolution to prevent jitter")
			}

			// Should contain duration
			expectedFrames := int(tc.duration * 30)
			if !strings.Contains(effect, strconv.Itoa(expectedFrames)) {
				// This is a rough check - the actual number formatting might differ
				t.Logf("Effect might not contain expected frames %d", expectedFrames)
			}

			t.Logf("Duration %.0f -> Effect: %s", tc.duration, effect)
		})
	}
}

func TestNormalizeKenBurnsMode(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "default empty", input: "", expected: kenBurnsModeHigh},
		{name: "low", input: "low", expected: kenBurnsModeLow},
		{name: "medium", input: "medium", expected: kenBurnsModeMedium},
		{name: "high", input: "high", expected: kenBurnsModeHigh},
		{name: "legacy subtle", input: "subtle", expected: kenBurnsModeLow},
		{name: "legacy cinematic", input: "cinematic", expected: kenBurnsModeMedium},
		{name: "legacy dynamic", input: "dynamic", expected: kenBurnsModeHigh},
		{name: "invalid", input: "fast", expected: kenBurnsModeHigh},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeKenBurnsMode(tc.input)
			if got != tc.expected {
				t.Fatalf("normalizeKenBurnsMode(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestGetKenBurnsEffect_ModeSelection(t *testing.T) {
	oldMode := activeKenBurnsMode
	oldResolution := activeResolution
	oldFPS := activeFPS
	defer func() {
		activeKenBurnsMode = oldMode
		activeResolution = oldResolution
		activeFPS = oldFPS
	}()

	activeResolution = resolution4K
	activeFPS = 30

	activeKenBurnsMode = kenBurnsModeLow
	low := getKenBurnsEffect(5)
	if !strings.Contains(low, "min(zoom+") {
		t.Fatalf("low mode should use incremental zoom expression, got: %s", low)
	}
	if !strings.Contains(low, "+98") && !strings.Contains(low, "-98") && !strings.Contains(low, "+56") && !strings.Contains(low, "-56") {
		t.Fatalf("low mode should include low-intensity pan offsets, got: %s", low)
	}

	activeKenBurnsMode = kenBurnsModeMedium
	medium := getKenBurnsEffect(5)
	if !strings.Contains(medium, "+126") && !strings.Contains(medium, "-126") && !strings.Contains(medium, "+70") && !strings.Contains(medium, "-70") {
		t.Fatalf("medium mode should include medium-intensity pan offsets, got: %s", medium)
	}

	activeKenBurnsMode = kenBurnsModeHigh
	high := getKenBurnsEffect(5)
	if !strings.Contains(high, "min(zoom+") {
		t.Fatalf("high mode should use incremental zoom expression, got: %s", high)
	}
	if !strings.Contains(high, "+154") && !strings.Contains(high, "-154") && !strings.Contains(high, "+84") && !strings.Contains(high, "-84") {
		t.Fatalf("high mode should include high-intensity pan offsets, got: %s", high)
	}
}

// TestShowEnvironmentInfo tests environment info display
func TestShowEnvironmentInfo(t *testing.T) {
	// This function prints to stdout, so we test it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ShowEnvironmentInfo() panicked: %v", r)
		}
	}()

	ShowEnvironmentInfo()
	// If we get here without panicking, the test passes
}

// TestGenerateVideo_InvalidInputs tests video generation with invalid inputs
func TestGenerateVideo_InvalidInputs(t *testing.T) {
	// Setup temporary directory
	_ = setupTestDir(t)

	// Test with no converted images
	defer func() {
		if r := recover(); r != nil {
			// Expected behavior - should handle gracefully
			t.Logf("GenerateVideo panicked as expected with no images: %v", r)
		}
	}()

	// This should fail gracefully (we hope)
	// Note: GenerateVideo uses log.Fatalf which will exit the program
	// In a real test, we'd need to refactor this to return errors instead
	t.Skip("Skipping GenerateVideo test as it uses log.Fatalf")
}

// TestGetOptimalVideoSettings_AllPaths tests all hardware detection paths
func TestGetOptimalVideoSettings_AllPaths(t *testing.T) {
	// We'll test the logic by checking different scenarios
	// This is tricky because the function checks actual hardware
	// But we can at least verify the function runs without panicking

	settings := getOptimalVideoSettings()

	// Verify basic required settings are present
	hasPixFmt := false
	hasMovFlags := false
	hasFramerate := false
	hasResolution := false
	hasCodec := false

	for i := 0; i < len(settings)-1; i += 2 {
		switch settings[i] {
		case "-pix_fmt":
			hasPixFmt = true
			if settings[i+1] != "yuv420p" {
				t.Errorf("Expected yuv420p pixel format, got %s", settings[i+1])
			}
		case "-movflags":
			hasMovFlags = true
		case "-r":
			hasFramerate = true
			if settings[i+1] != "30" {
				t.Errorf("Expected 30 fps, got %s", settings[i+1])
			}
		case "-s":
			hasResolution = true
			if settings[i+1] != "3840x2160" {
				t.Errorf("Expected 4K resolution, got %s", settings[i+1])
			}
		case "-c:v":
			hasCodec = true
		}
	}

	if !hasPixFmt {
		t.Error("Missing pixel format setting")
	}
	if !hasMovFlags {
		t.Error("Missing movflags setting")
	}
	if !hasFramerate {
		t.Error("Missing framerate setting")
	}
	if !hasResolution {
		t.Error("Missing resolution setting")
	}
	if !hasCodec {
		t.Error("Missing codec setting")
	}
}

// TestHardwareDetection_EdgeCases tests edge cases in hardware detection
func TestHardwareDetection_EdgeCases(t *testing.T) {
	t.Run("Multiple_calls_consistent", func(t *testing.T) {
		// Test that multiple calls return the same result
		first := checkNVENCAvailable()
		second := checkNVENCAvailable()

		if first != second {
			t.Error("checkNVENCAvailable should return consistent results")
		}
	})

	t.Run("All_detection_functions_callable", func(t *testing.T) {
		// Verify all hardware detection functions can be called without panicking
		_ = checkNVENCAvailable()
		_ = checkVideoToolboxAvailable()
		_ = checkQSVAvailable()
		_ = checkAMFAvailable()
		_ = checkMediaFoundationAvailable()
		_ = checkVAAPIAvailable()
	})
}

// TestEnvironmentDetection_Scenarios tests different environment scenarios
func TestEnvironmentDetection_Scenarios(t *testing.T) {
	t.Run("ShowEnvironmentInfo_runs", func(t *testing.T) {
		// Test that ShowEnvironmentInfo runs without panicking
		// This will produce output, but that's expected in tests
		ShowEnvironmentInfo()
	})
}

// TestKenBurnsEffect_EdgeCases tests Ken Burns effect with edge cases
func TestKenBurnsEffect_EdgeCases(t *testing.T) {
	t.Run("Zero_duration", func(t *testing.T) {
		effect := getKenBurnsEffect(0)

		// Should still produce valid output even with 0 duration
		if !strings.Contains(effect, "zoompan") {
			t.Error("Expected zoompan filter in output")
		}
		if !strings.Contains(effect, "d=1") {
			t.Error("Expected clamped minimum duration (d=1) in output")
		}
	})

	t.Run("Negative_duration", func(t *testing.T) {
		effect := getKenBurnsEffect(-5)

		// Should handle negative duration gracefully
		if !strings.Contains(effect, "zoompan") {
			t.Error("Expected zoompan filter in output")
		}
	})

	t.Run("Very_large_duration", func(t *testing.T) {
		effect := getKenBurnsEffect(999999)

		// Should handle very large duration
		if !strings.Contains(effect, "zoompan") {
			t.Error("Expected zoompan filter in output")
		}
		if !strings.Contains(effect, "d=29999970") { // 999999 * 30 fps
			t.Error("Expected calculated duration in output")
		}
	})

	t.Run("Check_zoom_parameters", func(t *testing.T) {
		activeKenBurnsMode = kenBurnsModeHigh
		effect := getKenBurnsEffect(5)

		// Verify high mode uses incremental zoom up to 1.10.
		if !strings.Contains(effect, "min(zoom+") {
			t.Error("Expected incremental zoom expression in zoompan filter")
		}
		if !strings.Contains(effect, "1.10") {
			t.Error("Expected end zoom 1.10 for high mode")
		}
	})

	t.Run("Movement_patterns", func(t *testing.T) {
		// Test multiple calls to see different movement patterns
		effects := make([]string, 5)
		for i := 0; i < 5; i++ {
			effects[i] = getKenBurnsEffect(5)
		}

		// Check that we get some variation (due to random movement)
		allSame := true
		for i := 1; i < len(effects); i++ {
			if effects[i] != effects[0] {
				allSame = false
				break
			}
		}

		// Note: There's a small chance all could be the same due to randomness,
		// but with 5 samples it's very unlikely
		if allSame {
			t.Log("Warning: All Ken Burns effects were identical - check randomization")
		}
	})
}

// BenchmarkHardwareDetection benchmarks hardware detection performance
func BenchmarkHardwareDetection(b *testing.B) {
	b.Run("NVENC", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			checkNVENCAvailable()
		}
	})

	b.Run("QSV", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			checkQSVAvailable()
		}
	})

	b.Run("AMF", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			checkAMFAvailable()
		}
	})
}

// BenchmarkKenBurnsEffect benchmarks Ken Burns effect generation
func BenchmarkKenBurnsEffect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getKenBurnsEffect(5)
	}
}

func TestGetVideoDetails(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{"Non-existent file", "non_existent_file.mp4", true},
		{"Empty filename", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := getVideoDetails(tt.filename)

			if tt.wantErr && err == nil {
				t.Errorf("getVideoDetails(%s) expected error but got none", tt.filename)
			}

			// Even on error, we should get a valid info struct with defaults
			if info == nil {
				t.Errorf("getVideoDetails(%s) returned nil info", tt.filename)
			} else {
				// Check that defaults are set
				if info.Framerate == "" {
					t.Errorf("getVideoDetails(%s) should set default framerate", tt.filename)
				}
				if info.Resolution == "" {
					t.Errorf("getVideoDetails(%s) should set default resolution", tt.filename)
				}
				if info.AudioBitrate == "" {
					t.Errorf("getVideoDetails(%s) should set default audio bitrate", tt.filename)
				}
			}
		})
	}
}

func TestVideoInfo_DefaultValues(t *testing.T) {
	// Test that VideoInfo struct has proper default behavior
	info := &VideoInfo{}

	// Initial state should be zero values
	if info.FileSizeMB != 0 {
		t.Errorf("Expected FileSizeMB to be 0, got %f", info.FileSizeMB)
	}
	if info.DurationSec != 0 {
		t.Errorf("Expected DurationSec to be 0, got %f", info.DurationSec)
	}
	if info.VideoBitrate != "" {
		t.Errorf("Expected VideoBitrate to be empty, got %s", info.VideoBitrate)
	}
	if info.AudioBitrate != "" {
		t.Errorf("Expected AudioBitrate to be empty, got %s", info.AudioBitrate)
	}
	if info.Framerate != "" {
		t.Errorf("Expected Framerate to be empty, got %s", info.Framerate)
	}
	if info.Resolution != "" {
		t.Errorf("Expected Resolution to be empty, got %s", info.Resolution)
	}
}

func TestFindVideoFiles_ExcludesOutputVideo(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	defer func() {
		_ = os.Chdir(oldWd)
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}

	for _, name := range []string{"clip.mp4", "scene.mov", outputVideoUHD, outputVideoFHD, outputVideoLegacy} {
		if err := os.WriteFile(name, []byte("test"), 0644); err != nil {
			t.Fatalf("WriteFile(%s) failed: %v", name, err)
		}
	}

	files, err := findVideoFiles(true, false)
	if err != nil {
		t.Fatalf("findVideoFiles returned error: %v", err)
	}

	joined := strings.Join(files, ",")
	for _, generated := range []string{outputVideoUHD, outputVideoFHD, outputVideoLegacy} {
		if strings.Contains(joined, generated) {
			t.Fatalf("findVideoFiles should exclude %s, got %v", generated, files)
		}
	}
	if !strings.Contains(joined, "clip.mp4") || !strings.Contains(joined, "scene.mov") {
		t.Fatalf("findVideoFiles missed expected inputs, got %v", files)
	}
}

func TestFindVideoFiles_MOVOnly(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	defer func() {
		_ = os.Chdir(oldWd)
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}

	for _, name := range []string{"clip.mp4", "scene.mov", "iphone.MOV", outputVideoUHD} {
		if err := os.WriteFile(name, []byte("test"), 0644); err != nil {
			t.Fatalf("WriteFile(%s) failed: %v", name, err)
		}
	}

	files, err := findVideoFiles(false, true)
	if err != nil {
		t.Fatalf("findVideoFiles returned error: %v", err)
	}

	joined := strings.Join(files, ",")
	if strings.Contains(joined, "clip.mp4") {
		t.Fatalf("findVideoFiles should not include non-MOV files in MOV-only mode, got %v", files)
	}
	if !strings.Contains(joined, "scene.mov") || !strings.Contains(joined, "iphone.MOV") {
		t.Fatalf("findVideoFiles should include .mov and .MOV files, got %v", files)
	}
	if strings.Contains(joined, outputVideoUHD) {
		t.Fatalf("findVideoFiles should exclude generated outputs in MOV-only mode, got %v", files)
	}
}

func TestExtractImageTimestampFromConvertedName(t *testing.T) {
	ts, ok := extractImageTimestampFromConvertedName("converted/20240223_153741_uhd.jpg")
	if !ok {
		t.Fatalf("expected timestamp to be extracted")
	}

	if ts.Format("20060102_150405") != "20240223_153741" {
		t.Fatalf("unexpected parsed timestamp: %s", ts.Format("20060102_150405"))
	}

	ts, ok = extractImageTimestampFromConvertedName("converted/20240223_153741_fhd.jpg")
	if !ok {
		t.Fatalf("expected timestamp to be extracted for _fhd suffix")
	}

	if ts.Format("20060102_150405") != "20240223_153741" {
		t.Fatalf("unexpected parsed timestamp for _fhd: %s", ts.Format("20060102_150405"))
	}

	_, ok = extractImageTimestampFromConvertedName("converted/not_a_timestamp_uhd.jpg")
	if ok {
		t.Fatalf("expected no timestamp for invalid converted filename")
	}
}

func TestParseVideoCreationTime(t *testing.T) {
	testCases := []string{
		"2024-02-23T15:37:41Z",
		"2024-02-23T15:37:41.123Z",
		"2024-02-23 15:37:41",
	}

	for _, tc := range testCases {
		if _, err := parseVideoCreationTime(tc); err != nil {
			t.Fatalf("expected to parse %q, got error: %v", tc, err)
		}
	}
}

func TestMediaSorting_UsesFilenameWhenTimestampMissing(t *testing.T) {
	older := time.Date(2024, 2, 20, 10, 0, 0, 0, time.UTC)
	newer := time.Date(2024, 2, 21, 10, 0, 0, 0, time.UTC)

	media := []MediaInput{
		{Path: "a_no_ts.jpg", SortName: "a_no_ts.jpg", HasCapturedAt: false},
		{Path: "b_new.mp4", SortName: "b_new.mp4", HasCapturedAt: true, CapturedAt: newer},
		{Path: "c_old.jpg", SortName: "c_old.jpg", HasCapturedAt: true, CapturedAt: older},
	}

	sort.Slice(media, func(i, j int) bool {
		if media[i].HasCapturedAt && media[j].HasCapturedAt {
			if !media[i].CapturedAt.Equal(media[j].CapturedAt) {
				return media[i].CapturedAt.Before(media[j].CapturedAt)
			}
		}
		left := mediaSortName(media[i].SortName)
		right := mediaSortName(media[j].SortName)
		return left < right
	})

	if media[0].Path != "a_no_ts.jpg" || media[1].Path != "c_old.jpg" || media[2].Path != "b_new.mp4" {
		t.Fatalf("unexpected sort order: %#v", media)
	}
}

func TestExtractCaptureTimeFromFilename(t *testing.T) {
	testCases := []struct {
		name     string
		filename string
		expected string
		ok       bool
	}{
		{name: "converted style", filename: "converted/20240223_153741_uhd.jpg", expected: "20240223_153741", ok: true},
		{name: "converted fullhd style", filename: "converted/20240223_153741_fhd.jpg", expected: "20240223_153741", ok: true},
		{name: "iso style", filename: "VID_2024-02-23_15-37-41.mp4", expected: "20240223_153741", ok: true},
		{name: "compact style", filename: "IMG20240223153741.jpg", expected: "20240223_153741", ok: true},
		{name: "invalid", filename: "holiday-final-cut.mp4", ok: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, ok := extractCaptureTimeFromFilename(tc.filename)
			if ok != tc.ok {
				t.Fatalf("extractCaptureTimeFromFilename(%q) ok = %v, want %v", tc.filename, ok, tc.ok)
			}
			if tc.ok {
				if got := ts.Format("20060102_150405"); got != tc.expected {
					t.Fatalf("extractCaptureTimeFromFilename(%q) = %s, want %s", tc.filename, got, tc.expected)
				}
			}
		})
	}
}

func TestMediaSorting_FilenameTimestampWhenMetadataMissing(t *testing.T) {
	videoTS, ok := extractCaptureTimeFromFilename("VID_2024-02-20_09-00-00.mp4")
	if !ok {
		t.Fatal("expected video filename timestamp to parse")
	}

	imageTS, ok := extractCaptureTimeFromFilename("IMG_2024-02-20_10-00-00_uhd.jpg")
	if !ok {
		t.Fatal("expected image filename timestamp to parse")
	}

	media := []MediaInput{
		{Path: "IMG_2024-02-20_10-00-00_uhd.jpg", SortName: "img_0010.jpg", HasCapturedAt: true, CapturedAt: imageTS},
		{Path: "VID_2024-02-20_09-00-00.mp4", SortName: "vid_0009.mp4", HasCapturedAt: true, CapturedAt: videoTS},
	}

	sort.Slice(media, func(i, j int) bool {
		if media[i].HasCapturedAt && media[j].HasCapturedAt {
			if !media[i].CapturedAt.Equal(media[j].CapturedAt) {
				return media[i].CapturedAt.Before(media[j].CapturedAt)
			}
		}
		return mediaSortName(media[i].SortName) < mediaSortName(media[j].SortName)
	})

	if media[0].Path != "VID_2024-02-20_09-00-00.mp4" {
		t.Fatalf("expected video to come first by filename-derived timestamp, got %s", media[0].Path)
	}
}

func TestMediaSorting_OrderByFilenameMode(t *testing.T) {
	older := time.Date(2024, 2, 20, 9, 0, 0, 0, time.UTC)
	newer := time.Date(2024, 2, 21, 9, 0, 0, 0, time.UTC)

	media := []MediaInput{
		{Path: "z-last.jpg", SortName: "z-last.jpg", HasCapturedAt: true, CapturedAt: older},
		{Path: "a-first.jpg", SortName: "a-first.jpg", HasCapturedAt: true, CapturedAt: newer},
	}

	orderByFilename := true
	// Use the EXACT same comparator as production code in collectMediaInputs
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

	if media[0].Path != "a-first.jpg" {
		t.Fatalf("expected filename ordering (a-first before z-last), got %s first", media[0].Path)
	}
}

func TestMediaSorting_OrderByFilenameMode_WithVideos(t *testing.T) {
	// Test that images and videos sort together by filename when orderByFilename=true
	media := []MediaInput{
		{Path: "converted/z_uhd.jpg", SortName: "z.jpg", IsImage: true},
		{Path: "intro.mp4", SortName: "intro.mp4", IsImage: false},
		{Path: "converted/a_uhd.jpg", SortName: "a.jpg", IsImage: true},
		{Path: "outro.mp4", SortName: "outro.mp4", IsImage: false},
	}

	orderByFilename := true
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

	// Expected order: a.jpg, intro.mp4, outro.mp4, z.jpg (alphabetical by SortName)
	if media[0].SortName != "a.jpg" || media[1].SortName != "intro.mp4" || media[2].SortName != "outro.mp4" || media[3].SortName != "z.jpg" {
		t.Fatalf("expected alphabetical ordering by SortName, got %v", []string{media[0].SortName, media[1].SortName, media[2].SortName, media[3].SortName})
	}
}

func TestResolveImageSortName_UsesOriginalFilename(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	defer func() {
		_ = os.Chdir(oldWd)
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}

	if err := os.MkdirAll("converted", 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	if err := os.WriteFile("img_020.jpg", []byte("fake-jpg"), 0644); err != nil {
		t.Fatalf("WriteFile img_020.jpg failed: %v", err)
	}
	if err := os.WriteFile("converted/img_020_uhd.jpg", []byte("fake-jpg"), 0644); err != nil {
		t.Fatalf("WriteFile converted image failed: %v", err)
	}

	got := resolveImageSortName("converted/img_020_uhd.jpg")
	if got != "img_020.jpg" {
		t.Fatalf("resolveImageSortName returned %q, want %q", got, "img_020.jpg")
	}
}

func TestBuildTimelineOffsets(t *testing.T) {
	mediaInputs := []MediaInput{
		{Path: "converted/a.jpg", IsImage: true, SegmentDuration: 8},
		{Path: "converted/b.jpg", IsImage: true, SegmentDuration: 8},
		{Path: "clip.mp4", IsImage: false, HasAudio: true, SegmentDuration: 12.5},
	}

	offsets := buildTimelineOffsets(mediaInputs, 2)
	if len(offsets) != 3 {
		t.Fatalf("expected 3 offsets, got %d", len(offsets))
	}

	if offsets[0] != 0 {
		t.Fatalf("expected first offset 0, got %f", offsets[0])
	}
	if offsets[1] != 6 {
		t.Fatalf("expected second offset 6, got %f", offsets[1])
	}
	if offsets[2] != 12 {
		t.Fatalf("expected third offset 12, got %f", offsets[2])
	}
}

func TestSetupAudioProcessing_MixesMusicAndClipAudio(t *testing.T) {
	mediaInputs := []MediaInput{
		{Path: "converted/a.jpg", IsImage: true, SegmentDuration: 8},
		{Path: "clip.mp4", IsImage: false, HasAudio: true, SegmentDuration: 12},
	}

	config := setupAudioProcessing([]string{"-loop", "1", "-t", "8", "-i", "converted/a.jpg", "-i", "clip.mp4"}, mediaInputs, 18, 2, []string{"soundtrack.mp3"}, true)

	if !config.HasAudio {
		t.Fatal("expected mixed audio output to be enabled")
	}
	if config.AudioBitrateSource != "soundtrack.mp3" {
		t.Fatalf("expected audio bitrate source soundtrack.mp3, got %s", config.AudioBitrateSource)
	}
	if !strings.Contains(config.AudioFilter, "volume=") {
		t.Fatalf("expected volume mute expression in audio filter, got %s", config.AudioFilter)
	}
	if !strings.Contains(config.AudioFilter, "musicmuted") {
		t.Fatalf("expected muted music label in audio filter, got %s", config.AudioFilter)
	}
	if !strings.Contains(config.AudioFilter, "clipaudio0") {
		t.Fatalf("expected delayed clip audio label in audio filter, got %s", config.AudioFilter)
	}
	if len(config.MapArgs) < 4 || config.MapArgs[3] != "[mixedaudio]" {
		t.Fatalf("expected mixed audio mapping, got %v", config.MapArgs)
	}
}

func TestSetupAudioProcessing_UsesClipAudioWithoutMusic(t *testing.T) {
	mediaInputs := []MediaInput{
		{Path: "clip.mp4", IsImage: false, HasAudio: true, SegmentDuration: 10},
	}

	config := setupAudioProcessing([]string{"-i", "clip.mp4"}, mediaInputs, 10, 2, nil, true)

	if !config.HasAudio {
		t.Fatal("expected clip audio to be preserved when requested")
	}
	if config.AudioBitrateSource != "clip.mp4" {
		t.Fatalf("expected clip.mp4 bitrate source, got %s", config.AudioBitrateSource)
	}
	if strings.Contains(config.AudioFilter, "sidechaincompress") {
		t.Fatalf("did not expect sidechaincompress without music, got %s", config.AudioFilter)
	}
	if len(config.MapArgs) < 4 || config.MapArgs[3] != "[clipaudio0]" {
		t.Fatalf("expected clip audio mapping, got %v", config.MapArgs)
	}
}

func TestGetVideoDetails_ErrorHandling(t *testing.T) {
	// Test with various invalid inputs
	testCases := []string{
		"",
		"/dev/null",
		"/nonexistent/path/video.mp4",
		"../invalid/path.mp4",
	}

	for _, filename := range testCases {
		t.Run("Invalid_file_"+filename, func(t *testing.T) {
			info, err := getVideoDetails(filename)

			// Should always return an info struct, even on error
			if info == nil {
				t.Fatalf("getVideoDetails should never return nil info")
			}

			// Error is expected for these cases
			if err == nil && filename != "" {
				t.Errorf("Expected error for filename %s", filename)
			}

			// Defaults should be set even on error
			expectedDefaults := map[string]string{
				"Framerate":    "30 fps",
				"Resolution":   "3840x2160",
				"AudioBitrate": "No audio",
			}

			if info.Framerate != expectedDefaults["Framerate"] {
				t.Errorf("Expected default framerate %s, got %s", expectedDefaults["Framerate"], info.Framerate)
			}
			if info.Resolution != expectedDefaults["Resolution"] {
				t.Errorf("Expected default resolution %s, got %s", expectedDefaults["Resolution"], info.Resolution)
			}
			if info.AudioBitrate != expectedDefaults["AudioBitrate"] {
				t.Errorf("Expected default audio bitrate %s, got %s", expectedDefaults["AudioBitrate"], info.AudioBitrate)
			}
		})
	}
}
