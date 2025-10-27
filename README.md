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

### 🖼️ Image Processing
- **Intelligent Resizing:** Maintains aspect ratio while fitting target dimensions
- **EXIF Timestamp Extraction:** Automatic filename generation from photo metadata
- **Background Compositing:** Centers images on black backgrounds for consistent output
- **Batch Processing:** Handles multiple images with progress indicators
- **Format Optimization:** Separate processing pipelines for video vs. GIF quality

## Requirements

- Go 1.16 or later
- The following Go packages:
  - `github.com/disintegration/imaging`
  - `github.com/rwcarlsen/goexif/exif`
  - `github.com/schollz/progressbar/v3`
- FFMpeg

## Installation

1. Install FFmpeg: https://ffmpeg.org/download.html

- Optional steps, only for those who wants to code:
2. Clone the repository:

    ```sh
    git clone https://github.com/yourusername/go24k.git
    cd go24k
    ```

2. Install the required Go packages:

    ```sh
    go get github.com/disintegration/imaging
    go get github.com/rwcarlsen/goexif/exif
    go get github.com/schollz/progressbar/v3
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

#### 🔧 Processing Only
**Convert images only (no output generation):**
```sh
go run main.go --convert-only
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
- **Documentation**: Step-by-step visual guides and tutorials
- **Marketing**: Eye-catching promotional content

### ⚡ **Performance Advantages**
- **Smart Processing**: Separate pipelines optimize quality vs. file size
- **Time Control**: Precise total duration control for GIFs
- **Batch Efficiency**: Process dozens of images in seconds
- **Format Flexibility**: Choose the perfect output for your needs

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
└── optimized.gif          # Optimized animated GIF (output)
```

## License

This project is licensed under the MIT License. See the LICENSE file for details.