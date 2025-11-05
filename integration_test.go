//go:build integration

package main

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"go24k/utils"
)

// Integration tests require the binary to be built and FFmpeg to be available
// Run with: go test -tags=integration

func TestIntegrationFullWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if FFmpeg is available
	if !isFFmpegAvailable() {
		t.Skip("FFmpeg not available, skipping integration test")
	}

	// Setup test directory
	tempDir, err := os.MkdirTemp("", "go24k_integration_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to test directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Create test images (simple solid color JPEGs)
	createTestImages(t, tempDir)

	// Build the binary if it doesn't exist
	binaryPath := filepath.Join(originalDir, "go24k")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Log("Building go24k binary...")
		cmd := exec.Command("go", "build", "-o", binaryPath)
		cmd.Dir = originalDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to build binary: %v", err)
		}
	}

	// Test convert-only mode
	t.Run("ConvertOnly", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "-convert-only")
		cmd.Dir = tempDir

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("Convert-only failed: %v\nOutput: %s", err, output)
			return
		}

		// Verify converted directory exists
		convertedDir := filepath.Join(tempDir, "converted")
		if _, err := os.Stat(convertedDir); os.IsNotExist(err) {
			t.Error("Converted directory not created")
		}

		// Verify converted images exist
		convertedFiles, err := filepath.Glob(filepath.Join(convertedDir, "*.jpg"))
		if err != nil {
			t.Errorf("Failed to list converted files: %v", err)
		}

		if len(convertedFiles) == 0 {
			t.Error("No converted images found")
		}

		t.Logf("Convert-only successful: %d images converted", len(convertedFiles))
	})

	// Clean up for next test
	os.RemoveAll(filepath.Join(tempDir, "converted"))

	// Test full workflow (video generation)
	t.Run("FullWorkflow", func(t *testing.T) {
		// Use shorter duration for faster testing
		cmd := exec.Command(binaryPath, "-d", "2", "-t", "1")
		cmd.Dir = tempDir

		start := time.Now()
		output, err := cmd.CombinedOutput()
		duration := time.Since(start)

		if err != nil {
			t.Errorf("Full workflow failed: %v\nOutput: %s", err, output)
			return
		}

		// Verify video was created
		videoPath := filepath.Join(tempDir, "video.mp4")
		if _, err := os.Stat(videoPath); os.IsNotExist(err) {
			t.Error("Video file not created")
		} else {
			// Check video file size (should be > 0)
			if info, err := os.Stat(videoPath); err == nil {
				if info.Size() == 0 {
					t.Error("Video file is empty")
				} else {
					t.Logf("Video created: %d bytes in %v", info.Size(), duration)
				}
			}
		}
	})
}

func TestIntegrationStaticMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if !isFFmpegAvailable() {
		t.Skip("FFmpeg not available")
	}

	// Similar setup as above but test static mode
	tempDir, err := os.MkdirTemp("", "go24k_static_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	createTestImages(t, tempDir)

	binaryPath := filepath.Join(originalDir, "go24k")

	// Test static mode (no Ken Burns effect)
	cmd := exec.Command(binaryPath, "-static", "-d", "2", "-t", "1")
	cmd.Dir = tempDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Static mode failed: %v\nOutput: %s", err, output)
		return
	}

	// Verify video was created
	if _, err := os.Stat("video.mp4"); os.IsNotExist(err) {
		t.Error("Video file not created in static mode")
	}

	t.Log("Static mode video generation successful")
}

func TestIntegrationDebugMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	originalDir, _ := os.Getwd()
	binaryPath := filepath.Join(originalDir, "go24k")

	// Test debug mode (should not process any files)
	cmd := exec.Command(binaryPath, "--debug")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Debug mode failed: %v", err)
		return
	}

	// Debug mode should output system information
	outputStr := string(output)
	if len(outputStr) == 0 {
		t.Error("Debug mode produced no output")
	}

	// Should contain environment info
	expectedStrings := []string{"Environment Detection", "Hardware"}
	for _, expected := range expectedStrings {
		if !contains(outputStr, expected) {
			t.Errorf("Debug output missing expected string: %s", expected)
		}
	}

	t.Log("Debug mode successful")
}

// Helper functions

func isFFmpegAvailable() bool {
	cmd := exec.Command("ffmpeg", "-version")
	return cmd.Run() == nil
}

func createTestImages(t *testing.T, dir string) {
	// Create simple test JPEG images
	images := []string{"test1.jpg", "test2.jpg", "test3.jpg"}

	for _, imgName := range images {
		// Use ImageMagick or create programmatically
		// For simplicity, we'll create via Go's image package
		createSimpleJPEG(t, filepath.Join(dir, imgName))
	}
}

func createSimpleJPEG(t *testing.T, filename string) {
	// Create a simple colored image
	// This is a simplified version - in practice you'd want more realistic test images

	cmd := exec.Command("convert", "-size", "1920x1080", "xc:blue", filename)
	if err := cmd.Run(); err != nil {
		// Fallback: try to create with Go's image package
		t.Logf("ImageMagick not available, using Go image creation for %s", filename)
		createImageWithGo(t, filename)
	}
}

func createImageWithGo(t *testing.T, _ string) {
	// Fallback method using Go's image packages
	// (Implementation similar to what we have in the unit tests)
	t.Skip("Image creation fallback not fully implemented")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Helper function to create test images for benchmarks
func createTestImage(t testing.TB, filename string, width, height int) {
	// Create a simple test image
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with a simple pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Create a simple gradient
			r := uint8((x * 255) / width)
			g := uint8((y * 255) / height)
			b := uint8(128)
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	// Save as JPEG
	file, err := os.Create(filename)
	if err != nil {
		t.Fatalf("Failed to create test image file: %v", err)
	}
	defer file.Close()

	err = jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
	if err != nil {
		t.Fatalf("Failed to encode test image: %v", err)
	}
}

// BenchmarkIntegrationConversion benchmarks the full conversion process
func BenchmarkIntegrationConversion(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "go24k_bench_integration_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		// Setup for each iteration
		os.Chdir(originalDir)
		os.RemoveAll(tempDir)
		os.MkdirAll(tempDir, os.ModePerm)
		os.Chdir(tempDir)

		// Create test images using existing helper
		createTestImage(b, "bench1.jpg", 1920, 1080)
		createTestImage(b, "bench2.jpg", 2048, 1536)

		b.StartTimer()

		err := utils.ConvertImages()
		if err != nil {
			b.Errorf("Benchmark conversion failed: %v", err)
		}
	}
}
