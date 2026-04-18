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

// CameraInfo contains EXIF data about the camera and photo settings
type CameraInfo struct {
	Make         string // Camera manufacturer
	Model        string // Camera model
	LensModel    string // Lens model
	FocalLength  string // Focal length (e.g., "50mm")
	ISO          string // ISO speed (e.g., "400")
	ExposureTime string // Shutter speed (e.g., "1/125s")
	FNumber      string // Aperture (e.g., "f/2.8")
	DateTaken    string // Date the photo was taken (DD/MM/YYYY)
}

// ConvertImages processes each .jpg file in the working directory, applies scaling,
// compositing on a black background, and saves the output to the "converted" folder.
// If fullHD is true, the target canvas is Full HD (1920x1080); otherwise it is 4K UHD (3840x2160).
func ConvertImages(fullHD bool) error {
	// Determine canvas dimensions.
	targetWidth, targetHeight := 3840, 2160
	resLabel := "4K UHD"
	imageSuffix := "uhd"
	if fullHD {
		targetWidth, targetHeight = 1920, 1080
		resLabel = "Full HD"
		imageSuffix = "fhd"
	}

	// Check if "converted" directory already exists.
	if _, err := os.Stat("converted"); err == nil {
		convertedFiles, globErr := filepath.Glob(filepath.Join("converted", "*.jpg"))
		if globErr != nil {
			return fmt.Errorf("failed to inspect converted images: %v", globErr)
		}

		if len(convertedFiles) > 0 {
			sampleImg, openErr := imaging.Open(convertedFiles[0], imaging.AutoOrientation(true))
			if openErr != nil {
				fmt.Printf("Converted images are unreadable (%v), regenerating for %s...\n", openErr, resLabel)
				if rmErr := os.RemoveAll("converted"); rmErr != nil {
					return fmt.Errorf("failed to remove invalid converted folder: %v", rmErr)
				}
			} else {
				bounds := sampleImg.Bounds()
				if bounds.Dx() == targetWidth && bounds.Dy() == targetHeight {
					fmt.Println("The 'converted' folder already exists, skipping image conversion...")
					return nil // Existing converted images already match requested output resolution.
				}

				fmt.Printf("Converted images are %dx%d but requested output is %dx%d, rebuilding...\n", bounds.Dx(), bounds.Dy(), targetWidth, targetHeight)
				if rmErr := os.RemoveAll("converted"); rmErr != nil {
					return fmt.Errorf("failed to remove converted folder for resolution rebuild: %v", rmErr)
				}
			}
		} else {
			fmt.Println("The 'converted' folder already exists, skipping image conversion...")
			return nil // Preserve existing behavior for an empty converted folder.
		}
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
	fmt.Printf("Converting %d images to %s...\n", fileCount, resLabel)

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

		// Fit image inside target canvas without cropping, allowing upscale when needed.
		imgResized := resizeImageToCanvas(img, targetWidth, targetHeight)

		// Create a black background.
		uhdBlack := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
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
		filenameConverted := filepath.Join("converted", fmt.Sprintf("%s_%s.jpg", timestamp, imageSuffix))
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

func resizeImageToCanvas(img image.Image, targetWidth, targetHeight int) *image.NRGBA {
	bounds := img.Bounds()
	sourceWidth := bounds.Dx()
	sourceHeight := bounds.Dy()

	if sourceWidth <= 0 || sourceHeight <= 0 {
		return imaging.Resize(img, targetWidth, targetHeight, imaging.Lanczos)
	}

	widthScale := float64(targetWidth) / float64(sourceWidth)
	heightScale := float64(targetHeight) / float64(sourceHeight)
	scale := widthScale
	if heightScale < scale {
		scale = heightScale
	}

	resizedWidth := int(float64(sourceWidth)*scale + 0.5)
	resizedHeight := int(float64(sourceHeight)*scale + 0.5)
	if resizedWidth < 1 {
		resizedWidth = 1
	}
	if resizedHeight < 1 {
		resizedHeight = 1
	}

	return imaging.Resize(img, resizedWidth, resizedHeight, imaging.Lanczos)
}

func trimConvertedImageResolutionSuffix(baseName string) string {
	if strings.HasSuffix(baseName, "_uhd.jpg") {
		return strings.TrimSuffix(baseName, "_uhd.jpg")
	}
	if strings.HasSuffix(baseName, "_fhd.jpg") {
		return strings.TrimSuffix(baseName, "_fhd.jpg")
	}
	return strings.TrimSuffix(baseName, filepath.Ext(baseName))
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

// simplifyBrandName removes common suffixes from camera brand names
func simplifyBrandName(brand string) string {
	// List of suffixes to remove
	suffixes := []string{
		" Corporation",
		" CORPORATION",
		" Corp.",
		" Corp",
		" Co., Ltd.",
		" Co., Ltd",
		" Ltd.",
		" Ltd",
		" Inc.",
		" Inc",
	}

	for _, suffix := range suffixes {
		if strings.HasSuffix(brand, suffix) {
			return strings.TrimSpace(strings.TrimSuffix(brand, suffix))
		}
	}

	return brand
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
		makeStr := strings.TrimSpace(tag.String())
		// Remove surrounding quotes if present
		makeStr = strings.Trim(makeStr, `"`)
		// Simplify brand name (remove Corporation, etc.)
		makeStr = simplifyBrandName(makeStr)
		info.Make = makeStr
	}

	// Extract camera model
	if tag, err := x.Get(exif.Model); err == nil {
		modelStr := strings.TrimSpace(tag.String())
		// Remove surrounding quotes if present
		modelStr = strings.Trim(modelStr, `"`)
		info.Model = modelStr
	}

	// Extract lens model
	if tag, err := x.Get(exif.LensModel); err == nil {
		lensStr := strings.TrimSpace(tag.String())
		// Remove surrounding quotes if present
		lensStr = strings.Trim(lensStr, `"`)
		info.LensModel = lensStr
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

	// Extract date taken (DateTime or DateTimeOriginal)
	if tag, err := x.Get(exif.DateTimeOriginal); err == nil {
		dateStr := strings.TrimSpace(tag.String())
		dateStr = strings.Trim(dateStr, `"`)
		// Parse EXIF date format: "2006:01:02 15:04:05"
		if t, err := time.Parse("2006:01:02 15:04:05", dateStr); err == nil {
			info.DateTaken = t.Format("02/01/2006")
		}
	} else if tag, err := x.Get(exif.DateTime); err == nil {
		// Fallback to DateTime if DateTimeOriginal is not available
		dateStr := strings.TrimSpace(tag.String())
		dateStr = strings.Trim(dateStr, `"`)
		if t, err := time.Parse("2006:01:02 15:04:05", dateStr); err == nil {
			info.DateTaken = t.Format("02/01/2006")
		}
	}

	return info, nil
}

// FormatCameraInfoOverlay formats camera information and creates FFmpeg drawtext filter
// with specified fontSize, positioned in the footer (bottom center)
func FormatCameraInfoOverlay(info *CameraInfo, fontSize, imageIndex int) string {
	if info == nil {
		return ""
	}

	// Show only camera model in overlay (omit manufacturer).
	cameraName := info.Model

	// If no camera info, return empty
	if cameraName == "" {
		return ""
	}

	// Build technical settings with dash separators
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

	// Use photo date if available, otherwise current date as fallback
	var dateStr string
	if info.DateTaken != "" {
		dateStr = info.DateTaken
	} else {
		// Fallback to current date if no date found in EXIF
		currentTime := time.Now()
		dateStr = currentTime.Format("02/01/2006")
	}

	// Build final string: "Camera - TechSettings - Date"
	var overlayText string
	if len(techSettings) > 0 {
		overlayText = fmt.Sprintf("%s - %s - %s", cameraName, strings.Join(techSettings, " - "), dateStr)
	} else {
		overlayText = fmt.Sprintf("%s - %s", cameraName, dateStr)
	}

	// Write text to a temporary file to avoid escaping issues
	// Each image gets its own overlay file
	textFile := fmt.Sprintf("converted/overlay_%d.txt", imageIndex)
	if err := os.WriteFile(textFile, []byte(overlayText), 0644); err != nil {
		// Fallback to inline text with escaping if file write fails
		overlayText = strings.ReplaceAll(overlayText, "|", "-")
		overlayText = strings.ReplaceAll(overlayText, ":", " ")
		overlayText = strings.ReplaceAll(overlayText, "/", "\\/")
		overlayText = strings.ReplaceAll(overlayText, " ", "\\ ")

		xPosition := "(w-tw)/2"
		yPosition := "h-th-40"
		if activeResolution == resolutionFullHD {
			yPosition = "h-th-30"
		}
		return fmt.Sprintf(",drawtext=text=%s:fontsize=%d:fontcolor=white:x=%s:y=%s:box=1:boxcolor=black@0.5:boxborderw=5",
			overlayText, fontSize, xPosition, yPosition)
	}

	// Position fixed at footer (bottom center)
	xPosition := "(w-tw)/2" // Horizontal center
	yPosition := "h-th-40"  // Bottom with 40px margin in UHD, 30px in Full HD
	if activeResolution == resolutionFullHD {
		yPosition = "h-th-30"
	}

	// Build the complete FFmpeg drawtext filter using textfile parameter with reload
	// Add reload=1 to force FFmpeg to read the file content for each frame
	drawtextFilter := fmt.Sprintf(",drawtext=textfile='%s':reload=1:fontsize=%d:fontcolor=white:x=%s:y=%s:box=1:boxcolor=black@0.5:boxborderw=5",
		textFile, fontSize, xPosition, yPosition)

	return drawtextFilter
}

// GetOriginalFilename attempts to find the original image file corresponding to a converted file
// by matching the timestamp pattern in the converted filename
func GetOriginalFilename(convertedFile string) string {
	// Extract timestamp from converted filename
	// Format: converted/YYYYMMDD_HHMMSS_uhd.jpg or converted/YYYYMMDD_HHMMSS_fhd.jpg
	baseName := filepath.Base(convertedFile)
	timestamp := trimConvertedImageResolutionSuffix(baseName)

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

	// Fallback: when timestamp matching fails, use the timestamp from the converted name
	// This preserves the alphabetical ordering when original files can't be found.
	// Only do this when the extracted name looks like a valid timestamp (YYYYMMDD_HHMMSS).
	if len(timestamp) == 15 && timestamp[8] == '_' {
		return timestamp + ".jpg"
	}

	return ""
}
