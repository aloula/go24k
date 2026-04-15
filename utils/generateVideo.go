package utils

import (
	"fmt"
	"log"
	"os"
)

const (
	linuxOS            = "linux"
	resolution4K       = "3840x2160"
	resolutionFullHD   = "1920x1080"
	outputVideoLegacy  = "video.mp4"
	outputVideoUHD     = "video_uhd.mp4"
	outputVideoFHD     = "video_fhd.mp4"
	kenBurnsModeLow    = "low"
	kenBurnsModeMedium = "medium"
	kenBurnsModeHigh   = "high"
)

// activeResolution holds the target output resolution for the current run.
// Set at the start of GenerateVideo() based on the fullHD flag.
var activeResolution = resolution4K

// activeFPS holds the target output framerate for the current run.
// Set at the start of GenerateVideo().
var activeFPS = 30

// activeKenBurnsMode holds the motion profile for Ken Burns when enabled.
// Set at the start of GenerateVideo().
var activeKenBurnsMode = kenBurnsModeHigh

// GenerateVideo creates a video from converted images with crossfade transitions,
// audio fades, and optionally a Ken Burns effect applied to each image.
// If applyKenBurns is false, the images remain static.
// If exifOverlay is true, camera info will be displayed in the footer with specified fontSize.
// If fullHD is true, the output resolution will be Full HD (1920x1080) instead of 4K UHD (3840x2160).
func GenerateVideo(duration, fadeDuration int, applyKenBurns, exifOverlay bool, fontSize int, fitAudio, includeVideos, includeMOV, keepVideoAudio, fullHD bool, fps int, orderByFilename, randomOrder bool, kenBurnsMode string) {
	// Set active resolution based on the fullHD flag.
	if fullHD {
		activeResolution = resolutionFullHD
	} else {
		activeResolution = resolution4K
	}

	if fps != 60 {
		fps = 30
	}
	activeFPS = fps
	activeKenBurnsMode = normalizeKenBurnsMode(kenBurnsMode)
	outputFilename := outputVideoFilename()

	durationSec := float64(duration)
	fadeSec := float64(fadeDuration)

	mediaInputs, err := collectMediaInputs(durationSec, includeVideos, includeMOV, orderByFilename, randomOrder)
	if err != nil {
		log.Fatalf("%v", err)
	}

	if randomOrder && orderByFilename {
		fmt.Printf("Both -random-order and -order-by-filename were set; using random order.\n")
	}

	if fadeSec <= 0 {
		log.Fatalf("transition duration must be greater than 0")
	}

	imageCount, videoCount, err := validateMediaInputs(mediaInputs, fadeSec)
	if err != nil {
		log.Fatalf("%v", err)
	}

	fmt.Printf("Generating video from %d media items (%d images, %d videos)...\n", len(mediaInputs), imageCount, videoCount)
	if applyKenBurns {
		fmt.Printf("Ken Burns mode: %s\n", activeKenBurnsMode)
	} else {
		fmt.Printf("Ken Burns mode: disabled (-static)\n")
	}

	// Detect music files once
	musicFiles, err := findMusicFiles()
	if err != nil {
		log.Fatalf("%v", err)
	}

	durationSec, fadeSec, err = applyFitAudioSettings(mediaInputs, durationSec, fadeSec, fitAudio, musicFiles, videoCount)
	if err != nil {
		log.Fatalf("%v", err)
	}

	inputs, filterComplex, finalLength := buildVideoFilterGraph(mediaInputs, fadeSec, applyKenBurns, exifOverlay, fontSize)

	// Setup audio processing
	totalDuration := finalLength
	audioConfig := setupAudioProcessing(inputs, mediaInputs, totalDuration, fadeSec, musicFiles, keepVideoAudio)

	// Add audio filter to filter complex if audio is present
	if audioConfig.HasAudio {
		filterComplex += audioConfig.AudioFilter
	}

	// Write filter complex to a file to avoid Windows command line length limits
	filterComplexFile := "filter_complex.txt"
	if err := os.WriteFile(filterComplexFile, []byte(filterComplex), 0644); err != nil {
		log.Fatalf("Failed to write filter complex file: %v", err)
	}
	defer os.Remove(filterComplexFile)

	// Build complete FFmpeg command
	args := []string{"-y"}
	args = append(args, audioConfig.Inputs...)
	args = append(args, "-filter_complex_script", filterComplexFile)
	args = append(args, audioConfig.MapArgs...)
	args = append(args, getOptimalVideoSettings()...)

	// Add audio encoding settings if audio is present, preserving input bitrate
	if audioConfig.HasAudio {
		audioBitrateSource := audioConfig.AudioBitrateSource
		if audioBitrateSource == "" && len(musicFiles) > 0 {
			audioBitrateSource = musicFiles[0]
		}
		audioBitrate := "192k"
		if audioBitrateSource != "" {
			audioBitrate = getAudioBitrateStr(audioBitrateSource)
		}
		args = append(args, "-c:a", "aac", "-b:a", audioBitrate)
	}

	args = append(args, "-t", formatSeconds(finalLength))
	args = append(args, outputFilename)

	// Execute FFmpeg command
	if err := runFFmpegCommand(args, audioConfig.HasAudio); err != nil {
		log.Fatalf("Video generation failed: %v", err)
	}

	// Clean up concat file if it was created
	if len(musicFiles) > 1 {
		os.Remove("audio_concat.txt")
	}

	// Display final information
	displayVideoInfo(outputFilename, finalLength)
}
