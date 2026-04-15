package utils

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// AudioConfig contains audio processing configuration
type AudioConfig struct {
	Inputs             []string
	MapArgs            []string
	AudioFilter        string
	HasAudio           bool
	AudioBitrateSource string
}

// findMusicFiles returns the list of mp3 files without logging
func findMusicFiles() ([]string, error) {
	musicFiles, err := filepath.Glob("*.mp3")
	if err != nil {
		return nil, fmt.Errorf("failed to list mp3 files: %v", err)
	}
	sort.Strings(musicFiles)
	return musicFiles, nil
}

func createAudioConcatFile(musicFiles []string) (string, error) {
	if len(musicFiles) == 0 {
		return "", fmt.Errorf("no music files provided")
	}
	if len(musicFiles) == 1 {
		return musicFiles[0], nil
	}

	concatFile := "audio_concat.txt"
	var content strings.Builder
	for _, file := range musicFiles {
		escapedFile := strings.ReplaceAll(file, "'", "'\\''")
		content.WriteString(fmt.Sprintf("file '%s'\n", escapedFile))
	}

	if err := os.WriteFile(concatFile, []byte(content.String()), 0644); err != nil {
		return "", fmt.Errorf("failed to create concat file: %v", err)
	}

	return concatFile, nil
}

func getTotalAudioDurationSeconds(musicFiles []string) (float64, error) {
	if len(musicFiles) == 0 {
		return 0, fmt.Errorf("no music files provided")
	}

	totalDuration := 0.0
	for _, file := range musicFiles {
		dur, err := getAudioDurationSeconds(file)
		if err != nil {
			return 0, err
		}
		totalDuration += dur
	}

	return totalDuration, nil
}

func getAudioBitrateStr(filename string) string {
	cmd := newExecCommand("ffprobe", "-v", "error", "-select_streams", "a:0",
		"-show_entries", "stream=bit_rate", "-of", "default=noprint_wrappers=1:nokey=1", filename)
	out, err := cmd.Output()
	if err != nil {
		return "192k"
	}
	bitrateStr := strings.TrimSpace(string(out))
	if bitrateStr == "" || bitrateStr == "N/A" {
		cmd2 := newExecCommand("ffprobe", "-v", "error",
			"-show_entries", "format=bit_rate", "-of", "default=noprint_wrappers=1:nokey=1", filename)
		out2, err2 := cmd2.Output()
		if err2 != nil {
			return "192k"
		}
		bitrateStr = strings.TrimSpace(string(out2))
	}
	bps, err := strconv.Atoi(bitrateStr)
	if err != nil || bps <= 0 {
		return "192k"
	}
	return fmt.Sprintf("%dk", bps/1000)
}

func getMediaDurationSeconds(filename string) (float64, error) {
	cmd := newExecCommand("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", filename)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe duration failed: %v", err)
	}
	durStr := strings.TrimSpace(string(output))
	if durStr == "" {
		return 0, fmt.Errorf("empty duration from ffprobe")
	}
	dur, err := strconv.ParseFloat(durStr, 64)
	if err != nil {
		return 0, fmt.Errorf("parse duration: %w", err)
	}
	return dur, nil
}

func getAudioDurationSeconds(filename string) (float64, error) {
	return getMediaDurationSeconds(filename)
}

func hasAudioStream(filename string) (bool, error) {
	cmd := newExecCommand("ffprobe", "-v", "error", "-select_streams", "a", "-show_entries", "stream=index", "-of", "csv=p=0", filename)
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("ffprobe audio stream failed: %v", err)
	}
	return strings.TrimSpace(string(output)) != "", nil
}

func buildTimelineOffsets(mediaInputs []MediaInput, fadeDuration float64) []float64 {
	offsets := make([]float64, len(mediaInputs))
	currentOffset := 0.0

	for index := 1; index < len(mediaInputs); index++ {
		currentOffset += mediaInputs[index-1].SegmentDuration - fadeDuration
		offsets[index] = currentOffset
	}

	return offsets
}

