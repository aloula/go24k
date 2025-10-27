# Go24K

Go24K is a versatile Go program that transforms JPEG images into stunning **4K UHD videos** and **animated GIFs**. It intelligently processes images with automatic resizing, timestamp-based naming, and creates professional-quality output with smooth transitions and effects.

**Key Capabilities:**
- 🎬 **UHD Video Generation**: Creates 4K videos (3840x2160) with Ken Burns effects, crossfade transitions, and synchronized audio
- 🎞️ **Animated GIF Creation**: Generates optimized GIFs with customizable timing, frame rates, and palette optimization
- 🖼️ **Smart Image Processing**: Automatically resizes, centers, and timestamps images from EXIF data
- ⚡ **Dual Processing Modes**: Separate optimization pipelines for video (4K) and GIF (1080p) output

## Features

### 🎬 Video Generation
- **4K UHD Output:** Creates stunning 3840x2160 videos with H.264 codec
- **Ken Burns Effect:** Optional zoom/pan effects for dynamic presentation
- **Audio Integration:** Synchronized music with fade-in/fade-out effects
- **Crossfade Transitions:** Smooth transitions between images
- **Professional Quality:** Optimized encoding with CUDA acceleration support

### 🎞️ Animated GIF Creation
- **Smart Time Control:** Set total GIF duration or per-image timing
- **Palette Optimization:** Advanced color reduction for smaller file sizes
- **Flexible Scaling:** Customizable output resolution and quality
- **High Frame Rates:** Support for up to 60 FPS for smooth animation
- **Dual Processing:** Separate 1080p pipeline optimized for GIF output
- **Instant Preview:** Quick generation with real-time progress feedback

### 📱 WhatsApp Sticker Creation
- **WebP Format:** Uses specialized `gif2webp` tool for maximum WhatsApp compatibility
- **Optimal Dimensions:** Precisely sized to 512x512 pixels with transparent padding
- **Transparent Background:** Clean transparent background for professional stickers
- **Animation Support:** Guaranteed animated stickers that work in WhatsApp conversations
- **Size Optimization:** Advanced compression to stay under 500KB limit (typically ~300KB)
- **Duration Control:** Respects 8-second maximum duration for stickers
- **Frame Rate Optimization:** 6-10 fps specifically tuned for WhatsApp performance

### 🖼️ Image Processing
- **Intelligent Resizing:** Maintains aspect ratio while fitting target dimensions
- **EXIF Timestamp Extraction:** Automatic filename generation from photo metadata
- **Background Compositing:** Centers images on black backgrounds for consistent output
- **Batch Processing:** Handles multiple images with progress indicators
- **Format Optimization:** Separate processing pipelines for video vs. GIF quality

## Requirements

### Core Dependencies
- Go 1.16 or later
- The following Go packages:
  - `github.com/disintegration/imaging`
  - `github.com/rwcarlsen/goexif/exif`
  - `github.com/schollz/progressbar/v3`

### System Requirements
- **FFmpeg**: Required for video and GIF processing
- **WebP tools** (for WhatsApp stickers): Install with `sudo apt install webp` (Ubuntu/Debian) or equivalent

### Platform Support
- Linux (tested on ARM64 and x86_64)
- macOS (Intel and Apple Silicon)
- Windows (x86_64 and ARM64)

## Installation

### System Dependencies

1. **Install FFmpeg**: https://ffmpeg.org/download.html

2. **Install WebP tools** (required for WhatsApp stickers):
   ```sh
   # Ubuntu/Debian
   sudo apt install webp
   
   # macOS with Homebrew
   brew install webp
   
   # Windows: Download from https://developers.google.com/speed/webp/download
   ```

### Optional: Development Setup

For those who want to build from source:

1. Clone the repository:
   ```sh
   git clone https://github.com/yourusername/go24k.git
   cd go24k
   ```

2. Install Go dependencies:
   ```sh
   go mod tidy
   ```

