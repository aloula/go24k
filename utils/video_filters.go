package utils

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
)

// formatSeconds ensures FFmpeg receives consistent decimal timing values.
func formatSeconds(seconds float64) string {
	if seconds < 0 {
		seconds = 0
	}
	return strconv.FormatFloat(seconds, 'f', 3, 64)
}

// supersampledResolution returns a 2x upscaled version of activeResolution.
func supersampledResolution() string {
	parts := strings.SplitN(activeResolution, "x", 2)
	if len(parts) != 2 {
		return activeResolution
	}
	w, err1 := strconv.Atoi(parts[0])
	h, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return activeResolution
	}
	return fmt.Sprintf("%dx%d", w*2, h*2)
}

// processImageFilter creates the video filter for a single image.
func processImageFilter(file string, index int, duration, fadeDuration float64, applyKenBurns, exifOverlay bool, fontSize int) string {
	var videoFilter string

	if applyKenBurns {
		superRes := supersampledResolution()
		superResScale := strings.Replace(superRes, "x", ":", 1)
		activeResScale := strings.Replace(activeResolution, "x", ":", 1)
		effect := getKenBurnsEffect(duration)
		if index == 0 {
			videoFilter = fmt.Sprintf("[0:v]scale=%s,%s,scale=%s,fade=t=in:st=0:d=%s,fps=%d,settb=AVTB,setsar=1,format=yuv420p", superResScale, effect, activeResScale, formatSeconds(fadeDuration), activeFPS)
		} else {
			videoFilter = fmt.Sprintf("[%d:v]scale=%s,%s,scale=%s,fps=%d,settb=AVTB,setsar=1,format=yuv420p", index, superResScale, effect, activeResScale, activeFPS)
		}
	} else {
		if index == 0 {
			videoFilter = fmt.Sprintf("[0:v]fps=%d,settb=AVTB,setsar=1,format=yuv420p,fade=t=in:st=0:d=%s", activeFPS, formatSeconds(fadeDuration))
		} else {
			videoFilter = fmt.Sprintf("[%d:v]fps=%d,settb=AVTB,setsar=1,format=yuv420p", index, activeFPS)
		}
	}

	if exifOverlay {
		originalFile := GetOriginalFilename(file)
		if originalFile != "" {
			if cameraInfo, err := ExtractCameraInfo(originalFile); err == nil && cameraInfo != nil {
				drawtextFilter := FormatCameraInfoOverlay(cameraInfo, fontSize, index)
				if drawtextFilter != "" {
					videoFilter += drawtextFilter
				}
			}
		}
	}

	return videoFilter
}

// buildCrossfadeFilters creates crossfade transitions for variable media segment lengths.
func buildCrossfadeFilters(segmentDurations []float64, fadeDuration float64) string {
	var filterComplex string
	numItems := len(segmentDurations)
	if numItems < 2 {
		return filterComplex
	}

	cumulative := 0.0
	for i := 0; i < numItems-1; i++ {
		cumulative += segmentDurations[i]
		next := i + 1
		offset := cumulative - (float64(i+1) * fadeDuration)
		if i == 0 {
			filterComplex += fmt.Sprintf("[v%d][v%d]xfade=transition=fade:duration=%s:offset=%s[x%d]; ", i, next, formatSeconds(fadeDuration), formatSeconds(offset), next)
		} else {
			filterComplex += fmt.Sprintf("[x%d][v%d]xfade=transition=fade:duration=%s:offset=%s[x%d]; ", i, next, formatSeconds(fadeDuration), formatSeconds(offset), next)
		}
	}

	return filterComplex
}

func calculateFinalLength(segmentDurations []float64, fadeDuration float64) float64 {
	total := 0.0
	for _, d := range segmentDurations {
		total += d
	}
	if len(segmentDurations) < 2 {
		return total
	}
	return total - (float64(len(segmentDurations)-1) * fadeDuration)
}

