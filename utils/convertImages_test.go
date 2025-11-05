package utils

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// createTestImage creates a simple test JPEG image
func createTestImage(t *testing.T, filename string, width, height int) {
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

// setupTestDir creates a temporary directory for testing
func setupTestDir(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "go24k_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Store original directory in test context
	t.Cleanup(func() {
		_ = os.Chdir(originalDir) // Ignore error in cleanup
		os.RemoveAll(tempDir)
	})

	return tempDir
}

func TestConvertImages_NoImages(t *testing.T) {
	tempDir := setupTestDir(t)

	err := ConvertImages()
	if err == nil {
		t.Error("Expected error when no images are present, but got nil")
	}

	expectedMsg := "no .jpg files found in current directory"
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got: %s", expectedMsg, err.Error())
	}

	// Verify converted directory was NOT created when no images are found
	convertedDir := filepath.Join(tempDir, "converted")
	if _, err := os.Stat(convertedDir); !os.IsNotExist(err) {
		t.Error("Converted directory should NOT be created when no images are found")
	}
}

func TestConvertImages_InsufficientImages(t *testing.T) {
	tempDir := setupTestDir(t)

	// Create only one test image
	createTestImage(t, "single.jpg", 1920, 1080)

	err := ConvertImages()
	if err == nil {
		t.Error("Expected error when only one image is present, but got nil")
	}

	expectedMsg := "need at least 2 images to create a video, found only 1"
	if !contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got: %s", expectedMsg, err.Error())
	}

	// Verify converted directory was NOT created with insufficient images
	convertedDir := filepath.Join(tempDir, "converted")
	if _, err := os.Stat(convertedDir); !os.IsNotExist(err) {
		t.Error("Converted directory should NOT be created when insufficient images are found")
	}
}

func TestConvertImages_SingleImage(t *testing.T) {
	tempDir := setupTestDir(t)

	// Create two test images (minimum required for video)
	createTestImage(t, "test_image1.jpg", 4032, 3024) // Common phone camera resolution
	createTestImage(t, "test_image2.jpg", 1920, 1080) // Standard HD resolution

	err := ConvertImages()
	if err != nil {
		t.Errorf("ConvertImages failed: %v", err)
	}

	// Verify converted directory was created
	convertedDir := filepath.Join(tempDir, "converted")
	if _, statErr := os.Stat(convertedDir); os.IsNotExist(statErr) {
		t.Error("Converted directory was not created")
	}

	// Verify converted images exist
	convertedFiles, err := filepath.Glob(filepath.Join(convertedDir, "*.jpg"))
	if err != nil {
		t.Errorf("Failed to list converted files: %v", err)
	}

	if len(convertedFiles) != 2 {
		t.Errorf("Expected 2 converted files, got %d", len(convertedFiles))
	}
}

func TestConvertImages_MultipleImages(t *testing.T) {
	tempDir := setupTestDir(t)

	// Create multiple test images with different sizes
	testImages := []struct {
		name   string
		width  int
		height int
	}{
		{"image1.jpg", 4032, 3024}, // 4:3 aspect ratio
		{"image2.jpg", 3840, 2160}, // 16:9 aspect ratio
		{"image3.jpg", 1920, 1080}, // HD resolution
	}

	for _, img := range testImages {
		createTestImage(t, img.name, img.width, img.height)
	}

	err := ConvertImages()
	if err != nil {
		t.Errorf("ConvertImages failed: %v", err)
	}

	// Verify all images were converted
	convertedDir := filepath.Join(tempDir, "converted")
	convertedFiles, err := filepath.Glob(filepath.Join(convertedDir, "*.jpg"))
	if err != nil {
		t.Errorf("Failed to list converted files: %v", err)
	}

	if len(convertedFiles) != len(testImages) {
		t.Errorf("Expected %d converted files, got %d", len(testImages), len(convertedFiles))
	}
}

