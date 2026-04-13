//go:build fyne

package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go24k/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var guiIconCandidates = []string{
	"app-icon.png",
	"assets/app-icon.png",
}

const (
	guiMotionLabelLow    = "Pan + zoom (low)"
	guiMotionLabelMedium = "Pan + zoom (medium)"
	guiMotionLabelHigh   = "Pan + zoom (high)"
)

type guiOptions struct {
	inputFolder     string
	convertOnly     bool
	staticImages    bool
	duration        int
	transition      int
	fpsMode         string
	fitAudio        bool
	includeVideos   bool
	includeMOV      bool
	keepVideoAudio  bool
	orderByFilename bool
	fullHD          bool
	kenBurnsMode    string
	exifOverlay     bool
	overlayFontSize int
}

type guiLogBuffer struct {
	stable           strings.Builder
	currentLine      strings.Builder
	currentTransient bool
}

type disableable interface {
	Disable()
	Enable()
}

var errGUIProcessStopped = errors.New("generation stopped by user")

func (b *guiLogBuffer) Append(chunk string) string {
	for i := 0; i < len(chunk); i++ {
		switch chunk[i] {
		case '\r':
			b.currentLine.Reset()
			b.currentTransient = true
		case '\n':
			if !b.currentTransient && b.currentLine.Len() > 0 {
				b.stable.WriteString(b.currentLine.String())
			}
			b.stable.WriteByte('\n')
			b.currentLine.Reset()
			b.currentTransient = false
		default:
			b.currentLine.WriteByte(chunk[i])
		}
	}

	return b.stable.String()
}

func (b *guiLogBuffer) Flush() string {
	if !b.currentTransient && b.currentLine.Len() > 0 {
		b.stable.WriteString(b.currentLine.String())
		b.currentLine.Reset()
	}

	b.currentTransient = false
	return b.stable.String()
}

func motionStyleOptions() []string {
	return []string{guiMotionLabelLow, guiMotionLabelMedium, guiMotionLabelHigh}
}

func motionStyleToKenBurnsMode(label string) string {
	switch label {
	case guiMotionLabelLow:
		return "low"
	case guiMotionLabelMedium:
		return "medium"
	case guiMotionLabelHigh:
		return "high"
	default:
		return "high"
	}
}

func formatElapsedDuration(elapsed time.Duration) string {
	totalSeconds := int(elapsed.Seconds())
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("Elapsed: %02d:%02d", minutes, seconds)
}

