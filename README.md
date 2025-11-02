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
- **Hardware Acceleration**: Automatic NVIDIA NVENC detection for GPU-accelerated encoding

## Hardware Acceleration

Go24K automatically detects and uses hardware acceleration across multiple platforms:

### **Snapdragon X Plus Support** 🧠
- **Encoder**: Windows Media Foundation (`h264_mf`)
- **Tested Performance**: ~20% faster encoding (25.7s vs 30s+ CPU)
- **Power Efficiency**: Significantly lower power consumption
- **Quality**: Optimized for ARM SoC hardware encoding
- **Automatic Detection**: Works seamlessly on Windows ARM devices

### **NVIDIA NVENC Support** 🚀
- **Encoder**: NVIDIA hardware encoder (`h264_nvenc`)
- **Performance Boost**: 5-10x faster encoding compared to CPU-only
- **Reduced CPU Usage**: GPU handles video encoding workload
- **Quality Optimized**: Uses Constant Quality (CQ) mode for consistent results

### **Intel Quick Sync Video (QSV)** ⚡
- **Encoder**: Intel hardware encoder (`h264_qsv`)
- **Performance**: 2-4x faster than CPU encoding
- **Supported**: Intel processors with integrated graphics (HD 2000+)

### **AMD AMF Support** 🔥
- **Encoder**: AMD Advanced Media Framework (`h264_amf`)
- **Performance**: 2-4x faster than CPU encoding
- **Supported**: AMD GPUs and APUs with AMF support

### **Linux VAAPI** 🐧
- **Encoder**: Video Acceleration API (`h264_vaapi`)
- **Performance**: 2-3x faster than CPU encoding
- **Supported**: Intel and AMD integrated graphics on Linux

### **Automatic Selection Priority**
1. **NVIDIA NVENC** (highest performance)
2. **Windows Media Foundation** (Snapdragon X, Intel, AMD on Windows)
3. **Intel Quick Sync Video**
4. **AMD AMF**
5. **Linux VAAPI**
6. **CPU libx264** (universal fallback)

### **Real-World Performance** 📊
**Test Case**: 5 images → 21-second 4K video with audio
- **Snapdragon X Plus**: 25.7 seconds (Media Foundation)
- **Estimated CPU**: 30+ seconds (libx264 software)
- **NVIDIA RTX**: ~15-20 seconds (NVENC)

## Quality Optimization

Go24K automatically detects your hardware and environment to apply optimal encoding settings:

### **Encoding Method Selection**
- **NVIDIA NVENC** (if available): GPU-accelerated encoding with CQ 21
- **CPU libx264** (fallback): Software encoding with CRF 21

### **Performance Comparison**
| Encoder | Speed vs CPU | Power Usage | Quality | Best For |
|---------|--------------|-------------|---------|----------|
| NVENC | 5-10x faster | Very Low | High | Gaming PCs, Workstations |
| Media Foundation | 1.2x faster* | Very Low | High | Snapdragon X, Windows ARM |
| Intel QSV | 2-4x faster | Low | High | Intel CPUs with iGPU |
| AMD AMF | 2-4x faster | Low | High | AMD GPUs/APUs |
| VAAPI | 2-3x faster | Low | High | Linux with Intel/AMD iGPU |
| libx264 (CPU) | Baseline | High | Excellent | All systems (fallback) |

*Tested: 25.7s vs estimated 30s+ on Snapdragon X Plus

### Environment Detection
Use the `--debug` flag to see detected hardware and encoding settings:
```bash
./go24k --debug
```

**Example Output (NVIDIA GPU):**
```
🚀 NVIDIA NVENC: Available (GPU acceleration enabled)
⚡ Performance: ~5-10x faster than CPU encoding
💾 CPU Usage: Significantly reduced
```

**Example Output (Snapdragon X Plus):**
```
🧠 Windows Media Foundation: Available (Snapdragon X, Intel, AMD)
🎯 Using: Windows Media Foundation
🧠 Optimized for: Snapdragon X Plus hardware encoding
⚡ Performance: ~20% faster than CPU + power efficient
```

