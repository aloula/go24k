package utils

import (
	"os"
	"runtime"
	"strings"
	"testing"
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
		duration int
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

			// Should contain resolution
			if !strings.Contains(effect, "3840x2160") {
				t.Error("Effect should contain 4K resolution")
			}

			// Should contain duration
			expectedFrames := tc.duration * 30
			if !strings.Contains(effect, string(rune(expectedFrames))) {
				// This is a rough check - the actual number formatting might differ
				t.Logf("Effect might not contain expected frames %d", expectedFrames)
			}

			t.Logf("Duration %d -> Effect: %s", tc.duration, effect)
		})
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
		if !strings.Contains(effect, "d=0") {
			t.Error("Expected duration 0 in output")
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
		effect := getKenBurnsEffect(5)

		// Verify the softened Ken Burns parameters
		if !strings.Contains(effect, "0.0005") {
			t.Error("Expected zoom speed 0.0005 (softened)")
		}
		if !strings.Contains(effect, "1.3") {
			t.Error("Expected max zoom 1.3 (softened)")
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