func clipAudioFadeDuration(segmentDuration, fadeDuration float64) float64 {
	if segmentDuration <= 0 {
		return 0
	}

	maxFade := segmentDuration / 4
	if maxFade < 0.25 {
		maxFade = 0.25
	}

	if fadeDuration < maxFade {
		return fadeDuration
	}

	return maxFade
}

func joinFilterInputs(labels []string) string {
	var builder strings.Builder
	for _, label := range labels {
		builder.WriteString("[")
		builder.WriteString(label)
		builder.WriteString("]")
	}
	return builder.String()
}

func buildMusicMuteExpression(mediaInputs []MediaInput, offsets []float64, fadeDuration float64) string {
	var parts []string

	for i, media := range mediaInputs {
		if media.IsImage || !media.HasAudio {
			continue
		}

		fadeLen := math.Max(clipAudioFadeDuration(media.SegmentDuration, fadeDuration), 0.001)
		clipStart := offsets[i]
		clipEnd := clipStart + media.SegmentDuration
		muteStart := math.Max(clipStart-fadeLen, 0)
		muteEnd := clipEnd + fadeLen

		expr := fmt.Sprintf(
			"if(lt(t,%.3f),1,if(lt(t,%.3f),(%.3f-t)/%.3f,if(lt(t,%.3f),0,if(lt(t,%.3f),(t-%.3f)/%.3f,1))))",
			muteStart,
			clipStart, clipStart, fadeLen,
			clipEnd,
			muteEnd, clipEnd, fadeLen,
		)
		parts = append(parts, expr)
	}

	if len(parts) == 0 {
		return "1"
	}
	if len(parts) == 1 {
		return parts[0]
	}
	result := parts[0]
	for _, p := range parts[1:] {
		result = "min(" + result + "," + p + ")"
	}
	return result
}

func adjustDurationsToMusic(duration, fadeDuration float64, numImages int, audioSeconds float64) (float64, float64, bool) {
	if audioSeconds <= 0 || numImages < 2 {
		return duration, fadeDuration, false
	}

	baseTotal := (float64(numImages) * duration) - (float64(numImages-1) * fadeDuration)
	if baseTotal <= 0 {
		return duration, fadeDuration, false
	}

	scale := audioSeconds / baseTotal
	newDuration := duration * scale
	newFade := fadeDuration * scale

	if newDuration < 1 {
		newDuration = 1
	}
	if newFade < 1 {
		newFade = 1
	}
	if newFade >= newDuration {
		newFade = math.Max(1, newDuration*0.5)
		if newFade >= newDuration {
			newDuration = newFade + 0.1
		}
	}

	newTotal := (float64(numImages) * newDuration) - (float64(numImages-1) * newFade)
	diff := audioSeconds - newTotal
	if math.Abs(diff) >= 0.01 {
		newDuration += diff / float64(numImages)
		if newDuration < 1 {
			newDuration = 1
		}
	}

	if newFade < 1 {
		newFade = 1
	}
	if newFade >= newDuration {
		newFade = math.Max(1, newDuration*0.5)
	}

	return newDuration, newFade, true
}

