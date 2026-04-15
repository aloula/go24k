package utils

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// VideoInfo contains technical details about a video file.
type VideoInfo struct {
	FileSizeMB   float64
	DurationSec  float64
	VideoBitrate string
	AudioBitrate string
	Framerate    string
	Resolution   string
}

func getFileSize(filename string) float64 {
	if fileInfo, err := os.Stat(filename); err == nil {
		return float64(fileInfo.Size()) / (1024 * 1024)
	}
	return 0
}

func runFFProbe(filename string) (string, error) {
	cmd := newExecCommand("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filename)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("ffprobe failed: %v", err)
	}
	return string(output), nil
}

func getVideoDetails(filename string) (*VideoInfo, error) {
	info := &VideoInfo{}
	info.FileSizeMB = getFileSize(filename)
	if duration, err := getMediaDurationSeconds(filename); err == nil {
		info.DurationSec = duration
	}

	outputStr, err := runFFProbe(filename)
	if err != nil {
		info.Framerate = fmt.Sprintf("%d fps", activeFPS)
		info.Resolution = activeResolution
		info.AudioBitrate = "No audio"
		return info, err
	}

	info.VideoBitrate, info.Framerate, info.Resolution = extractVideoInfo(outputStr)
	info.AudioBitrate = extractAudioInfo(outputStr)

	if info.Framerate == "" {
		info.Framerate = fmt.Sprintf("%d fps", activeFPS)
	}
	if info.Resolution == "" {
		info.Resolution = activeResolution
	}
	if info.AudioBitrate == "" {
		info.AudioBitrate = "No audio"
	}

	return info, nil
}

func extractDuration(outputStr string) float64 {
	if !strings.Contains(outputStr, `"duration"`) {
		return 0
	}

	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, `"duration"`) && strings.Contains(line, `"format"`) {
			parts := strings.Split(line, `"`)
			for i, part := range parts {
				if part == "duration" && i+2 < len(parts) {
					if duration, err := strconv.ParseFloat(parts[i+2], 64); err == nil {
						return duration
					}
				}
			}
		}
	}
	return 0
}

func extractVideoInfo(outputStr string) (bitrate, framerate, resolution string) {
	lines := strings.Split(outputStr, "\n")
	var inVideoStream bool

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, `"codec_type": "video"`) {
			inVideoStream = true
			continue
		}
		if strings.Contains(line, `"codec_type": "audio"`) {
			inVideoStream = false
		}

		if inVideoStream {
			if strings.Contains(line, `"bit_rate"`) && bitrate == "" {
				parts := strings.Split(line, `"`)
				for i, part := range parts {
					if part == "bit_rate" && i+2 < len(parts) {
						if br, err := strconv.Atoi(parts[i+2]); err == nil {
							bitrate = fmt.Sprintf("%.1f Mbps", float64(br)/1000000)
						}
						break
					}
				}
			}
			if strings.Contains(line, `"r_frame_rate"`) && framerate == "" {
				parts := strings.Split(line, `"`)
				for i, part := range parts {
					if part == "r_frame_rate" && i+2 < len(parts) {
						frameRate := parts[i+2]
						if strings.Contains(frameRate, "/") {
							rateParts := strings.Split(frameRate, "/")
							if len(rateParts) == 2 {
								if num, err1 := strconv.ParseFloat(rateParts[0], 64); err1 == nil {
									if den, err2 := strconv.ParseFloat(rateParts[1], 64); err2 == nil && den != 0 {
										framerate = fmt.Sprintf("%.0f fps", num/den)
									}
								}
							}
						}
						break
					}
				}
			}
			if strings.Contains(line, `"width"`) && strings.Contains(line, `"height"`) && resolution == "" {
				resolution = activeResolution
			}
		}
	}
	return bitrate, framerate, resolution
}

func extractAudioInfo(outputStr string) string {
	if !strings.Contains(outputStr, `"codec_type": "audio"`) {
		return ""
	}

	lines := strings.Split(outputStr, "\n")
	var inAudioStream bool

	for _, line := range lines {
		if strings.Contains(line, `"codec_type": "audio"`) {
			inAudioStream = true
			continue
		}
		if strings.Contains(line, `"codec_type": "video"`) {
			inAudioStream = false
		}

		if inAudioStream && strings.Contains(line, `"bit_rate"`) {
			parts := strings.Split(line, `"`)
			for i, part := range parts {
				if part == "bit_rate" && i+2 < len(parts) {
					if bitrate, err := strconv.Atoi(parts[i+2]); err == nil {
						return fmt.Sprintf("%d kbps", bitrate/1000)
					}
				}
			}
			break
		}
	}
	return ""
}

func runFFmpegCommand(args []string, hasAudio bool) error {
	cmd := newExecCommand("ffmpeg", args...)
	var stderr bytes.Buffer

	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open /dev/null: %v", err)
	}
	defer devNull.Close()
	cmd.Stdout = devNull
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ffmpeg start failed: %v", err)
	}

	showSpinner := os.Getenv("GO24K_INTERNAL_CLI") != "1"
	var done chan struct{}
	if showSpinner {
		done = make(chan struct{})
		go func() {
			spinnerChars := []string{"|", "/", "-", "\\"}
			i := 0
			message := "Generating video (no audio)"
			if hasAudio {
				message = "Generating video with audio"
			}

			for {
				select {
				case <-done:
					fmt.Print("\r")
					return
				default:
					fmt.Printf("\r%s %s...", spinnerChars[i%len(spinnerChars)], message)
					i++
					time.Sleep(200 * time.Millisecond)
				}
			}
		}()
	}

	if err := cmd.Wait(); err != nil {
		if showSpinner {
			close(done)
		}
		stderrOutput := strings.TrimSpace(stderr.String())
		if stderrOutput != "" {
			return fmt.Errorf("ffmpeg command failed: %v\n%s", err, stderrOutput)
		}
		return fmt.Errorf("ffmpeg command failed: %v", err)
	}
	if showSpinner {
		close(done)
	}
	return nil
}

func displayVideoInfo(outputFilename string, finalLength float64) {
	resLabel := "4K UHD"
	if activeResolution == resolutionFullHD {
		resLabel = "Full HD"
	}
	fmt.Printf("\n=== Video generated successfully! ===\n")
	fmt.Printf("File: %s\n", outputFilename)

	if videoInfo, err := getVideoDetails(outputFilename); err == nil {
		fmt.Printf("Resolution: %s (%s)\n", videoInfo.Resolution, resLabel)
		fmt.Printf("Duration: %.2f sec. (%.1fs actual)\n", finalLength, videoInfo.DurationSec)
		fmt.Printf("File Size: %.1f MB\n", videoInfo.FileSizeMB)
		fmt.Printf("Video Bitrate: %s\n", videoInfo.VideoBitrate)
		fmt.Printf("Audio Bitrate: %s\n", videoInfo.AudioBitrate)
		fmt.Printf("Framerate: %s\n", videoInfo.Framerate)
	} else {
		fmt.Printf("Resolution: %s (%s)\n", activeResolution, resLabel)
		fmt.Printf("Duration: %.2f sec.\n", finalLength)
		if fileInfo, err := os.Stat(outputFilename); err == nil {
			sizeMB := float64(fileInfo.Size()) / (1024 * 1024)
			fmt.Printf("File Size: %.1f MB\n", sizeMB)
		}
	}
}