func launchGUI() {
	a := app.NewWithID("com.aloula.go24k")
	w := a.NewWindow(fmt.Sprintf("%s - Video Generator", utils.GetVersionInfo()))
	if icon := loadGUIIcon(); icon != nil {
		a.SetIcon(icon)
		w.SetIcon(icon)
	}
	w.Resize(fyne.NewSize(820, 620))

	folderEntry := widget.NewEntry()
	folderEntry.SetPlaceHolder("Select the folder that contains your images/music/videos")
	folderEntry.Disable()

	browseButton := widget.NewButton("Browse", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if uri == nil {
				return
			}
			folderPath := uri.Path()
			folderEntry.SetText(folderPath)
		}, w)
	})

	convertOnlyCheck := widget.NewCheck("Convert images only", nil)
	staticCheck := widget.NewCheck("Disable Ken Burns (static images)", nil)
	fitAudioCheck := widget.NewCheck("Fit timeline to audio length", nil)
	includeVideosCheck := widget.NewCheck("Include video files", nil)
	includeMOVCheck := widget.NewCheck("Include MOV files", nil)
	keepVideoAudioCheck := widget.NewCheck("Keep source video audio", nil)
	orderByFilenameCheck := widget.NewCheck("Order by filename", nil)
	fullHDCheck := widget.NewCheck("Output Full HD (1920x1080)", nil)
	exifOverlayCheck := widget.NewCheck("Enable EXIF overlay", nil)

	updateVideoAudioControl := func() {
		if includeVideosCheck.Checked || includeMOVCheck.Checked {
			keepVideoAudioCheck.Enable()
			return
		}
		keepVideoAudioCheck.SetChecked(false)
		keepVideoAudioCheck.Disable()
	}

	includeVideosCheck.OnChanged = func(bool) {
		updateVideoAudioControl()
	}
	includeMOVCheck.OnChanged = func(bool) {
		updateVideoAudioControl()
	}
	keepVideoAudioCheck.Disable()

	durationEntry := widget.NewEntry()
	durationEntry.SetText("5")
	transitionEntry := widget.NewEntry()
	transitionEntry.SetText("1")
	fontSizeEntry := widget.NewEntry()
	fontSizeEntry.SetText("36")

	fpsSelect := widget.NewSelect([]string{"auto", "30", "60"}, nil)
	fpsSelect.SetSelected("auto")

	kenBurnsSelect := widget.NewSelect(motionStyleOptions(), nil)
	kenBurnsSelect.SetSelected(guiMotionLabelHigh)
	staticCheck.OnChanged = func(checked bool) {
		if checked {
			kenBurnsSelect.Disable()
			return
		}
		kenBurnsSelect.Enable()
	}

	logOutput := widget.NewMultiLineEntry()
	logOutput.Wrapping = fyne.TextWrapWord
	logOutput.SetMinRowsVisible(14)
	logOutput.Disable()
	logBinding := binding.NewString()
	_ = logBinding.Set("")
	logOutput.Bind(logBinding)
	elapsedBinding := binding.NewString()
	_ = elapsedBinding.Set("Elapsed: 00:00")
	elapsedLabel := widget.NewLabelWithData(elapsedBinding)
	elapsedLabel.TextStyle = fyne.TextStyle{Monospace: true}

	runButton := widget.NewButton("Generate Video", nil)
	runButton.Importance = widget.HighImportance
	stopButton := widget.NewButton("Stop", nil)
	stopButton.Importance = widget.DangerImportance
	stopButton.Disable()
	stopButton.Hide()

	runButton.OnTapped = func() {
		durationValue, err := strconv.Atoi(strings.TrimSpace(durationEntry.Text))
		if err != nil || durationValue <= 0 {
			dialog.ShowError(fmt.Errorf("invalid image duration: use a value greater than 0"), w)
			return
		}

		transitionValue, err := strconv.Atoi(strings.TrimSpace(transitionEntry.Text))
		if err != nil || transitionValue <= 0 {
			dialog.ShowError(fmt.Errorf("invalid transition duration: use a value greater than 0"), w)
			return
		}

		fontSizeValue, err := strconv.Atoi(strings.TrimSpace(fontSizeEntry.Text))
		if err != nil || fontSizeValue <= 0 {
			dialog.ShowError(fmt.Errorf("invalid overlay font size: use a value greater than 0"), w)
			return
		}

		inputFolder := strings.TrimSpace(folderEntry.Text)
		if inputFolder == "" {
			dialog.ShowInformation("Input folder required", "Select the folder containing your media files.", w)
			return
		}

		folderInfo, err := os.Stat(inputFolder)
		if err != nil || !folderInfo.IsDir() {
			dialog.ShowError(fmt.Errorf("invalid input folder: %s", inputFolder), w)
			return
		}

		opts := guiOptions{
			inputFolder:     inputFolder,
			convertOnly:     convertOnlyCheck.Checked,
			staticImages:    staticCheck.Checked,
			duration:        durationValue,
			transition:      transitionValue,
			fpsMode:         fpsSelect.Selected,
			fitAudio:        fitAudioCheck.Checked,
			includeVideos:   includeVideosCheck.Checked,
			includeMOV:      includeMOVCheck.Checked,
			keepVideoAudio:  (includeVideosCheck.Checked || includeMOVCheck.Checked) && keepVideoAudioCheck.Checked,
			orderByFilename: orderByFilenameCheck.Checked,
			fullHD:          fullHDCheck.Checked,
			kenBurnsMode:    motionStyleToKenBurnsMode(kenBurnsSelect.Selected),
			exifOverlay:     exifOverlayCheck.Checked,
			overlayFontSize: fontSizeValue,
		}

		runButton.Disable()
		setControlsEnabled(false,
			browseButton,
			convertOnlyCheck,
			staticCheck,
			fitAudioCheck,
			includeVideosCheck,
			includeMOVCheck,
			keepVideoAudioCheck,
			orderByFilenameCheck,
			fullHDCheck,
			exifOverlayCheck,
			durationEntry,
			transitionEntry,
			fontSizeEntry,
			fpsSelect,
			kenBurnsSelect,
			runButton,
		)
		stopRequested := make(chan struct{})
		stopRequestIssued := false
		stopButton.OnTapped = func() {
			if stopRequestIssued {
				return
			}
			stopRequestIssued = true
			stopButton.Disable()
			close(stopRequested)
		}
		stopButton.Enable()
		stopButton.Show()
		_ = elapsedBinding.Set("Elapsed: 00:00")
		_ = logBinding.Set("")
		scrollEntryToEnd(logOutput, "")

		go func() {
			startedAt := time.Now()
			stopElapsedUpdates := make(chan struct{})
			go func() {
				ticker := time.NewTicker(time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-stopElapsedUpdates:
						return
					case <-ticker.C:
						_ = elapsedBinding.Set(formatElapsedDuration(time.Since(startedAt)))
					}
				}
			}()

			logChunks := make(chan string, 1024)
			logDone := make(chan string, 1)
			go func() {
				var fullLog guiLogBuffer
				var pending strings.Builder
				ticker := time.NewTicker(120 * time.Millisecond)
				defer ticker.Stop()

				flushPending := func() {
					if pending.Len() == 0 {
						return
					}
					updatedLog := fullLog.Append(pending.String())
					pending.Reset()
					_ = logBinding.Set(updatedLog)
					scrollEntryToEnd(logOutput, updatedLog)
				}

				for {
					select {
					case chunk, ok := <-logChunks:
						if !ok {
							flushPending()
							finalLog := fullLog.Flush()
							_ = logBinding.Set(finalLog)
							scrollEntryToEnd(logOutput, finalLog)
							logDone <- finalLog
							return
						}
						pending.WriteString(chunk)
						if pending.Len() >= 8192 {
							flushPending()
						}
					case <-ticker.C:
						flushPending()
					}
				}
			}()

			appendLog := func(chunk string) {
				if chunk == "" {
					return
				}
				logChunks <- chunk
			}

			runErr := runGeneratorFromGUIStreaming(opts, appendLog, stopRequested)
			close(stopElapsedUpdates)
			_ = elapsedBinding.Set(formatElapsedDuration(time.Since(startedAt)))
			close(logChunks)
			finalLog := <-logDone
			_ = logBinding.Set(finalLog)
			scrollEntryToEnd(logOutput, finalLog)

			stopButton.Disable()
			stopButton.Hide()
			setControlsEnabled(true,
				browseButton,
				convertOnlyCheck,
				staticCheck,
				fitAudioCheck,
				includeVideosCheck,
				includeMOVCheck,
				orderByFilenameCheck,
				fullHDCheck,
				exifOverlayCheck,
				durationEntry,
				transitionEntry,
				fontSizeEntry,
				fpsSelect,
				kenBurnsSelect,
				runButton,
			)
			updateVideoAudioControl()
			if staticCheck.Checked {
				kenBurnsSelect.Disable()
			}

			if runErr != nil {
				if errors.Is(runErr, errGUIProcessStopped) {
					dialog.ShowInformation("Stopped", "Generation stopped.", w)
					return
				}
				dialog.ShowError(runErr, w)
				return
			}

			dialog.ShowInformation("Done", "Video generated successfully.", w)
		}()
	}

	folderRow := container.NewBorder(nil, nil, nil, browseButton, folderEntry)

	timingGrid := container.NewGridWithColumns(3,
		labeledField("Image duration (sec)", durationEntry),
		labeledField("Transition duration (sec)", transitionEntry),
		labeledField("FPS", fpsSelect),
	)

	overlayGrid := container.NewGridWithColumns(1,
		labeledField("Overlay font size", fontSizeEntry),
	)

	optionsGrid := container.NewGridWithColumns(2,
		container.NewVBox(
			convertOnlyCheck,
			staticCheck,
			fitAudioCheck,
			labeledField("Motion style", kenBurnsSelect),
		),
		container.NewVBox(
			includeVideosCheck,
			includeMOVCheck,
			keepVideoAudioCheck,
			orderByFilenameCheck,
			fullHDCheck,
			exifOverlayCheck,
		),
	)

	optionsCol := container.NewVBox(
		widget.NewLabel("Input folder"),
		folderRow,
		widget.NewSeparator(),
		widget.NewLabel("Video options"),
		timingGrid,
		overlayGrid,
		optionsGrid,
		container.NewHBox(elapsedLabel, layout.NewSpacer(), stopButton, runButton),
	)

	logPanel := widget.NewCard("", "", logOutput)

	content := container.NewBorder(
		nil,
		logPanel,
		nil,
		nil,
		optionsCol,
	)

	w.SetContent(content)
	w.ShowAndRun()
}

