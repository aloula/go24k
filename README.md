# Go24K

Go24K is a Go program that processes JPEG images in the current directory, resizes them to a height of 2160 pixels, and composites them onto a black background image with dimensions 3840x2160 pixels (4K - UHD format). The processed images are saved with a new name that includes the timestamp from the image's EXIF data. The UHD resized images are used to create a MP4 Video (H.264 codec) with music and crossfade between each image.

## Features

- **4K Video Creation**: Generates high-quality 4K UHD (3840x2160) videos from JPEG images
- **Ken Burns Effect**: Applies dynamic zoom and pan effects to images for cinematic feel
- **Flexible Audio**: Works with or without MP3 audio - automatically detects and adapts
- **Image Processing**: Automatically resizes images to 2160p height while maintaining aspect ratio
- **Smart Compositing**: Centers images on black backgrounds for consistent 4K output
- **EXIF Timestamp Support**: Uses image metadata for chronological ordering and filename generation
- **Audio Synchronization**: Seamlessly integrates background music with fade-in/fade-out effects
- **Crossfade Transitions**: Smooth fade effects between images with customizable duration
- **Enhanced Progress Tracking**: Real-time progress with resolution info and processing statistics
- **Flexible Timing**: Configurable image duration and transition timing
- **Cross-Platform Compatibility**: Works on Windows CMD, Linux, and macOS with appropriate UI
- **Multi-Platform**: Pre-built binaries for Linux, macOS (Intel/ARM), and Windows

## Requirements

- Go 1.16 or later
- The following Go packages:
  - `github.com/disintegration/imaging`
  - `github.com/rwcarlsen/goexif/exif`
  - `github.com/schollz/progressbar/v3`
- FFMpeg

## Installation

### Quick Start

1. **Install FFmpeg**: https://ffmpeg.org/download.html
2. **Download the executable**: Get the pre-built binary from the [releases](../../releases) or [builds](./builds/) directory for your platform:
   - **Linux (x64)**: `builds/linux/amd64/go24k`
   - **Linux (ARM64)**: `builds/linux/arm64/go24k`
   - **macOS (Apple Silicon)**: `builds/macos/arm/go24k`
   - **macOS (Intel)**: `builds/macos/intel/go24k`
   - **Windows (x64)**: `builds/windows/amd64/go24k.exe`
   - **Windows (ARM64)**: `builds/windows/arm64/go24k.exe`

3. **Make executable** (Linux/macOS):
    ```sh
    chmod +x go24k
    ```

### Development Setup

For developers who want to build from source:

1. **Clone the repository**:
    ```sh
    git clone https://github.com/aloula/go24k.git
    cd go24k
    ```

2. **Install Go dependencies**:
    ```sh
    go mod download
    ```

3. **Build the executable**:
    ```sh
    go build -o go24k
    ```

## Usage

### Basic Usage

1. **Place your JPEG images** in the current directory (required)
2. **Optionally add an MP3 file** for background music (not required)
3. **Run the program**:

    ```sh
    ./go24k
    ```

4. **Watch the enhanced progress**:
   - Real-time resolution conversion display
   - Processing statistics and speed metrics
   - Audio detection and mode confirmation

5. **Output**: The program creates:
   - `converted/` directory with processed 4K images  
   - `video.mp4` with or without audio as detected

### Command Line Options

Go24K supports several command-line options to customize the video generation:

```sh
./go24k [OPTIONS]
```

**Available Options:**

- `-convert-only`  
  Convert images only, without generating the video. Useful for preprocessing images or testing conversion settings.
  
- `-d <seconds>`  
  Duration per image in seconds (default: 5). Controls how long each image is displayed in the final video.
  
- `-static`  
  Disable Ken Burns effect; use static images with transitions only. Creates a simpler slideshow without zoom/pan effects.
  
- `-t <seconds>`  
  Transition (fade) duration in seconds (default: 1). Controls the crossfade time between images.

**Examples:**

```sh
# Basic video with default settings (5s per image, 1s transitions, Ken Burns effect)
./go24k

# Quick slideshow (2s per image, 0.5s transitions)
./go24k -d 2 -t 0.5

# Static slideshow without Ken Burns effect
./go24k -static -d 3 -t 1

# Long-form video (10s per image, 2s transitions)
./go24k -d 10 -t 2

# Only convert images, don't generate video
./go24k -convert-only

# Works with or without audio automatically:
# • With music.mp3 present: Creates video with synchronized audio
# • No MP3 files: Creates silent video (no errors)
```

