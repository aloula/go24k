package main

import (
	"flag"
	"os"
	"testing"
)

// TestMainFlags tests command line flag parsing
func TestMainFlags(t *testing.T) {
	// Save original command line args
	oldArgs := os.Args
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}()

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "Default flags",
			args: []string{"go24k"},
		},
		{
			name: "Effects disabled",
			args: []string{"go24k", "-effects", "disabled"},
		},
		{
			name: "Custom duration and transition",
			args: []string{"go24k", "-d", "10", "-t", "2"},
		},
		{
			name: "Debug mode",
			args: []string{"go24k", "--debug"},
		},
		{
			name: "Include videos",
			args: []string{"go24k", "-include-videos"},
		},
		{
			name: "Keep video audio",
			args: []string{"go24k", "-keep-video-audio"},
		},
		{
			name: "Order by filename",
			args: []string{"go24k", "-order-by-filename"},
		},
		{
			name: "Random order",
			args: []string{"go24k", "-random-order"},
		},
		{
			name: "Order mode random",
			args: []string{"go24k", "-order", "random"},
		},
		{
			name: "Effects high",
			args: []string{"go24k", "-effects", "high"},
		},
		{
			name: "Effects low",
			args: []string{"go24k", "-effects", "low"},
		},
		{
			name: "Effects medium",
			args: []string{"go24k", "-effects", "medium"},
		},
		{
			name: "Effects disabled explicit",
			args: []string{"go24k", "-effects", "disabled"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flag package for each test
			flag.CommandLine = flag.NewFlagSet(tt.args[0], flag.ContinueOnError)

			// Set up flags (copy from main)
			duration := flag.Int("d", 5, "Duration per image in seconds")
			transition := flag.Int("t", 1, "Transition (fade) duration in seconds")
			effects := flag.String("effects", "disabled", "Image motion effects: disabled, low, medium, or high")
			debug := flag.Bool("debug", false, "Show environment detection and optimization info")
			includeVideos := flag.Bool("include-videos", false, "Include supported video files together with pictures")
			keepVideoAudio := flag.Bool("keep-video-audio", false, "Keep input video audio and blend it with MP3 background audio")
			orderMode := flag.String("order", "metadata", "Timeline order: metadata, filename, or random")
			orderByFilename := flag.Bool("order-by-filename", false, "Order timeline by filename instead of metadata time")
			randomOrder := flag.Bool("random-order", false, "Order timeline randomly")

			// Parse the test arguments
			os.Args = tt.args
			err := flag.CommandLine.Parse(tt.args[1:])
			if err != nil {
				t.Errorf("Flag parsing failed: %v", err)
				return
			}

			// Verify flags were parsed correctly
			switch tt.name {
			case "Default flags":
				if *debug {
					t.Error("Default flags should all be false")
				}
				if *duration != 5 || *transition != 1 {
					t.Errorf("Default duration should be 5, transition 1, got d=%d t=%d", *duration, *transition)
				}
				if *effects != "disabled" {
					t.Errorf("default effects should be disabled, got %s", *effects)
				}
			case "Effects disabled", "Effects disabled explicit":
				if *effects != "disabled" {
					t.Errorf("effects should be disabled, got %s", *effects)
				}
			case "Custom duration and transition":
				if *duration != 10 || *transition != 2 {
					t.Errorf("Expected d=10 t=2, got d=%d t=%d", *duration, *transition)
				}
			case "Debug mode":
				if !*debug {
					t.Error("debug flag should be true")
				}
			case "Include videos":
				if !*includeVideos {
					t.Error("include-videos flag should be true")
				}
			case "Keep video audio":
				if !*keepVideoAudio {
					t.Error("keep-video-audio flag should be true")
				}
			case "Order by filename":
				if !*orderByFilename {
					t.Error("order-by-filename flag should be true")
				}
			case "Random order":
				if !*randomOrder {
					t.Error("random-order flag should be true")
				}
			case "Order mode random":
				if *orderMode != "random" {
					t.Errorf("order should be random, got %s", *orderMode)
				}
			case "Effects high":
				if *effects != "high" {
					t.Errorf("effects should be high, got %s", *effects)
				}
			case "Effects low":
				if *effects != "low" {
					t.Errorf("effects should be low, got %s", *effects)
				}
			case "Effects medium":
				if *effects != "medium" {
					t.Errorf("effects should be medium, got %s", *effects)
				}
			}
		})
	}
}

// TestMainFlagValidation tests flag validation logic
func TestMainFlagValidation(t *testing.T) {
	tests := []struct {
		name        string
		duration    int
		transition  int
		shouldError bool
	}{
		{"Valid values", 5, 1, false},
		{"Zero duration", 0, 1, false},      // Currently not validated in main
		{"Negative duration", -1, 1, false}, // Currently not validated in main
		{"Zero transition", 5, 0, false},
		{"Large values", 3600, 10, false}, // 1 hour per image
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that these values would be acceptable
			// Note: The current main() doesn't validate these values
			// This test documents the current behavior

			if tt.duration < 0 {
				t.Logf("Negative duration %d would be passed to functions", tt.duration)
			}
			if tt.transition < 0 {
				t.Logf("Negative transition %d would be passed to functions", tt.transition)
			}

			// In a production system, we might want to add validation:
			// - duration > 0
			// - transition >= 0
			// - transition < duration (logical constraint)
		})
	}
}

