package utils

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/schollz/progressbar/v3"
)

// ConvertImages processes each .jpg file in the working directory, applies scaling,
// compositing on a black background, and saves the output to the "converted" folder.
func ConvertImages() error {
	// Check if "converted" directory already exists.
	if _, err := os.Stat("converted"); err == nil {
		fmt.Println("The 'converted' folder already exists, skipping image conversion...")
		return nil // Exit the function without an error.
	}

	// Create "converted" directory.
	if err := os.MkdirAll("converted", os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Process each .jpg file.
	files, err := filepath.Glob("*.jpg")
	if err != nil {
		return fmt.Errorf("failed to list .jpg files: %v", err)
	}

	fileCount := len(files)

	if fileCount == 0 {
		return fmt.Errorf("no .jpg files found in current directory")
	}

	// Display conversion info
	fmt.Printf("\n=== Starting Image Conversion ===\n")
	fmt.Printf("Images to process: %d\n", fileCount)
	fmt.Printf("Target: 4K UHD (3840x2160) with 2160p height scaling\n")
	fmt.Printf("Output: converted/ directory\n\n")

	startTime := time.Now()
	var totalOriginalSize, totalConvertedSize int64

	// On Windows, use a simpler progress bar to avoid issues.
	var bar *progressbar.ProgressBar
	if runtime.GOOS == "windows" {
		bar = progressbar.NewOptions(fileCount,
			progressbar.OptionSetDescription("Converting"),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetWidth(40),
			progressbar.OptionSetRenderBlankState(true),
			progressbar.OptionOnCompletion(func() {
				fmt.Printf("\n=== Image conversion completed! ===\n\n")
			}),
		)
	} else {
		// For non-Windows, you can add an animated spinner and emojis.
		bar = progressbar.NewOptions(fileCount,
			progressbar.OptionSetDescription("🔄 Converting"),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetWidth(40),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionSetRenderBlankState(true),
			progressbar.OptionOnCompletion(func() {
				fmt.Printf("\n✅ Image conversion completed!\n\n")
			}),
		)
	}

	for _, file := range files {
		// Get original file size
		if info, err := os.Stat(file); err == nil {
			totalOriginalSize += info.Size()
		}

		// Open image.
		img, err := imaging.Open(file, imaging.AutoOrientation(true))
		if err != nil {
			return fmt.Errorf("failed to open image %s: %v", file, err)
		}

		// Get original image dimensions
		bounds := img.Bounds()
		originalWidth := bounds.Dx()
		originalHeight := bounds.Dy()

		// Resize and process image.
		imgResized := imaging.Resize(img, 0, 2160, imaging.Lanczos)

		// Get resized image dimensions
		resizedBounds := imgResized.Bounds()
		resizedWidth := resizedBounds.Dx()
		resizedHeight := resizedBounds.Dy()

		// Create a black background.
		uhdBlack := image.NewRGBA(image.Rect(0, 0, 3840, 2160))
		black := color.RGBA{0, 0, 0, 255}
		draw.Draw(uhdBlack, uhdBlack.Bounds(), &image.Uniform{black}, image.Point{}, draw.Src)

		// Composite the resized image onto the black background.
		imgConverted := imaging.OverlayCenter(uhdBlack, imgResized, 1.0)

		// Get image timestamp.
		timestamp, err := FetchImageTimestamp(file)
		if err != nil {
			return fmt.Errorf("failed to get image timestamp for %s: %v", file, err)
		}

		// Save converted image.
		filenameConverted := filepath.Join("converted", fmt.Sprintf("%s_uhd.jpg", timestamp))
		if err := imaging.Save(imgConverted, filenameConverted); err != nil {
			return fmt.Errorf("failed to save converted image %s: %v", filenameConverted, err)
		}

		// Get converted file size
		if info, err := os.Stat(filenameConverted); err == nil {
			totalConvertedSize += info.Size()
		}

		// Update progress bar with resolution information
		shortFilename := filepath.Base(file)
		if len(shortFilename) > 20 {
			shortFilename = shortFilename[:17] + "..."
		}
		// Use different format based on OS for better compatibility
		if runtime.GOOS == "windows" {
			bar.Describe(fmt.Sprintf("Converting %s (%dx%d->%dx%d)",
				shortFilename, originalWidth, originalHeight, resizedWidth, resizedHeight))
		} else {
			bar.Describe(fmt.Sprintf("Converting %s (%dx%d→%dx%d)",
				shortFilename, originalWidth, originalHeight, resizedWidth, resizedHeight))
		}
		bar.Add(1)
	}

	// Display final statistics
	elapsed := time.Since(startTime)
	avgSpeed := float64(fileCount) / elapsed.Seconds()

	if runtime.GOOS == "windows" {
		fmt.Printf("=== Conversion Statistics ===\n")
		fmt.Printf("   Processing time: %.1f seconds\n", elapsed.Seconds())
		fmt.Printf("   Average speed: %.1f images/sec\n", avgSpeed)
		fmt.Printf("   Original total size: %.1f MB\n", float64(totalOriginalSize)/(1024*1024))
		fmt.Printf("   Converted total size: %.1f MB\n", float64(totalConvertedSize)/(1024*1024))
		if totalOriginalSize > 0 {
			fmt.Printf("   Size ratio: %.1fx\n", float64(totalConvertedSize)/float64(totalOriginalSize))
		}
	} else {
		fmt.Printf("📈 Conversion Statistics:\n")
		fmt.Printf("   ⏱️  Processing time: %.1f seconds\n", elapsed.Seconds())
		fmt.Printf("   🚀 Average speed: %.1f images/sec\n", avgSpeed)
		fmt.Printf("   📁 Original total size: %.1f MB\n", float64(totalOriginalSize)/(1024*1024))
		fmt.Printf("   📁 Converted total size: %.1f MB\n", float64(totalConvertedSize)/(1024*1024))
		if totalOriginalSize > 0 {
			fmt.Printf("   📊 Size ratio: %.1fx\n", float64(totalConvertedSize)/float64(totalOriginalSize))
		}
	}

	return nil
}

// FetchImageTimestamp reads the timestamp from the image's EXIF data and returns it in YYYYMMDD_HHMMSS format.
// If decoding fails or the DateTime field is missing, the function returns the original filename without extension.
func FetchImageTimestamp(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	x, err := exif.Decode(file)
	if err != nil {
		return strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename)), nil
	}

	tm, err := x.DateTime()
	if err != nil {
		return strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename)), nil
	}

	return tm.Format("20060102_150405"), nil
}