## Building

### Cross-Platform Build

To build the program for all supported platforms, use the provided build script:

1. **Make the script executable**:
    ```sh
    chmod +x build.sh
    ```

2. **Generate builds for all platforms**:
    ```sh
    ./build.sh
    ```

    This creates binaries for:
    - Linux (x64)
    - Linux (ARM64)
    - macOS (Apple Silicon ARM64)
    - macOS (Intel x64)
    - Windows (x64)
    - Windows (ARM64)

3. **Find your builds** in the `builds/` directory:
    ```
    builds/
    ├── linux/
    │   ├── amd64/go24k
    │   └── arm64/go24k
    ├── macos/
    │   ├── arm/go24k
    │   └── intel/go24k
    └── windows/
        ├── amd64/go24k.exe
        └── arm64/go24k.exe
    ```

### Manual Build

For a single platform build:

```sh
# Build for current platform
go build -o go24k

# Build for specific platform (examples)
GOOS=linux GOARCH=amd64 go build -o go24k-linux
GOOS=darwin GOARCH=arm64 go build -o go24k-macos-arm
GOOS=windows GOARCH=amd64 go build -o go24k-windows.exe
```

## Output

Go24K generates the following files and provides detailed feedback:

- **`converted/`** - Directory containing processed 4K images with EXIF timestamps
- **`video.mp4`** - Final 4K video with or without audio (H.264 codec)

### Processing Feedback

During execution, you'll see:

```
=== Starting Image Conversion ===
Images to process: 12
Target: 4K UHD (3840x2160) with 2160p height scaling
Output: converted/ directory

Converting IMG_001.jpg (4032x3024->3840x2160) [████████   ] 8/12

=== Image conversion completed! ===

=== Conversion Statistics ===
   Processing time: 15.3 seconds
   Average speed: 0.8 images/sec
   Original total size: 45.2 MB
   Converted total size: 67.8 MB
   Size ratio: 1.5x

Audio file found: background.mp3
Generating video with audio...: |

=== Video generated successfully! ===
File: video.mp4
Resolution: 4K UHD (3840x2160)
Duration: 45 seconds
Images: 12
Audio: background.mp3
Size: 156.3 MB
```

### Technical Specifications

The output video uses:
- **Resolution**: 3840x2160 (4K UHD)
- **Codec**: H.264 (libx264) with CRF 23 quality
- **Audio**: AAC 192kbps with fade-in/fade-out (when MP3 present)
- **Frame Rate**: 30fps for smooth playback
- **Compatibility**: Optimized for universal playback

## Troubleshooting

### Common Issues

**FFmpeg not found**
```
Error: ffmpeg command not found
```
- Install FFmpeg from https://ffmpeg.org/download.html
- Ensure FFmpeg is in your system PATH

**No images found**
```
No .jpg files found in current directory
```
- Place JPEG images (*.jpg) in the same directory as go24k
- Check file extensions (must be .jpg, not .jpeg)

**Permission denied (Linux/macOS)**
```
zsh: permission denied: ./go24k
```
- Make the binary executable: `chmod +x go24k`

**Audio handling (improved)**
- ✅ **No MP3 required**: Program automatically creates silent video if no MP3 found
- ✅ **Auto-detection**: Shows "Audio file found" or "No MP3 file found" messages  
- ✅ **Flexible**: Works equally well with or without background music
- If you want audio: Place any `.mp3` file in the directory
- If audio issues persist: Try a different MP3 file or run without audio

### Recent Improvements (v2.0)

Go24K has been significantly enhanced with:

- **🔧 Robust Audio Handling**: No longer requires MP3 files - works with or without audio
- **📊 Enhanced Progress Display**: Real-time resolution conversion info and statistics
- **🖥️ Cross-Platform UI**: Optimized display for Windows CMD, Linux, and macOS terminals
- **⚡ Better Performance**: Software-based H.264 encoding for universal compatibility
- **📈 Detailed Feedback**: Processing statistics, speeds, file sizes, and completion summaries
- **🎯 Improved Reliability**: Better error handling and user guidance

### Performance Tips

- **Large images**: The program automatically resizes to 2160p, but starting with smaller images (2K-4K) improves processing speed
- **Many images**: Processing time scales with image count; expect 1-2 seconds per image on modern systems
- **Long videos**: For videos longer than 10 minutes, ensure sufficient disk space for the 4K output
- **Audio optional**: Skip MP3 files for faster processing or when audio isn't needed

## License

This project is licensed under the MIT License. See the LICENSE file for details.