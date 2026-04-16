package utils

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// MediaInput represents an item (image or video) to be included in the timeline.
type MediaInput struct {
	Path            string
	IsImage         bool
	HasAudio        bool
	SegmentDuration float64
	CapturedAt      time.Time
	HasCapturedAt   bool
	SortName        string
}

// findVideoFiles returns video files in the current directory based on selected options.
func findVideoFiles(includeVideos bool) ([]string, error) {
	if !includeVideos {
		return []string{}, nil
	}

	allowedExtensions := map[string]struct{}{}
	for _, ext := range []string{".mp4", ".mov", ".mkv", ".avi", ".webm", ".m4v"} {
		allowedExtensions[ext] = struct{}{}
	}

	generatedOutputs := generatedOutputVideoNames()
	var files []string

	directoryEntries, err := os.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("failed to read current directory: %v", err)
	}

	for _, entry := range directoryEntries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if _, isGeneratedOutput := generatedOutputs[strings.ToLower(filepath.Base(name))]; isGeneratedOutput {
			continue
		}

		ext := strings.ToLower(filepath.Ext(name))
		if _, isAllowed := allowedExtensions[ext]; !isAllowed {
			continue
		}

		files = append(files, name)
	}

	sort.Strings(files)
	return files, nil
}

func generatedOutputVideoNames() map[string]struct{} {
	return map[string]struct{}{
		strings.ToLower(outputVideoLegacy): {},
		strings.ToLower(outputVideoUHD):    {},
		strings.ToLower(outputVideoFHD):    {},
	}
}

func outputVideoFilename() string {
	if activeResolution == resolutionFullHD {
		return outputVideoFHD
	}
	return outputVideoUHD
}

// collectMediaInputs builds a sorted timeline from converted images and optional videos.
// Default ordering is capture metadata time, with filename as deterministic fallback.
// If orderByFilename is true, ordering uses filenames only.
// If randomOrder is true, timeline entries are shuffled randomly.
func collectMediaInputs(imageDuration float64, includeVideos, orderByFilename, randomOrder bool) ([]MediaInput, error) {
	imageFiles, err := filepath.Glob("converted/*.jpg")
	if err != nil {
		return nil, fmt.Errorf("failed to list converted images: %v", err)
	}
	sort.Strings(imageFiles)

	var media []MediaInput
	for _, file := range imageFiles {
		capturedAt, hasCapturedAt := extractImageTimestampFromConvertedName(file)
		if !hasCapturedAt {
			capturedAt, hasCapturedAt = extractCaptureTimeFromFilename(file)
		}
		sortName := mediaSortName(file)
		if orderByFilename {
			sortName = resolveImageSortName(file)
		}
		media = append(media, MediaInput{
			Path:            file,
			IsImage:         true,
			SegmentDuration: imageDuration,
			CapturedAt:      capturedAt,
			HasCapturedAt:   hasCapturedAt,
			SortName:        sortName,
		})
	}

	if includeVideos {
		videoFiles, err := findVideoFiles(includeVideos)
		if err != nil {
			return nil, err
		}

		for _, file := range videoFiles {
			duration, err := getMediaDurationSeconds(file)
			if err != nil {
				return nil, fmt.Errorf("failed to read video duration for %s: %v", file, err)
			}
			capturedAt, hasCapturedAt := getVideoCaptureTime(file)
			if !hasCapturedAt {
				capturedAt, hasCapturedAt = extractCaptureTimeFromFilename(file)
			}
			hasAudio, err := hasAudioStream(file)
			if err != nil {
				return nil, fmt.Errorf("failed to inspect audio stream for %s: %v", file, err)
			}
			if duration <= 0 {
				return nil, fmt.Errorf("video %s has invalid duration %.2f", file, duration)
			}

			videoSortName := mediaSortName(file)
			media = append(media, MediaInput{
				Path:            file,
				IsImage:         false,
				HasAudio:        hasAudio,
				SegmentDuration: duration,
				CapturedAt:      capturedAt,
				HasCapturedAt:   hasCapturedAt,
				SortName:        videoSortName,
			})
		}
	}

	if randomOrder {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		rng.Shuffle(len(media), func(i, j int) {
			media[i], media[j] = media[j], media[i]
		})
	} else {
		sort.Slice(media, func(i, j int) bool {
			if orderByFilename {
				return media[i].SortName < media[j].SortName
			}

			if media[i].HasCapturedAt && media[j].HasCapturedAt {
				if !media[i].CapturedAt.Equal(media[j].CapturedAt) {
					return media[i].CapturedAt.Before(media[j].CapturedAt)
				}
			}
			return media[i].SortName < media[j].SortName
		})
	}

	if len(media) == 0 {
		if includeVideos {
			return nil, fmt.Errorf("no converted images or supported videos found")
		}
		return nil, fmt.Errorf("no converted images found in 'converted/' directory.\nPlease convert your images first using the image conversion feature")
	}

	if len(media) < 2 {
		return nil, fmt.Errorf("not enough media found. Need at least 2 images/videos to create a video with transitions.\nFound: %d item(s)", len(media))
	}

	return media, nil
}