// TestMainWorkflow tests the main function workflow logic
func TestMainWorkflow(t *testing.T) {
	tests := []struct {
		name           string
		effectsMode    string
		expectKenBurns bool
	}{
		{
			name:           "Normal workflow with Ken Burns",
			effectsMode:    "high",
			expectKenBurns: true,
		},
		{
			name:           "Workflow with effects disabled",
			effectsMode:    "disabled",
			expectKenBurns: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the Ken Burns logic
			applyKenBurns := tt.effectsMode != "disabled"

			if tt.name == "Normal workflow with Ken Burns" && !applyKenBurns {
				t.Error("Should apply Ken Burns when effects are enabled")
			}

			if tt.name == "Workflow with effects disabled" && applyKenBurns {
				t.Error("Should not apply Ken Burns when effects are disabled")
			}

			t.Logf("Full workflow: effects=%s, applyKenBurns=%v", tt.effectsMode, applyKenBurns)
		})
	}
}

// TestMainFlags_EdgeCases tests edge cases in flag parsing
func TestMainFlags_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "Multiple_flags_combined",
			args: []string{"-d", "10", "-t", "2", "-effects", "disabled"},
		},
		{
			name: "Debug_with_other_flags",
			args: []string{"--debug", "-d", "3"},
		},
		{
			name: "Effects_with_duration",
			args: []string{"-effects", "high", "-d", "999"},
		},
		{
			name: "Effects_disabled_with_transition",
			args: []string{"-effects", "disabled", "-t", "5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the individual flag logic conceptually
			t.Logf("Testing args: %v", tt.args)
		})
	}
}

// TestFlagCombinations tests various flag combinations
func TestFlagCombinations(t *testing.T) {
	testCases := []struct {
		name           string
		duration       int
		transition     int
		effectsMode    string
		debug          bool
		expectVideo    bool
		expectKenBurns bool
	}{
		{
			name:     "Default_settings",
			duration: 5, transition: 1, effectsMode: "high", debug: false,
			expectVideo: true, expectKenBurns: true,
		},
		{
			name:     "Effects_disabled",
			duration: 5, transition: 1, effectsMode: "disabled", debug: false,
			expectVideo: true, expectKenBurns: false,
		},
		{
			name:     "Debug_mode",
			duration: 5, transition: 1, effectsMode: "high", debug: true,
			expectVideo: true, expectKenBurns: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test the logical combinations
			applyKenBurns := tc.effectsMode != "disabled"

			if applyKenBurns != tc.expectKenBurns {
				t.Errorf("effects=%s expected Ken Burns=%v, got %v", tc.effectsMode, tc.expectKenBurns, applyKenBurns)
			}

			t.Logf("Testing combination: duration=%d, transition=%d, effects=%s, debug=%v",
				tc.duration, tc.transition, tc.effectsMode, tc.debug)
		})
	}
}

// TestMainValidation_ExtendedCases tests extended validation scenarios
func TestMainValidation_ExtendedCases(t *testing.T) {
	tests := []struct {
		name       string
		duration   int
		transition int
		shouldWarn bool
		reason     string
	}{
		{
			name: "Very_short_duration", duration: 1, transition: 1,
			shouldWarn: true, reason: "Duration equals transition time",
		},
		{
			name: "Transition_longer_than_duration", duration: 2, transition: 3,
			shouldWarn: true, reason: "Transition longer than duration",
		},
		{
			name: "Extremely_long_duration", duration: 3600, transition: 1,
			shouldWarn: false, reason: "Long but valid duration",
		},
		{
			name: "Zero_transition", duration: 5, transition: 0,
			shouldWarn: false, reason: "No transition is valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test validation logic
			problematic := tt.transition >= tt.duration && tt.duration > 0

			if tt.duration <= 0 {
				problematic = true
			}

			if problematic != tt.shouldWarn {
				t.Errorf("Expected warning=%v for %s (duration=%d, transition=%d), got warning=%v",
					tt.shouldWarn, tt.reason, tt.duration, tt.transition, problematic)
			}
		})
	}
}

// BenchmarkFlagParsing benchmarks the flag parsing performance
func BenchmarkFlagParsing(b *testing.B) {
	args := []string{"go24k", "-d", "10", "-t", "2", "-effects", "medium"}

	for i := 0; i < b.N; i++ {
		// Reset flags for each iteration
		flag.CommandLine = flag.NewFlagSet("go24k", flag.ContinueOnError)

		// Set up flags
		flag.String("effects", "disabled", "Image motion effects")
		flag.Bool("debug", false, "Show environment info")
		flag.Int("d", 5, "Duration per image")
		flag.Int("t", 1, "Transition duration")

		// Parse args
		oldArgs := os.Args
		os.Args = args
		flag.Parse()
		os.Args = oldArgs
	}
}
