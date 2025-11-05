package utils

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
)

// ConvertImages processes each .jpg file in the working directory, applies scaling,
// compositing on a black background, and saves the output to the "converted" folder.
func ConvertImages() error {
	// Check if "converted" directory already exists.
	if _, err := os.Stat("converted"); err == nil {
		fmt.Println("The 'converted' folder already exists, skipping image conversion...")
		return nil // Exit the function without an error.
	}

	// First, check how many .jpg files we have before creating the directory.
	files, err := filepath.Glob("*.jpg")
	if err != nil {
		return fmt.Errorf("failed to list .jpg files: %v", err)
	}

	fileCount := len(files)

	if fileCount == 0 {
		return fmt.Errorf("no .jpg files found in current directory")
	}

	if fileCount < 2 {
		return fmt.Errorf("need at least 2 images to create a video, found only %d", fileCount)
	}

	// Create "converted" directory only after confirming we have enough images.
	if err := os.MkdirAll("converted", os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Display simple conversion info
	fmt.Printf("Converting %d images to 4K UHD...\n", fileCount)

	startTime := time.Now()
	var totalOriginalSize, totalConvertedSize int64

	for i, file := range files {
		// Simple progress indicator
		fmt.Printf("[%d/%d] Processing %s\n", i+1, fileCount, filepath.Base(file))

		// Get original file size
		if info, err := os.Stat(file); err == nil {
			totalOriginalSize += info.Size()
		}

		// Open image.
		img, err := imaging.Open(file, imaging.AutoOrientation(true))
		if err != nil {
			return fmt.Errorf("failed to open image %s: %v", file, err)
		}

		// Resize and process image.
		imgResized := imaging.Resize(img, 0, 2160, imaging.Lanczos)

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
	}

	// Display final statistics
	elapsed := time.Since(startTime)
	fmt.Printf("Converted %d images in %.1f seconds\n", fileCount, elapsed.Seconds())
	fmt.Printf("Total size: %.1f MB\n\n", float64(totalConvertedSize)/(1024*1024))

	return nil
}

// FetchImageTimestamp reads the timestamp from the image's EXIF data and returns it in YYYYMMDD_HHMMSS format.
// If decoding fails or the DateTime field is missing, the function returns the original filename without extension.
func FetchImageTimestamp(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = file.Close() // Ignore close errors in defer
	}()

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