// buildFinalFilters creates the fade-out and trim filters.
func buildFinalFilters(segmentDurations []float64, fadeDuration float64) (string, float64) {
	numItems := len(segmentDurations)
	finalLength := calculateFinalLength(segmentDurations, fadeDuration)
	fadeOutStart := finalLength - fadeDuration

	var filterComplex string
	inputLabel := "v0"
	if numItems > 1 {
		inputLabel = fmt.Sprintf("x%d", numItems-1)
	}
	filterComplex += fmt.Sprintf("[%s]trim=duration=%s,setpts=PTS-STARTPTS[xt]; ", inputLabel, formatSeconds(finalLength))
	filterComplex += fmt.Sprintf("[xt]fade=t=out:st=%s:d=%s[xfout]; ", formatSeconds(fadeOutStart), formatSeconds(fadeDuration))

	return filterComplex, finalLength
}

// processVideoFilter creates a normalized filter for a video input.
func processVideoFilter(index int, fadeDuration float64) string {
	res := activeResolution
	parts := strings.SplitN(res, "x", 2)
	w, h := parts[0], parts[1]
	base := fmt.Sprintf("[%d:v]fps=%d,settb=AVTB,scale='if(gt(iw,%s)+gt(ih,%s),%s,iw)':'if(gt(iw,%s)+gt(ih,%s),%s,ih)':force_original_aspect_ratio=decrease,pad=%s:%s:(ow-iw)/2:(oh-ih)/2:black,setsar=1,format=yuv420p,setpts=PTS-STARTPTS", index, activeFPS, w, h, w, w, h, h, w, h)
	if index == 0 {
		return fmt.Sprintf("%s,fade=t=in:st=0:d=%s", base, formatSeconds(fadeDuration))
	}
	return base
}

func normalizeKenBurnsMode(mode string) string {
	mode = strings.TrimSpace(strings.ToLower(mode))
	switch mode {
	case kenBurnsModeLow:
		return kenBurnsModeLow
	case kenBurnsModeMedium:
		return kenBurnsModeMedium
	case kenBurnsModeHigh:
		return kenBurnsModeHigh
	case "subtle":
		return kenBurnsModeLow
	case "cinematic":
		return kenBurnsModeMedium
	case "dynamic":
		return kenBurnsModeHigh
	default:
		return kenBurnsModeHigh
	}
}

// getKenBurnsEffect generates a Ken Burns effect using a fixed zoompan expression.
func getKenBurnsEffect(duration float64) string {
	totalFrames := int(math.Round(duration * float64(activeFPS)))
	if totalFrames < 1 {
		totalFrames = 1
	}

	mode := normalizeKenBurnsMode(activeKenBurnsMode)

	startZoom := 1.00
	endZoom := 1.07
	offsetX := "126"
	offsetY := "70"

	if mode == kenBurnsModeLow {
		endZoom = 1.04
		offsetX = "98"
		offsetY = "56"
	} else if mode == kenBurnsModeHigh {
		endZoom = 1.10
		offsetX = "154"
		offsetY = "84"
	}

	if activeResolution == resolutionFullHD {
		if mode == kenBurnsModeLow {
			endZoom = 1.03
			offsetX = "70"
			offsetY = "40"
		} else if mode == kenBurnsModeMedium {
			endZoom = 1.05
			offsetX = "84"
			offsetY = "48"
		} else {
			endZoom = 1.07
			offsetX = "98"
			offsetY = "56"
		}
	}

	superRes := supersampledResolution()
	zoomStep := (endZoom - startZoom) / float64(totalFrames)
	if zoomStep < 0.0001 {
		zoomStep = 0.0001
	}
	zoomStepStr := strconv.FormatFloat(zoomStep, 'f', 6, 64)
	endZoomStr := strconv.FormatFloat(endZoom, 'f', 2, 64)

	buildExpr := func(xOffset, yOffset string) string {
		return fmt.Sprintf(
			"zoompan=z='min(zoom+%s,%s)':x='iw/2-(iw/zoom/2)%s':y='ih/2-(ih/zoom/2)%s':d=%d:s=%s",
			zoomStepStr,
			endZoomStr,
			xOffset,
			yOffset,
			totalFrames,
			superRes,
		)
	}

	variants := []string{
		buildExpr("+"+offsetX, ""),
		buildExpr("-"+offsetX, ""),
		buildExpr("", "+"+offsetY),
		buildExpr("", "-"+offsetY),
		buildExpr("+"+offsetX, "+"+offsetY),
		buildExpr("-"+offsetX, "+"+offsetY),
		buildExpr("+"+offsetX, "-"+offsetY),
		buildExpr("-"+offsetX, "-"+offsetY),
	}

	return variants[rand.Intn(len(variants))]
}