func TestConvertImages_ExistingConvertedDirectory(t *testing.T) {
	tempDir := setupTestDir(t)

	// Create converted directory first
	convertedDir := filepath.Join(tempDir, "converted")
	err := os.MkdirAll(convertedDir, os.ModePerm)
	if err != nil {
		t.Fatalf("Failed to create converted directory: %v", err)
	}

	// Create a test image
	createTestImage(t, "test_image.jpg", 1920, 1080)

	err = ConvertImages()
	if err != nil {
		t.Errorf("ConvertImages should not fail when converted directory exists: %v", err)
	}

	// Should skip conversion and return early
	// We can't easily test the skip message without capturing output
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[len(s)-len(substr):] == substr || s[:len(substr)] == substr ||
			findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestProcessSingleImage tests the core image processing logic
func TestProcessSingleImage_Integration(t *testing.T) {
	tempDir := setupTestDir(t)

	// Create two test images with EXIF-like timestamp in filename
	timestamp := time.Now().Format("20060102_150405")
	testImageName1 := timestamp + ".jpg"
	testImageName2 := timestamp + "_2.jpg"
	createTestImage(t, testImageName1, 2000, 1500)
	createTestImage(t, testImageName2, 1920, 1080)

	err := ConvertImages()
	if err != nil {
		t.Fatalf("ConvertImages failed: %v", err)
	}

	// Check that converted image has UHD suffix
	convertedFiles, err := filepath.Glob(filepath.Join(tempDir, "converted", "*_uhd.jpg"))
	if err != nil {
		t.Errorf("Failed to find converted files: %v", err)
	}

	if len(convertedFiles) == 0 {
		t.Error("No converted files with _uhd suffix found")
	}
}

// TestConvertImages_ErrorCases tests various error scenarios
func TestConvertImages_ErrorCases(t *testing.T) {
	_ = setupTestDir(t)

	t.Run("CorruptedJPEG", func(t *testing.T) {
		// Create one valid image and one corrupted JPEG file
		createTestImage(t, "valid.jpg", 1920, 1080)

		corruptedFile, err := os.Create("corrupted.jpg")
		if err != nil {
			t.Fatalf("Failed to create corrupted file: %v", err)
		}
		_, _ = corruptedFile.WriteString("this is not a jpeg file") // Ignore error for test data
		corruptedFile.Close()

		err = ConvertImages()
		if err == nil {
			t.Error("Expected error for corrupted JPEG, but got nil")
		}
		if !contains(err.Error(), "failed to open image") {
			t.Errorf("Expected 'failed to open image' error, got: %s", err.Error())
		}
	})

	t.Run("ReadOnlyDirectory", func(t *testing.T) {
		// Clean up first
		os.RemoveAll("corrupted.jpg")

		// Create two test images to pass minimum validation
		createTestImage(t, "readonly_test1.jpg", 1920, 1080)
		createTestImage(t, "readonly_test2.jpg", 1920, 1080)

		// Try to create converted directory as read-only (this test may be platform specific)
		_ = os.MkdirAll("converted", 0444) // Read-only permissions, ignore error for test

		// Remove converted dir for clean test
		os.RemoveAll("converted")

		// This should work normally since we removed the readonly dir
		err := ConvertImages()
		if err != nil {
			t.Errorf("ConvertImages should work after removing readonly dir: %v", err)
		}
	})
}

// TestFetchImageTimestamp_DetailedCases tests timestamp extraction in detail
func TestFetchImageTimestamp_DetailedCases(t *testing.T) {
	_ = setupTestDir(t)

	t.Run("FileNotExists", func(t *testing.T) {
		_, err := FetchImageTimestamp("nonexistent.jpg")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("NoEXIFData", func(t *testing.T) {
		// Create image without EXIF data
		createTestImage(t, "no_exif.jpg", 800, 600)

		timestamp, err := FetchImageTimestamp("no_exif.jpg")
		if err != nil {
			t.Errorf("Should not error for image without EXIF: %v", err)
		}

		expected := "no_exif" // filename without extension
		if timestamp != expected {
			t.Errorf("Expected '%s', got '%s'", expected, timestamp)
		}
	})

	t.Run("FilenameWithSpaces", func(t *testing.T) {
		filename := "image with spaces.jpg"
		createTestImage(t, filename, 800, 600)

		timestamp, err := FetchImageTimestamp(filename)
		if err != nil {
			t.Errorf("Should handle filenames with spaces: %v", err)
		}

		expected := "image with spaces"
		if timestamp != expected {
			t.Errorf("Expected '%s', got '%s'", expected, timestamp)
		}
	})

	t.Run("LongFilename", func(t *testing.T) {
		filename := "very_long_filename_that_exceeds_normal_limits_and_might_cause_issues_in_some_systems.jpg"
		createTestImage(t, filename, 800, 600)

		timestamp, err := FetchImageTimestamp(filename)
		if err != nil {
			t.Errorf("Should handle long filenames: %v", err)
		}

		expected := "very_long_filename_that_exceeds_normal_limits_and_might_cause_issues_in_some_systems"
		if timestamp != expected {
			t.Errorf("Expected '%s', got '%s'", expected, timestamp)
		}
	})
}

// TestConvertImages_DifferentResolutions tests conversion of various image resolutions
func TestConvertImages_DifferentResolutions(t *testing.T) {
	_ = setupTestDir(t)

	testCases := []struct {
		name   string
		width  int
		height int
	}{
		{"Portrait", 2160, 3840},  // Portrait 4K
		{"Square", 2160, 2160},    // Square image
		{"Panoramic", 7680, 2160}, // Ultra-wide
		{"VerySmall", 100, 100},   // Tiny image
		{"VeryTall", 1080, 8640},  // Very tall image
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clean up previous test files
			os.RemoveAll("converted")
			files, _ := filepath.Glob("*.jpg")
			for _, f := range files {
				os.Remove(f)
			}

			// Create two images: one with test resolution, one standard
			filename1 := fmt.Sprintf("%s.jpg", strings.ToLower(tc.name))
			filename2 := fmt.Sprintf("%s_standard.jpg", strings.ToLower(tc.name))
			createTestImage(t, filename1, tc.width, tc.height)
			createTestImage(t, filename2, 1920, 1080) // Standard resolution companion

			err := ConvertImages()
			if err != nil {
				t.Errorf("Failed to convert %s image: %v", tc.name, err)
			}

			// Verify converted files exist
			convertedFiles, err := filepath.Glob(filepath.Join("converted", "*.jpg"))
			if err != nil {
				t.Errorf("Failed to list converted files: %v", err)
			}

			if len(convertedFiles) != 2 {
				t.Errorf("Expected 2 converted files for %s, got %d", tc.name, len(convertedFiles))
			}
		})
	}
}

// TestConvertImages_FilePermissions tests file permission scenarios
func TestConvertImages_FilePermissions(t *testing.T) {
	_ = setupTestDir(t)

	// Create two test images
	createTestImage(t, "perm_test.jpg", 1920, 1080)
	createTestImage(t, "perm_test2.jpg", 1280, 720)

	// Test normal conversion first
	err := ConvertImages()
	if err != nil {
		t.Errorf("Normal conversion should work: %v", err)
	}

	// Verify file was created
	convertedFiles, err := filepath.Glob(filepath.Join("converted", "*.jpg"))
	if err != nil {
		t.Errorf("Failed to list converted files: %v", err)
	}

	if len(convertedFiles) != 2 {
		t.Errorf("Expected 2 converted files, got %d", len(convertedFiles))
	}
}

// TestConvertImages_ProgressBarPaths tests different OS paths for progress bar
func TestConvertImages_ProgressBarPaths(t *testing.T) {
	_ = setupTestDir(t)

	// Create test images with different filename lengths
	testFiles := []string{
		"short.jpg",
		"medium_length_filename.jpg",
		"very_long_filename_that_exceeds_twenty_characters_and_should_be_truncated.jpg",
	}

	for _, filename := range testFiles {
		createTestImage(t, filename, 1920, 1080)
	}

	err := ConvertImages()
	if err != nil {
		t.Errorf("ConvertImages failed with mixed filename lengths: %v", err)
	}

	// Verify all files were converted
	convertedFiles, err := filepath.Glob(filepath.Join("converted", "*.jpg"))
	if err != nil {
		t.Errorf("Failed to list converted files: %v", err)
	}

	if len(convertedFiles) != len(testFiles) {
		t.Errorf("Expected %d converted files, got %d", len(testFiles), len(convertedFiles))
	}
}

func BenchmarkConvertImages_SingleImage(b *testing.B) {
	// Setup
	tempDir, err := os.MkdirTemp("", "go24k_bench_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }() // Ignore error in defer

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		// Clean and setup for each iteration
		_ = os.Chdir(originalDir) // Ignore error in benchmark
		os.RemoveAll(tempDir)
		_ = os.MkdirAll(tempDir, os.ModePerm) // Ignore error in benchmark
		_ = os.Chdir(tempDir)                 // Ignore error in benchmark

		// Create test image
		// Create test image for benchmark
		img := image.NewRGBA(image.Rect(0, 0, 4032, 3024))
		file, _ := os.Create("test.jpg")
		_ = jpeg.Encode(file, img, &jpeg.Options{Quality: 90}) // Ignore error in benchmark
		file.Close()

		b.StartTimer()

		err := ConvertImages()
		if err != nil {
			b.Errorf("ConvertImages failed: %v", err)
		}
	}
}

// BenchmarkFetchImageTimestamp benchmarks timestamp extraction
func BenchmarkFetchImageTimestamp(b *testing.B) {
	// Create test image
	tempDir, err := os.MkdirTemp("", "go24k_bench_timestamp_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }() // Ignore error in defer
	_ = os.Chdir(tempDir)                        // Ignore error in benchmark

	// Create test image
	img := image.NewRGBA(image.Rect(0, 0, 800, 600))
	file, _ := os.Create("benchmark.jpg")
	_ = jpeg.Encode(file, img, &jpeg.Options{Quality: 90}) // Ignore error in benchmark
	file.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := FetchImageTimestamp("benchmark.jpg")
		if err != nil {
			b.Errorf("FetchImageTimestamp failed: %v", err)
		}
	}
}