### Quality Settings Explained
- **NVENC CQ 18-20**: Visually lossless (GPU encoding)
- **NVENC CQ 21-23**: High quality (recommended for GPU)
- **libx264 CRF 18-20**: Visually lossless (CPU encoding)  
- **libx264 CRF 21-23**: High quality (recommended for CPU)

## Requirements

### **Essential**
- Go 1.16 or later (for building from source)
- FFmpeg (with H.264 support)
- The following Go packages (for development):
  - `github.com/disintegration/imaging`
  - `github.com/rwcarlsen/goexif/exif`
  - `github.com/schollz/progressbar/v3`

### **Optional (for Hardware Acceleration)**
- **NVIDIA GPU**: GeForce GTX 10-series or newer, RTX series, Quadro with NVENC
- **FFmpeg with NVENC**: Build or binary that includes `h264_nvenc` encoder
- **NVIDIA Drivers**: Recent version with NVENC support

### **FFmpeg Installation Notes**
- **Standard FFmpeg**: CPU encoding with libx264 (works everywhere)
- **FFmpeg with NVENC**: GPU acceleration (Windows/Linux with NVIDIA drivers)
- Use `ffmpeg -encoders | grep nvenc` to check NVENC availability

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
  
- `--debug`  
  Show environment detection and video encoding optimization info. Displays detected OS, architecture, and quality settings that will be used.
  
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

# Check hardware acceleration and environment settings
./go24k --debug

# Example output with NVIDIA GPU:
# 🚀 Hardware: NVIDIA NVENC detected - using GPU acceleration
# ⚡ Performance: ~5-10x faster than CPU encoding
# 💾 CPU Usage: Significantly reduced

# Works with or without audio automatically:
# • With music.mp3 present: Creates video with synchronized audio
# • No MP3 files: Creates silent video (no errors)
```

## Performance Benchmarks

### **Real-World Test Results** 📊

**Test Configuration:**
- 5 JPEG images (25.2 MB total)
- Output: 21-second 4K video with audio
- Command: `go24k -d 5 -t 1`

**Hardware Performance:**

| Platform | Hardware | Encoding Time | Improvement | Power Usage |
|----------|----------|---------------|-------------|-------------|
| **Snapdragon X Plus** | Media Foundation | **25.7 seconds** | ~20% faster | Very Low |
| Intel/AMD Desktop | CPU libx264 | ~30+ seconds | Baseline | High |
| NVIDIA RTX | NVENC | ~15-20 seconds* | 5-10x faster | Low |

*Estimated based on typical NVENC performance

**Key Insights:**
- ✅ **Snapdragon X Plus**: Excellent power efficiency with solid performance gains
- ✅ **NVIDIA GPUs**: Highest performance for video-intensive workflows
- ✅ **Automatic Selection**: Always chooses the best available encoder
- ✅ **Quality Maintained**: Hardware acceleration preserves video quality

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

### Hardware Acceleration Issues

**NVENC not detected despite having NVIDIA GPU**
```
💻 CPU: Using libx264 software encoding
```
**Solutions:**
- Update NVIDIA drivers to latest version
- Verify FFmpeg includes NVENC: `ffmpeg -encoders | grep nvenc`
- Install FFmpeg build with NVENC support
- Check GPU supports NVENC (GTX 10-series or newer)

**NVENC encoding fails**
```
ffmpeg command failed: exit status 1
```
**Solutions:**
- Check available VRAM (4K encoding needs ~2GB+)
- Close other GPU-intensive applications
- Fallback will use CPU encoding automatically
- Use `./go24k --debug` to verify settings

**Performance not improved with NVENC**
- NVENC provides encoding speed boost, not necessarily file generation speed
- Biggest improvement seen with longer videos (>30 seconds)
- I/O operations (reading images) still use CPU/storage

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