func loadGUIIcon() fyne.Resource {
	for _, candidate := range guiIconCandidates {
		if icon := loadIconFromPath(candidate); icon != nil {
			return icon
		}
	}

	execPath, err := os.Executable()
	if err != nil {
		return nil
	}

	execDir := filepath.Dir(execPath)
	for _, candidate := range guiIconCandidates {
		if icon := loadIconFromPath(filepath.Join(execDir, candidate)); icon != nil {
			return icon
		}
	}

	return nil
}

func loadIconFromPath(iconPath string) fyne.Resource {
	data, err := os.ReadFile(iconPath)
	if err != nil {
		return nil
	}

	return fyne.NewStaticResource(filepath.Base(iconPath), data)
}

func labeledField(label string, field fyne.CanvasObject) fyne.CanvasObject {
	return container.NewVBox(widget.NewLabel(label), field)
}

func setControlsEnabled(enabled bool, controls ...disableable) {
	for _, control := range controls {
		if enabled {
			control.Enable()
			continue
		}
		control.Disable()
	}
}

func scrollEntryToEnd(entry *widget.Entry, text string) {
	entry.CursorRow = strings.Count(text, "\n")
	lastLineBreak := strings.LastIndex(text, "\n")
	if lastLineBreak == -1 {
		entry.CursorColumn = len(text)
	} else {
		entry.CursorColumn = len(text) - lastLineBreak - 1
	}
	entry.Refresh()
}