func setupAudioProcessing(inputs []string, mediaInputs []MediaInput, finalLength, fadeDuration float64, musicFiles []string, keepVideoAudio bool) AudioConfig {
	config := AudioConfig{Inputs: inputs}

	hasMusic := len(musicFiles) > 0
	musicInputIndex := len(mediaInputs)

	if hasMusic {
		if len(musicFiles) > 1 {
			fmt.Printf("Audio files found: %d MP3 files\n", len(musicFiles))
			for _, file := range musicFiles {
				fmt.Printf("  - %s\n", file)
			}

			concatFile, err := createAudioConcatFile(musicFiles)
			if err != nil {
				fmt.Printf("Warning: Failed to create audio concat: %v\n", err)
				fmt.Printf("Using single audio file: %s\n", musicFiles[0])
				config.Inputs = append(config.Inputs, "-i", musicFiles[0])
			} else {
				config.Inputs = append(config.Inputs, "-f", "concat", "-safe", "0", "-i", concatFile)
			}
		} else {
			fmt.Printf("Audio file found: %s\n", musicFiles[0])
			config.Inputs = append(config.Inputs, "-i", musicFiles[0])
		}
		config.AudioBitrateSource = musicFiles[0]
	}

	offsets := buildTimelineOffsets(mediaInputs, fadeDuration)
	videoAudioLabels := []string{}

	if keepVideoAudio {
		for index, media := range mediaInputs {
			if media.IsImage || !media.HasAudio {
				continue
			}

			fadeLength := clipAudioFadeDuration(media.SegmentDuration, fadeDuration)
			fadeOutStart := media.SegmentDuration - fadeLength
			delayMs := int(math.Round(offsets[index] * 1000))
			label := fmt.Sprintf("clipaudio%d", len(videoAudioLabels))

			config.AudioFilter += fmt.Sprintf("[%d:a]aformat=sample_fmts=fltp:sample_rates=48000:channel_layouts=stereo,aresample=48000,atrim=duration=%s,asetpts=PTS-STARTPTS,afade=t=in:st=0:d=%s,afade=t=out:st=%s:d=%s,adelay=%d|%d[%s]; ", index, formatSeconds(media.SegmentDuration), formatSeconds(fadeLength), formatSeconds(fadeOutStart), formatSeconds(fadeLength), delayMs, delayMs, label)
			videoAudioLabels = append(videoAudioLabels, label)

			if config.AudioBitrateSource == "" {
				config.AudioBitrateSource = media.Path
			}
		}

		if len(videoAudioLabels) > 0 {
			fmt.Printf("Keeping audio from %d input video(s)\n", len(videoAudioLabels))
		} else {
			fmt.Printf("keep-video-audio requested, but no input videos with audio were found\n")
		}
	}

	var clipAudioBusLabel string
	if len(videoAudioLabels) == 1 {
		clipAudioBusLabel = videoAudioLabels[0]
	} else if len(videoAudioLabels) > 1 {
		clipAudioBusLabel = "clipaudiobus"
		config.AudioFilter += fmt.Sprintf("%samix=inputs=%d:duration=longest:normalize=0:dropout_transition=%s[%s]; ", joinFilterInputs(videoAudioLabels), len(videoAudioLabels), formatSeconds(fadeDuration), clipAudioBusLabel)
	}

	var finalAudioLabel string
	if hasMusic {
		musicFadeOutStart := finalLength - fadeDuration
		if musicFadeOutStart < 0 {
			musicFadeOutStart = 0
		}

		config.AudioFilter += fmt.Sprintf("[%d:a]aformat=sample_fmts=fltp:sample_rates=48000:channel_layouts=stereo,aresample=48000,loudnorm=I=-16:TP=-1.5:LRA=11,atrim=duration=%s,asetpts=PTS-STARTPTS,afade=t=in:st=0:d=%s,afade=t=out:st=%s:d=%s[musicout]; ", musicInputIndex, formatSeconds(finalLength), formatSeconds(fadeDuration), formatSeconds(musicFadeOutStart), formatSeconds(fadeDuration))

		if clipAudioBusLabel != "" {
			muteExpr := buildMusicMuteExpression(mediaInputs, offsets, fadeDuration)
			config.AudioFilter += fmt.Sprintf("[musicout]volume='%s':eval=frame[musicmuted]; ", muteExpr)
			config.AudioFilter += fmt.Sprintf("[musicmuted][%s]amix=inputs=2:duration=first:normalize=0:dropout_transition=%s[mixedaudio]; ", clipAudioBusLabel, formatSeconds(fadeDuration))
			finalAudioLabel = "mixedaudio"
		} else {
			finalAudioLabel = "musicout"
		}
	} else if clipAudioBusLabel != "" {
		finalAudioLabel = clipAudioBusLabel
	}

	config.HasAudio = finalAudioLabel != ""
	if config.HasAudio {
		if !hasMusic && clipAudioBusLabel != "" {
			fmt.Printf("No MP3 file found - using input video audio only\n")
		}
		config.MapArgs = []string{"-map", "[xfout]", "-map", fmt.Sprintf("[%s]", finalAudioLabel), "-shortest"}
	} else {
		fmt.Printf("No MP3 file found - generating video without audio\n")
		config.MapArgs = []string{"-map", "[xfout]"}
	}

	return config
}
