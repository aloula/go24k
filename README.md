# Go24K

Go24K is a Go program that processes JPEG images in the current directory, resizes them to a height of 2160 pixels, and composites them onto a black background image with dimensions 3840x2160 pixels (4K - UHD format). The processed images are saved with a new name that includes the timestamp from the image's EXIF data. The UHD resized images are used to create a MP4 Video (H.264 codec) with music and crossfade between each image.

## Features

- **Image Conversion:** Resizes JPEG images to a height of 2160 pixels while maintaining the aspect ratio.
- **Image Compositing:** Composites the resized images onto a black background image (3840x2160 pixels).
- **Timestamp Naming:** Saves processed images with a new name that includes the timestamp from the image's EXIF data in `YYYYMMDD_HHMMSS` format.
- **Video Generation:** Generates a video from converted images with crossfade transitions and audio fades.

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

1. Place your JPEG images and MP3 music in the current directory.
2. Run the Go program:

    ```sh
    go run go24k.go
    ```

3. The processed images will be saved in the `converted` directory with names that include the image's timestamp.
4. After conversion, a video is generated from the processed images with smooth crossfade transitions and synchronized audio fades. Progress bars provide animated feedback during these stages.

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

## License

This project is licensed under the MIT License. See the LICENSE file for details.