3. Build the executable:
   ```sh
   go build -o go24k
   ```

## Usage

### Basic Usage

1. Place your JPEG images in the current directory.
2. For video generation, also place an MP3 music file in the directory.
3. Run the Go program:

    ```sh
    go run main.go
    ```

4. The processed images will be saved in the `uhd_converted` directory with names that include the image's timestamp.
5. After conversion, a video is generated from the processed images with smooth crossfade transitions and synchronized audio fades.

### Command Line Options

#### General Options
- `--convert-only`: Convert images only, without generating video/GIF
- `-d <seconds>`: Duration per image in seconds (default: 5 for video, 1 for GIF)
- `-t <seconds>`: Transition (fade) duration in seconds (default: 1)

#### Video Options
- `--static`: Use static images without Ken Burns effect

#### GIF Options
- `--gif`: Create regular animated GIF instead of video
- `--gif-optimized`: Create optimized animated GIF with palette (recommended)
- `--gif-total-time <seconds>`: **Total duration of GIF in seconds (recommended method)**
- `--gif-fps <fps>`: Frames per second for GIF (default: 15, up to 60 for smooth animation)
- `--gif-scale <scale>`: Scale factor for GIF output (default: 1.0 = full size)

#### WhatsApp Sticker Options
- `--whatsapp-sticker`: Create WhatsApp sticker (WebP 512x512, transparent background, <8s, <500KB)

### Examples

#### 🎬 Video Creation
**Create a standard UHD video:**
```sh
go run main.go -d 3 -t 1
```

**Create a static slideshow (no Ken Burns effect):**
```sh
go run main.go --static -d 4 -t 2
```

#### 🎞️ GIF Creation

**Quick 5-second GIF (recommended):**
```sh
go run main.go --gif-optimized --gif-total-time 5
```

**Ultra-fast 3-second GIF:**
```sh
go run main.go --gif-optimized --gif-total-time 3 --gif-fps 60
```

**Presentation-style 15-second GIF:**
```sh
go run main.go --gif-optimized --gif-total-time 15
```

**Web-optimized small GIF:**
```sh
go run main.go --gif-optimized --gif-total-time 4 --gif-scale 0.3
```

**Traditional per-image timing:**
```sh
go run main.go --gif-optimized -d 2 --gif-fps 12 --gif-scale 0.5
```

#### 📱 WhatsApp Sticker Creation

**Create animated WhatsApp sticker (default 6-second duration):**
```sh
./go24k --whatsapp-sticker
```

**Create 4-second animated sticker:**
```sh
./go24k --whatsapp-sticker --gif-total-time 4
```

**Maximum 8-second animated sticker:**
```sh
./go24k --whatsapp-sticker --gif-total-time 8
```

> **Note**: WhatsApp stickers use specialized `gif2webp` conversion for guaranteed animation compatibility. The tool automatically optimizes frame rate (6-10 fps) and file size (~300KB) for WhatsApp requirements.

####  Processing Only
**Convert images only (no output generation):**
```sh
./go24k --convert-only
```

## Enhanced User Interface

Go24K features a modern, informative interface with real-time progress tracking and detailed statistics:

### 🎨 Visual Improvements
- **Colorful Progress Indicators**: Emoji-enhanced progress bars with clear section headers
- **Real-time File Information**: Shows current file dimensions and processing status
- **Smart Processing Stats**: Displays speed (images/sec), file sizes, and compression ratios
- **Completion Feedback**: Clear success messages with checkmark confirmations

### 📊 Example Output
```
🎬 Starting UHD Video Conversion
📊 Found 12 images to process
🎯 Target: 4K UHD (3840x2160) with black padding
💾 Output: uhd_converted/ directory

🔄 Converting Kart-01.jpg (4032x4032) 8% |███████████████ (12/12, 2 it/s)
✅ UHD conversion completed!

📈 Conversion Statistics:
   ⏱️  Processing time: 6.5 seconds
   🚀 Average speed: 1.9 images/sec
   📁 Original size: 53.7 MB
   📁 UHD size: 21.1 MB
   📊 Size ratio: 0.4x
```

