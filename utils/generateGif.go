package utils

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/schollz/progressbar/v3"
)

// CountImages returns the number of JPEG images in the current directory
func CountImages() int {
	files, err := filepath.Glob("*.jpg")
	if err != nil {
		return 0
	}
	return len(files)
}

// ConvertImagesForGif processes JPEG images optimized for GIF creation
// maxHeight: maximum height for the converted images (e.g., 1080 for better GIF performance)
func ConvertImagesForGif(maxHeight int) error {
	// Check if "gif_converted" directory already exists
	if _, err := os.Stat("gif_converted"); err == nil {
		fmt.Println("The 'gif_converted' folder already exists, skipping image conversion...")
		return nil
	}

	// Create "gif_converted" directory
	if err := os.MkdirAll("gif_converted", os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Process each .jpg file
	files, err := filepath.Glob("*.jpg")
	if err != nil {
		return fmt.Errorf("failed to list .jpg files: %v", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no .jpg files found in current directory")
	}

	fileCount := len(files)

	// Create progress bar
	var bar *progressbar.ProgressBar
	if runtime.GOOS == "windows" {
		bar = progressbar.NewOptions(fileCount,
			progressbar.OptionSetDescription("Converting for GIF: "),
			progressbar.OptionShowCount(),
			progressbar.OptionSetWidth(30),
			progressbar.OptionOnCompletion(func() {
				fmt.Println()
			}),
		)
	} else {
		bar = progressbar.NewOptions(fileCount,
			progressbar.OptionSetDescription("Converting for GIF: "),
			progressbar.OptionShowCount(),
			progressbar.OptionSetWidth(30),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionOnCompletion(func() {
				fmt.Println()
			}),
		)
	}

	for i, file := range files {
		// Open image
		img, err := imaging.Open(file, imaging.AutoOrientation(true))
		if err != nil {
			return fmt.Errorf("failed to open image %s: %v", file, err)
		}

		// Get original dimensions
		bounds := img.Bounds()
		originalWidth := bounds.Dx()
		originalHeight := bounds.Dy()

		// Calculate new dimensions maintaining aspect ratio
		var newWidth, newHeight int
		if originalHeight > maxHeight {
			// Resize based on height
			newHeight = maxHeight
			newWidth = int(float64(originalWidth) * float64(maxHeight) / float64(originalHeight))
		} else {
			// Keep original size if it's already smaller
			newWidth = originalWidth
			newHeight = originalHeight
		}

		// Resize image
		imgResized := imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)

		// Create a black background with appropriate aspect ratio for the final image
		// We'll use a 16:9 aspect ratio as a good default for GIFs
		finalWidth := newWidth
		finalHeight := newHeight

		// If the image is very wide or very tall, we might want to center it on a black background
		// But for GIFs, it's often better to just use the natural dimensions

		blackBg := image.NewRGBA(image.Rect(0, 0, finalWidth, finalHeight))
		black := color.RGBA{0, 0, 0, 255}
		draw.Draw(blackBg, blackBg.Bounds(), &image.Uniform{black}, image.Point{}, draw.Src)

		// Center the resized image on the black background
		imgFinal := imaging.OverlayCenter(blackBg, imgResized, 1.0)

		// Generate filename with index to maintain order
		filename := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
		filenameConverted := filepath.Join("gif_converted", fmt.Sprintf("%03d_%s.jpg", i, filename))

		// Save converted image
		if err := imaging.Save(imgFinal, filenameConverted); err != nil {
			return fmt.Errorf("failed to save converted image %s: %v", filenameConverted, err)
		}

		bar.Add(1)
	}

	return nil
}

// GenerateGif creates an animated GIF from GIF-optimized images with transitions.
// duration: seconds per image
// transitionDuration: fade transition duration in seconds
// fps: frames per second for the GIF (lower values = smaller files)
// scale: additional scale factor if needed (usually 1.0 since images are already optimized)
func GenerateGif(duration, transitionDuration int, fps int, scale float64) {
	// First, convert images optimized for GIF (1080p max height)
	if err := ConvertImagesForGif(1080); err != nil {
		log.Fatalf("Failed to convert images for GIF: %v", err)
	}

	// Find all GIF-optimized .jpg files
	files, err := filepath.Glob("gif_converted/*.jpg")
	if err != nil {
		log.Fatalf("Failed to list gif_converted .jpg files: %v", err)
	}

	if len(files) == 0 {
		log.Fatalf("No converted images found for GIF generation.")
	}

	fmt.Printf("Creating animated GIF from %d images...\n", len(files))

	// Show progress
	done := make(chan struct{})
	go func() {
		spinnerChars := []string{"|", "/", "-", "\\"}
		i := 0
		for {
			select {
			case <-done:
				fmt.Print("\r")
				return
			default:
				fmt.Printf("\rGenerating GIF... %s", spinnerChars[i%len(spinnerChars)])
				i++
				time.Sleep(200 * time.Millisecond)
			}
		}
	}()

	// Build FFmpeg command - much simpler since images are already optimized
	args := []string{"-y"}

	// Add all input files
	for _, file := range files {
		args = append(args, "-loop", "1", "-t", fmt.Sprintf("%d", duration), "-i", file)
	}

	// Build simple filter complex
	filterComplex := ""

	// Apply scale if needed, otherwise just prepare videos
	for i := 0; i < len(files); i++ {
		if scale != 1.0 {
			filterComplex += fmt.Sprintf("[%d:v]scale=iw*%.2f:ih*%.2f,setsar=1[v%d];", i, scale, scale, i)
		} else {
			filterComplex += fmt.Sprintf("[%d:v]setsar=1[v%d];", i, i)
		}
	}

	// Concatenate all videos
	for i := 0; i < len(files); i++ {
		filterComplex += fmt.Sprintf("[v%d]", i)
	}
	filterComplex += fmt.Sprintf("concat=n=%d:v=1:a=0[out]", len(files))

	args = append(args, "-filter_complex", filterComplex)
	args = append(args, "-map", "[out]")
	args = append(args, "-r", fmt.Sprintf("%d", fps))
	args = append(args, "-f", "gif")
	args = append(args, "animated.gif")

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stderr = os.Stderr // Show FFmpeg output for debugging

	if err := cmd.Run(); err != nil {
		close(done)
		log.Fatalf("FFmpeg command failed: %v", err)
	}
	close(done)

	// Get file size for user feedback
	if fileInfo, err := os.Stat("animated.gif"); err == nil {
		sizeKB := fileInfo.Size() / 1024
		sizeMB := float64(sizeKB) / 1024
		fmt.Printf("\nGIF created successfully: animated.gif (%.1f MB)\n", sizeMB)
	}
}

// GenerateOptimizedGif creates a smaller, optimized GIF using palette optimization
func GenerateOptimizedGif(duration, transitionDuration int, fps int, scale float64) {
	// First, convert images optimized for GIF (1080p max height)
	if err := ConvertImagesForGif(1080); err != nil {
		log.Fatalf("Failed to convert images for GIF: %v", err)
	}

	// Find all GIF-optimized .jpg files
	files, err := filepath.Glob("gif_converted/*.jpg")
	if err != nil {
		log.Fatalf("Failed to list gif_converted .jpg files: %v", err)
	}

	if len(files) == 0 {
		log.Fatalf("No converted images found for GIF generation.")
	}

	fmt.Printf("Creating optimized animated GIF from %d images...\n", len(files))

	// Step 1: Create a simplified palette from just the first image
	fmt.Println("Generating optimized palette...")
	paletteArgs := []string{"-y", "-i", files[0]}

	// Create palette filter with optional scaling
	paletteFilter := "palettegen=max_colors=256"
	if scale != 1.0 {
		paletteFilter = fmt.Sprintf("scale=iw*%.2f:ih*%.2f,%s", scale, scale, paletteFilter)
	}

	paletteArgs = append(paletteArgs, "-vf", paletteFilter)
	paletteArgs = append(paletteArgs, "-t", "1") // Only generate 1 second for palette
	paletteArgs = append(paletteArgs, "palette.png")

	cmd := exec.Command("ffmpeg", paletteArgs...)
	// Don't show stderr unless there's an error
	if err := cmd.Run(); err != nil {
		log.Printf("Palette generation failed, falling back to regular GIF generation: %v", err)
		// Fallback to regular GIF generation
		GenerateGif(duration, transitionDuration, fps, scale)
		return
	}

	// Step 2: Create basic GIF first, then optimize with palette
	fmt.Println("Creating optimized GIF...")

	// First create a basic gif
	tempGifArgs := []string{"-y"}

	// Add all input files
	for _, file := range files {
		tempGifArgs = append(tempGifArgs, "-loop", "1", "-t", fmt.Sprintf("%d", duration), "-i", file)
	}

	// Build simple concatenation filter
	filterComplex := ""
	for i := 0; i < len(files); i++ {
		if scale != 1.0 {
			filterComplex += fmt.Sprintf("[%d:v]scale=iw*%.2f:ih*%.2f,setsar=1[v%d];", i, scale, scale, i)
		} else {
			filterComplex += fmt.Sprintf("[%d:v]setsar=1[v%d];", i, i)
		}
	}

	for i := 0; i < len(files); i++ {
		filterComplex += fmt.Sprintf("[v%d]", i)
	}
	filterComplex += fmt.Sprintf("concat=n=%d:v=1:a=0[out]", len(files))

	tempGifArgs = append(tempGifArgs, "-filter_complex", filterComplex)
	tempGifArgs = append(tempGifArgs, "-map", "[out]")
	tempGifArgs = append(tempGifArgs, "-r", fmt.Sprintf("%d", fps))
	tempGifArgs = append(tempGifArgs, "temp.gif")

	cmd = exec.Command("ffmpeg", tempGifArgs...)
	if err := cmd.Run(); err != nil {
		log.Printf("Basic GIF creation failed, cleaning up: %v", err)
		os.Remove("palette.png")
		return
	}

	// Step 3: Apply palette to the basic GIF
	fmt.Println("Applying palette optimization...")

	paletteApplyArgs := []string{"-y", "-i", "temp.gif", "-i", "palette.png"}
	paletteApplyArgs = append(paletteApplyArgs, "-lavfi", "paletteuse=dither=bayer:bayer_scale=3")
	paletteApplyArgs = append(paletteApplyArgs, "optimized.gif")

	// Show progress
	done := make(chan struct{})
	go func() {
		spinnerChars := []string{"|", "/", "-", "\\"}
		i := 0
		for {
			select {
			case <-done:
				fmt.Print("\r")
				return
			default:
				fmt.Printf("\rApplying optimization... %s", spinnerChars[i%len(spinnerChars)])
				i++
				time.Sleep(200 * time.Millisecond)
			}
		}
	}()

	cmd = exec.Command("ffmpeg", paletteApplyArgs...)
	if err := cmd.Run(); err != nil {
		close(done)
		log.Printf("Palette application failed: %v", err)
		// At least we have the basic GIF, rename it
		os.Rename("temp.gif", "optimized.gif")
	} else {
		close(done)
		// Clean up temp file
		os.Remove("temp.gif")
	}

	// Clean up palette file
	os.Remove("palette.png")

	// Get file size for user feedback
	if fileInfo, err := os.Stat("optimized.gif"); err == nil {
		sizeKB := fileInfo.Size() / 1024
		sizeMB := float64(sizeKB) / 1024
		fmt.Printf("\nOptimized GIF created successfully: optimized.gif (%.1f MB)\n", sizeMB)
	}
}

// GenerateGifWithTotalTime creates an animated GIF with a specific total duration
func GenerateGifWithTotalTime(totalTimeSeconds, transitionDuration int, fps int, scale float64) {
	// First, convert images optimized for GIF (1080p max height)
	if err := ConvertImagesForGif(1080); err != nil {
		log.Fatalf("Failed to convert images for GIF: %v", err)
	}

	// Find all GIF-optimized .jpg files
	files, err := filepath.Glob("gif_converted/*.jpg")
	if err != nil {
		log.Fatalf("Failed to list gif_converted .jpg files: %v", err)
	}

	if len(files) == 0 {
		log.Fatalf("No converted images found for GIF generation.")
	}

	fmt.Printf("Creating animated GIF with total time %d seconds from %d images...\n", totalTimeSeconds, len(files))

	// Calculate duration per frame in seconds
	durationPerFrame := float64(totalTimeSeconds) / float64(len(files))

	// Show progress
	done := make(chan struct{})
	go func() {
		spinnerChars := []string{"|", "/", "-", "\\"}
		i := 0
		for {
			select {
			case <-done:
				fmt.Print("\r")
				return
			default:
				fmt.Printf("\rGenerating GIF... %s", spinnerChars[i%len(spinnerChars)])
				i++
				time.Sleep(200 * time.Millisecond)
			}
		}
	}()

	// Build FFmpeg command with precise timing
	args := []string{"-y"}

	// Add all input files with calculated duration
	for _, file := range files {
		args = append(args, "-loop", "1", "-t", fmt.Sprintf("%.3f", durationPerFrame), "-i", file)
	}

	// Build filter complex
	filterComplex := ""

	// Apply scale if needed, otherwise just prepare videos
	for i := 0; i < len(files); i++ {
		if scale != 1.0 {
			filterComplex += fmt.Sprintf("[%d:v]scale=iw*%.2f:ih*%.2f,setsar=1[v%d];", i, scale, scale, i)
		} else {
			filterComplex += fmt.Sprintf("[%d:v]setsar=1[v%d];", i, i)
		}
	}

	// Concatenate all videos
	for i := 0; i < len(files); i++ {
		filterComplex += fmt.Sprintf("[v%d]", i)
	}
	filterComplex += fmt.Sprintf("concat=n=%d:v=1:a=0[out]", len(files))

	args = append(args, "-filter_complex", filterComplex)
	args = append(args, "-map", "[out]")
	args = append(args, "-r", fmt.Sprintf("%d", fps))
	args = append(args, "-t", fmt.Sprintf("%d", totalTimeSeconds)) // Force exact total duration
	args = append(args, "-f", "gif")
	args = append(args, "animated.gif")

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stderr = os.Stderr // Show FFmpeg output for debugging

	if err := cmd.Run(); err != nil {
		close(done)
		log.Fatalf("FFmpeg command failed: %v", err)
	}
	close(done)

	// Get file size for user feedback
	if fileInfo, err := os.Stat("animated.gif"); err == nil {
		sizeKB := fileInfo.Size() / 1024
		sizeMB := float64(sizeKB) / 1024
		fmt.Printf("\nGIF created successfully: animated.gif (%.1f MB)\n", sizeMB)
	}
}

// GenerateOptimizedGifWithTotalTime creates an optimized GIF with a specific total duration
func GenerateOptimizedGifWithTotalTime(totalTimeSeconds, transitionDuration int, fps int, scale float64) {
	// First, convert images optimized for GIF (1080p max height)
	if err := ConvertImagesForGif(1080); err != nil {
		log.Fatalf("Failed to convert images for GIF: %v", err)
	}

	// Find all GIF-optimized .jpg files
	files, err := filepath.Glob("gif_converted/*.jpg")
	if err != nil {
		log.Fatalf("Failed to list gif_converted .jpg files: %v", err)
	}

	if len(files) == 0 {
		log.Fatalf("No converted images found for GIF generation.")
	}

	fmt.Printf("Creating optimized animated GIF with total time %d seconds from %d images...\n", totalTimeSeconds, len(files))

	// Calculate duration per frame in seconds
	durationPerFrame := float64(totalTimeSeconds) / float64(len(files))

	// Step 1: Create palette from first image
	fmt.Println("Generating optimized palette...")
	paletteArgs := []string{"-y", "-i", files[0]}

	paletteFilter := "palettegen=max_colors=256"
	if scale != 1.0 {
		paletteFilter = fmt.Sprintf("scale=iw*%.2f:ih*%.2f,%s", scale, scale, paletteFilter)
	}

	paletteArgs = append(paletteArgs, "-vf", paletteFilter)
	paletteArgs = append(paletteArgs, "-t", "1")
	paletteArgs = append(paletteArgs, "palette.png")

	cmd := exec.Command("ffmpeg", paletteArgs...)
	if err := cmd.Run(); err != nil {
		log.Printf("Palette generation failed, falling back to regular GIF generation: %v", err)
		GenerateGifWithTotalTime(totalTimeSeconds, transitionDuration, fps, scale)
		return
	}

	// Step 2: Create basic GIF with exact timing
	fmt.Println("Creating optimized GIF...")

	tempGifArgs := []string{"-y"}

	// Add all input files with precise duration
	for _, file := range files {
		tempGifArgs = append(tempGifArgs, "-loop", "1", "-t", fmt.Sprintf("%.3f", durationPerFrame), "-i", file)
	}

	// Build filter complex
	filterComplex := ""
	for i := 0; i < len(files); i++ {
		if scale != 1.0 {
			filterComplex += fmt.Sprintf("[%d:v]scale=iw*%.2f:ih*%.2f,setsar=1[v%d];", i, scale, scale, i)
		} else {
			filterComplex += fmt.Sprintf("[%d:v]setsar=1[v%d];", i, i)
		}
	}

	for i := 0; i < len(files); i++ {
		filterComplex += fmt.Sprintf("[v%d]", i)
	}
	filterComplex += fmt.Sprintf("concat=n=%d:v=1:a=0[out]", len(files))

	tempGifArgs = append(tempGifArgs, "-filter_complex", filterComplex)
	tempGifArgs = append(tempGifArgs, "-map", "[out]")
	tempGifArgs = append(tempGifArgs, "-r", fmt.Sprintf("%d", fps))
	tempGifArgs = append(tempGifArgs, "-t", fmt.Sprintf("%d", totalTimeSeconds)) // Force exact duration
	tempGifArgs = append(tempGifArgs, "temp.gif")

	cmd = exec.Command("ffmpeg", tempGifArgs...)
	if err := cmd.Run(); err != nil {
		log.Printf("Basic GIF creation failed, cleaning up: %v", err)
		os.Remove("palette.png")
		return
	}

	// Step 3: Apply palette optimization
	fmt.Println("Applying palette optimization...")

	paletteApplyArgs := []string{"-y", "-i", "temp.gif", "-i", "palette.png"}
	paletteApplyArgs = append(paletteApplyArgs, "-lavfi", "paletteuse=dither=bayer:bayer_scale=3")
	paletteApplyArgs = append(paletteApplyArgs, "optimized.gif")

	// Show progress
	done := make(chan struct{})
	go func() {
		spinnerChars := []string{"|", "/", "-", "\\"}
		i := 0
		for {
			select {
			case <-done:
				fmt.Print("\r")
				return
			default:
				fmt.Printf("\rApplying optimization... %s", spinnerChars[i%len(spinnerChars)])
				i++
				time.Sleep(200 * time.Millisecond)
			}
		}
	}()

	cmd = exec.Command("ffmpeg", paletteApplyArgs...)
	if err := cmd.Run(); err != nil {
		close(done)
		log.Printf("Palette application failed: %v", err)
		os.Rename("temp.gif", "optimized.gif")
	} else {
		close(done)
		os.Remove("temp.gif")
	}

	// Clean up
	os.Remove("palette.png")

	// Get file size for user feedback
	if fileInfo, err := os.Stat("optimized.gif"); err == nil {
		sizeKB := fileInfo.Size() / 1024
		sizeMB := float64(sizeKB) / 1024
		fmt.Printf("\nOptimized GIF created successfully: optimized.gif (%.1f MB)\n", sizeMB)
	}
}
