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

	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/schollz/progressbar/v3"
)

// ConvertImages processes each .jpg file in the working directory, applies scaling,
// compositing on a black background, and saves the output to the "converted" folder.
func ConvertImages() error {
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

	// On Windows, use a simpler progress bar to avoid issues.
	var bar *progressbar.ProgressBar
	if runtime.GOOS == "windows" {
		bar = progressbar.NewOptions(fileCount,
			progressbar.OptionSetDescription("Converting Pictures: "),
			progressbar.OptionShowCount(),
			progressbar.OptionSetWidth(30),
			progressbar.OptionOnCompletion(func() {
				fmt.Println()
			}),
		)
	} else {
		// For non-Windows, you can add an animated spinner.
		bar = progressbar.NewOptions(fileCount,
			progressbar.OptionSetDescription("Converting Pictures: "),
			progressbar.OptionShowCount(),
			progressbar.OptionSetWidth(30),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionOnCompletion(func() {
				fmt.Println()
			}),
		)
	}

	for _, file := range files {
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

		bar.Add(1)
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