func mediaSortName(path string) string {
	name := strings.TrimSpace(path)
	if name == "" {
		return ""
	}
	return strings.ToLower(filepath.Base(name))
}

// resolveImageSortName returns the preferred sort key for converted images.
// In filename-order mode we sort by original source filename when possible.
func resolveImageSortName(convertedPath string) string {
	original := GetOriginalFilename(convertedPath)
	if original != "" {
		return mediaSortName(original)
	}
	return mediaSortName(convertedPath)
}

func extractImageTimestampFromConvertedName(path string) (time.Time, bool) {
	base := filepath.Base(path)
	base = trimConvertedImageResolutionSuffix(base)

	if len(base) >= len("20060102_150405") {
		candidate := base[:len("20060102_150405")]
		if ts, err := time.Parse("20060102_150405", candidate); err == nil {
			return ts, true
		}
	}

	return time.Time{}, false
}

func extractCaptureTimeFromFilename(path string) (time.Time, bool) {
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	if name == "" {
		return time.Time{}, false
	}

	if ts, ok := parseTimestampCandidate(name); ok {
		return ts, true
	}

	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(\d{8}[_-]\d{6})`),
		regexp.MustCompile(`(\d{14})`),
		regexp.MustCompile(`(\d{4}[-_]\d{2}[-_]\d{2}[T _-]\d{2}[:._-]\d{2}[:._-]\d{2})`),
	}

	for _, pattern := range patterns {
		match := pattern.FindString(name)
		if match == "" {
			continue
		}
		if ts, ok := parseTimestampCandidate(match); ok {
			return ts, true
		}
	}

	return time.Time{}, false
}

func parseTimestampCandidate(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}

	if ts, err := time.Parse("20060102_150405", value); err == nil {
		return ts, true
	}
	if ts, err := time.Parse("20060102-150405", value); err == nil {
		return ts, true
	}
	if ts, err := time.Parse("20060102150405", value); err == nil {
		return ts, true
	}

	normalized := strings.NewReplacer("_", "-", ".", ":").Replace(value)
	layouts := []string{
		"2006-01-02-15-04-05",
		"2006-01-02-15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if ts, err := time.Parse(layout, normalized); err == nil {
			return ts, true
		}
	}

	return time.Time{}, false
}

func parseVideoCreationTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("empty timestamp")
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05Z07:00",
	}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported timestamp format: %s", value)
}

func getVideoCaptureTime(filename string) (time.Time, bool) {
	cmd := newExecCommand("ffprobe", "-v", "error",
		"-show_entries", "format_tags=creation_time:stream_tags=creation_time",
		"-of", "default=noprint_wrappers=1:nokey=1", filename)
	output, err := cmd.Output()
	if err != nil {
		return time.Time{}, false
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if ts, parseErr := parseVideoCreationTime(line); parseErr == nil {
			return ts, true
		}
	}

	return time.Time{}, false
}