## Use Cases

### 🎬 **UHD Video Creation**
- **Professional Presentations**: High-quality slideshows with music
- **Social Media Content**: 4K videos for YouTube, Instagram, etc.
- **Photo Montages**: Transform photo collections into cinematic experiences
- **Digital Storytelling**: Create narrative videos from image sequences

### 🎞️ **Animated GIF Generation**
- **Social Media**: Quick, engaging content for Twitter, Discord, Reddit
- **Web Content**: Lightweight animated headers, banners, demos
- **Presentations**: Dynamic slides without video player requirements

### 📱 **WhatsApp Sticker Creation**
- **Personal Stickers**: Transform photo sequences into animated stickers
- **Business Communication**: Create branded animated stickers for WhatsApp Business
- **Creative Expression**: Convert moments into shareable WebP animations
- **Meme Creation**: Quick animated responses and reactions
- **Documentation**: Step-by-step visual guides and tutorials
- **Marketing**: Eye-catching promotional content

### ⚡ **Performance Advantages**
- **Smart Processing**: Separate pipelines optimize quality vs. file size
- **Time Control**: Precise total duration control for GIFs
- **Batch Efficiency**: Process dozens of images in seconds
- **Format Flexibility**: Choose the perfect output for your needs

## Troubleshooting

### WhatsApp Sticker Issues

**Sticker appears static in WhatsApp:**
- Ensure `webp` tools are installed: `sudo apt install webp`
- Use shorter durations (3-4 seconds work best)
- Check file size is under 500KB (automatically optimized)

**File size too large:**
- Reduce duration: `--gif-total-time 3`
- Use fewer source images
- The tool automatically optimizes to ~300KB

### General Issues

**"No .jpg files found" error:**
- Ensure JPEG images are in the current directory
- Check file extensions (must be `.jpg`, not `.jpeg`)
- Verify file permissions

**FFmpeg not found:**
- Install FFmpeg: https://ffmpeg.org/download.html
- Ensure it's in your system PATH
- Test with: `ffmpeg -version`

**Slow processing:**
- Reduce image count for faster processing
- Use SSD storage for better I/O performance
- Check available RAM (processing is memory-intensive)

### Performance Tips

- **For GIFs**: Use `--gif-total-time` instead of per-image duration
- **For Videos**: Use `--static` flag to disable Ken Burns effects for faster processing
- **For Stickers**: Keep duration under 5 seconds for best results

## Building

To build the program for different platforms, you can use the provided [build.sh](#) script:

1. Make the script executable:

    ```sh
    chmod +x build.sh
    ```

2. Run the script to generate builds for Linux, Windows and MacOS (ARM and Intel):

    ```sh
    ./build.sh
    ```

The builds will be saved in the [builds](./builds/) directory.

## File Structure

After running the program, the following directory structure will be created:

```
your-project/
├── *.jpg                   # Original JPEG images (input)
├── *.mp3                   # Music file (input, for video)
├── uhd_converted/          # UHD 4K images for video (3840x2160)
│   └── *_uhd.jpg
├── gif_converted/          # Optimized images for GIF (~1080p)
│   └── 000_*.jpg
├── video.mp4              # Generated UHD video (output)
├── animated.gif           # Regular animated GIF (output)
├── optimized.gif          # Optimized animated GIF (output)
└── go24k_sticker.webp     # WhatsApp animated sticker (output)
```

## Downloads & Releases

Pre-built binaries are available for multiple platforms:

- **Linux** (x86_64, ARM64)
- **Windows** (x86_64, ARM64)  
- **macOS** (Intel, Apple Silicon)

Download the latest release from the [releases page](https://github.com/yourusername/go24k/releases).

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

### Development
1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License. See the LICENSE file for details.

---

**⭐ Star this repository if you find it useful!**