func runGeneratorFromGUIStreaming(opts guiOptions, onOutput func(string), stopRequested <-chan struct{}) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	exePath, err = filepath.Abs(exePath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable absolute path: %w", err)
	}

	args := []string{}
	args = append(args, "-d", strconv.Itoa(opts.duration))
	args = append(args, "-t", strconv.Itoa(opts.transition))
	args = append(args, "-kenburns-mode", opts.kenBurnsMode)
	args = append(args, "-overlay-font-size", strconv.Itoa(opts.overlayFontSize))

	if opts.convertOnly {
		args = append(args, "-convert-only")
	}
	if opts.staticImages {
		args = append(args, "-static")
	}
	if opts.fpsMode == "30" || opts.fpsMode == "60" {
		args = append(args, "-fps", opts.fpsMode)
	}
	if opts.fitAudio {
		args = append(args, "-fit-audio")
	}
	if opts.includeVideos {
		args = append(args, "-include-videos")
	}
	if opts.includeMOV {
		args = append(args, "-include-mov")
	}
	if opts.keepVideoAudio {
		args = append(args, "-keep-video-audio")
	}
	if opts.orderByFilename {
		args = append(args, "-order-by-filename")
	}
	if opts.fullHD {
		args = append(args, "-fullhd")
	}
	if opts.exifOverlay {
		args = append(args, "-exif-overlay")
	}

	cmd := exec.Command(exePath, args...)
	cmd.Dir = opts.inputFolder
	cmd.Env = append(os.Environ(), "GO24K_INTERNAL_CLI=1")
	prepareGUICommand(cmd)

	reader, writer, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("failed to create output pipe: %w", err)
	}
	defer func() {
		_ = reader.Close()
	}()

	cmd.Stdout = writer
	cmd.Stderr = writer

	if err := cmd.Start(); err != nil {
		_ = writer.Close()
		return fmt.Errorf("failed to start go24k: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close parent output pipe writer: %w", err)
	}

	readDone := make(chan error, 1)
	go func() {
		buf := make([]byte, 4096)
		for {
			n, readErr := reader.Read(buf)
			if n > 0 {
				onOutput(string(buf[:n]))
			}

			if readErr != nil {
				if readErr != io.EOF {
					readDone <- readErr
					return
				}
				readDone <- nil
				return
			}
		}
	}()

	waitDone := make(chan error, 1)
	go func() {
		waitDone <- cmd.Wait()
	}()

	stopped := false
	var waitErr error
	select {
	case waitErr = <-waitDone:
	case <-stopRequested:
		stopped = true
		onOutput("\nStopping generation...\n")
		if err := terminateGUICommand(cmd); err != nil {
			onOutput(fmt.Sprintf("Failed to stop process tree cleanly: %v\n", err))
		}
		waitErr = <-waitDone
	}

	readErr := <-readDone
	if readErr != nil {
		return fmt.Errorf("failed reading process output: %w", readErr)
	}

	if stopped {
		return errGUIProcessStopped
	}

	if waitErr != nil {
		return fmt.Errorf("go24k failed: %w", waitErr)
	}

	return nil
}
