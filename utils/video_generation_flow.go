package utils

import "fmt"

func validateMediaInputs(mediaInputs []MediaInput, fadeSec float64) (int, int, error) {
	imageCount := 0
	videoCount := 0

	for _, media := range mediaInputs {
		if media.IsImage {
			imageCount++
		} else {
			videoCount++
		}
		if media.SegmentDuration <= fadeSec {
			return 0, 0, fmt.Errorf("media item %s has duration %.2fs which must be greater than transition %.2fs", media.Path, media.SegmentDuration, fadeSec)
		}
	}

	return imageCount, videoCount, nil
}

func setImageDurations(mediaInputs []MediaInput, durationSec float64) {
	for i := range mediaInputs {
		if mediaInputs[i].IsImage {
			mediaInputs[i].SegmentDuration = durationSec
		}
	}
}

func minimumAudioLength(itemCount int) float64 {
	return float64(itemCount*5 - (itemCount - 1))
}

func applyFitAudioSettings(mediaInputs []MediaInput, durationSec, fadeSec float64, fitAudio bool, musicFiles []string, videoCount int) (float64, float64, error) {
	if !fitAudio {
		return durationSec, fadeSec, nil
	}

	if videoCount > 0 {
		fmt.Printf("fit-audio with mixed images/videos keeps original video lengths and uses provided image/transition durations.\n")
	}
	if len(musicFiles) == 0 {
		fmt.Printf("fit-audio requested but no MP3 found; using provided durations.\n")
		return durationSec, fadeSec, nil
	}

	applyAdjustedDurations := func(audioSeconds float64, label string) (float64, float64, error) {
		if videoCount > 0 {
			fmt.Printf("fit-audio skipped in mixed media mode; keeping source video durations.\n")
			return durationSec, fadeSec, nil
		}

		minLength := minimumAudioLength(len(mediaInputs))
		if audioSeconds < minLength {
			return 0, 0, fmt.Errorf("Audio duration (%.1fs) is too short for %d images.\nMinimum required: %.1fs (5s per image × %d - 1s transitions × %d)\nPlease use fewer images or add more audio files.", audioSeconds, len(mediaInputs), minLength, len(mediaInputs), len(mediaInputs)-1)
		}

		oldDuration, oldFade := durationSec, fadeSec
		durationSec, fadeSec, _ = adjustDurationsToMusic(durationSec, fadeSec, len(mediaInputs), audioSeconds)
		fmt.Printf("Auto-fit to music (%s): duration %.2fs → %.2fs, transition %.2fs → %.2fs\n", label, oldDuration, durationSec, oldFade, fadeSec)
		setImageDurations(mediaInputs, durationSec)
		return durationSec, fadeSec, nil
	}

	if len(musicFiles) > 1 {
		audioSeconds, err := getTotalAudioDurationSeconds(musicFiles)
		if err != nil || audioSeconds <= 0 {
			fmt.Printf("fit-audio requested but could not read music duration; using provided durations.\n")
			return durationSec, fadeSec, nil
		}
		return applyAdjustedDurations(audioSeconds, fmt.Sprintf("%.1fs total", audioSeconds))
	}

	audioSeconds, err := getAudioDurationSeconds(musicFiles[0])
	if err != nil || audioSeconds <= 0 {
		fmt.Printf("fit-audio requested but could not read music duration; using provided durations.\n")
		return durationSec, fadeSec, nil
	}
	return applyAdjustedDurations(audioSeconds, fmt.Sprintf("%.1fs", audioSeconds))
}

func buildVideoFilterGraph(mediaInputs []MediaInput, fadeSec float64, applyKenBurns, exifOverlay bool, fontSize int) ([]string, string, float64) {
	inputs := []string{}
	filterComplex := ""
	segmentDurations := make([]float64, 0, len(mediaInputs))

	for index, media := range mediaInputs {
		var videoFilter string
		if media.IsImage {
			inputs = append(inputs, "-loop", "1", "-t", formatSeconds(media.SegmentDuration), "-i", media.Path)
			videoFilter = processImageFilter(media.Path, index, media.SegmentDuration, fadeSec, applyKenBurns, exifOverlay, fontSize)
		} else {
			inputs = append(inputs, "-i", media.Path)
			videoFilter = processVideoFilter(index, fadeSec)
		}
		segmentDurations = append(segmentDurations, media.SegmentDuration)
		filterComplex += fmt.Sprintf("%s[v%d]; ", videoFilter, index)
	}

	filterComplex += buildCrossfadeFilters(segmentDurations, fadeSec)
	finalFilters, finalLength := buildFinalFilters(segmentDurations, fadeSec)
	filterComplex += finalFilters

	return inputs, filterComplex, finalLength
}
