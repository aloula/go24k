package utils

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
)

// CameraInfo contains EXIF data about the camera and photo settings
type CameraInfo struct {
	Make         string // Camera manufacturer
	Model        string // Camera model
	LensModel    string // Lens model
	FocalLength  string // Focal length (e.g., "50mm")
	ISO          string // ISO speed (e.g., "400")
	ExposureTime string // Shutter speed (e.g., "1/125s")
	FNumber      string // Aperture (e.g., "f/2.8")
}

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

	var totalOriginalSize, totalConvertedSize int64

	for i, file := range files {
		// Simple progress indicator
		fmt.Printf("[%d/%d] %s...\n", i+1, fileCount, filepath.Base(file))

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

// ExtractCameraInfo extracts camera and lens information from EXIF data
func ExtractCameraInfo(filename string) (*CameraInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	x, err := exif.Decode(file)
	if err != nil {
		return &CameraInfo{}, nil // Return empty struct if no EXIF
	}

	info := &CameraInfo{}

	// Extract camera make
	if tag, err := x.Get(exif.Make); err == nil {
		info.Make = strings.TrimSpace(tag.String())
	}

	// Extract camera model
	if tag, err := x.Get(exif.Model); err == nil {
		info.Model = strings.TrimSpace(tag.String())
	}

	// Extract lens model
	if tag, err := x.Get(exif.LensModel); err == nil {
		info.LensModel = strings.TrimSpace(tag.String())
	}

	// Extract focal length
	if tag, err := x.Get(exif.FocalLength); err == nil {
		// Try to get as rational number
		if ratNum, ratDenom, err := tag.Rat2(0); err == nil && ratDenom != 0 {
			focal := float64(ratNum) / float64(ratDenom)
			info.FocalLength = fmt.Sprintf("%.0fmm", focal)
		}
	}

	// Extract ISO
	if tag, err := x.Get(exif.ISOSpeedRatings); err == nil {
		if iso, err := tag.Int(0); err == nil {
			info.ISO = fmt.Sprintf("ISO %d", iso)
		}
	}

	// Extract exposure time (shutter speed)
	if tag, err := x.Get(exif.ExposureTime); err == nil {
		if expNum, expDenom, err := tag.Rat2(0); err == nil && expDenom != 0 {
			exp := float64(expNum) / float64(expDenom)
			if exp >= 1 {
				info.ExposureTime = fmt.Sprintf("%.1fs", exp)
			} else {
				// Convert to fraction format (e.g., 1/125s)
				denom := 1.0 / exp
				info.ExposureTime = fmt.Sprintf("1/%.0fs", denom)
			}
		}
	}

	// Extract f-number (aperture)
	if tag, err := x.Get(exif.FNumber); err == nil {
		if fNum, fDenom, err := tag.Rat2(0); err == nil && fDenom != 0 {
			f := float64(fNum) / float64(fDenom)
			info.FNumber = fmt.Sprintf("f/%.1f", f)
		}
	}

	return info, nil
}

// FormatCameraInfoOverlay formats camera information into a readable string for video overlay
func FormatCameraInfoOverlay(info *CameraInfo) string {
	if info == nil {
		return ""
	}

	var parts []string

	// Add camera info if available
	if info.Make != "" && info.Model != "" {
		parts = append(parts, fmt.Sprintf("%s %s", info.Make, info.Model))
	} else if info.Model != "" {
		parts = append(parts, info.Model)
	}

	// Add lens info if available
	if info.LensModel != "" {
		parts = append(parts, info.LensModel)
	}

	// Create second line with technical settings
	var techSettings []string
	if info.FocalLength != "" {
		techSettings = append(techSettings, info.FocalLength)
	}
	if info.FNumber != "" {
		techSettings = append(techSettings, info.FNumber)
	}
	if info.ExposureTime != "" {
		techSettings = append(techSettings, info.ExposureTime)
	}
	if info.ISO != "" {
		techSettings = append(techSettings, info.ISO)
	}

	if len(techSettings) > 0 {
		parts = append(parts, strings.Join(techSettings, " • "))
	}

	return strings.Join(parts, "\\n")
}

// GetOriginalFilename attempts to find the original image file corresponding to a converted file
// by matching the timestamp pattern in the converted filename
func GetOriginalFilename(convertedFile string) string {
	// Extract timestamp from converted filename
	// Format: converted/YYYYMMDD_HHMMSS_uhd.jpg
	baseName := filepath.Base(convertedFile)
	timestamp := strings.TrimSuffix(baseName, "_uhd.jpg")

	// Look for original files with matching timestamps
	files, err := filepath.Glob("*.jpg")
	if err != nil {
		return ""
	}

	for _, file := range files {
		// Skip if this is in the converted directory
		if strings.Contains(file, "converted/") {
			continue
		}

		// Extract timestamp from original file
		originalTimestamp, err := FetchImageTimestamp(file)
		if err != nil {
			continue
		}

		if originalTimestamp == timestamp {
			return file
		}
	}

	// Fallback: try to match by similar naming patterns
	for _, file := range files {
		if strings.Contains(file, "converted/") {
			continue
		}

		// If we can't find by timestamp, return the first available original file
		// This is a simple fallback that works for single-image scenarios
		return file
	}

	return ""
}
