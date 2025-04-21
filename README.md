# rmbg - Background Removal Tool

A command-line tool for removing backgrounds from images using the [remove.bg](https://www.remove.bg/) API.

## Disclaimer

This tool is provided as-is and is not affiliated with [remove.bg](remove.bg) in any way.

## Features

- Remove backgrounds from single images or entire directories
- Support for JPEG, PNG, and WebP input formats
- Option to save output as PNG or WebP
- Compression settings with customizable quality
- Optimized file size while preserving transparency

## Installation

### Prerequisites

- Go 1.16 or higher
- [remove.bg API key](https://www.remove.bg/api)
- [libvips](https://github.com/libvips/libvips) (required by the bimg dependency)

### Installing libvips

#### macOS

```bash
brew install vips
```

#### Ubuntu/Debian

```bash
apt-get update && apt-get install -y libvips-dev
```

#### Other platforms

See the [libvips installation instructions](https://github.com/libvips/libvips/wiki/Installation).

### Building from source

1. Clone the repository

```bash
git clone <repository-url>
cd rmbg
```

2. Build the application

```bash
make
```

3. Install to your bin directory (optional)

```bash
make install
```

By default, this will install to `/usr/local/bin`. To change the installation directory:

```bash
make install BINDIR=~/bin
```

## Usage

First, set your remove.bg API key as an environment variable:

```bash
export REMOVE_BG_API_KEY=your-api-key
```

### Basic usage

Process a single image:

```bash
rmbg image.jpg
```

Process a directory of images:

```bash
rmbg images/
```

### Output format

Save as WebP (smaller file size):

```bash
rmbg -f webp image.jpg
```

### Compression

Enable compression with default quality (90):

```bash
rmbg -c image.jpg
```

Specify compression quality (1-100, higher is better quality):

```bash
rmbg -c 75 image.jpg
# OR
rmbg -c=75 image.jpg
```

### Custom output path

Process a single image with a custom output filename:

```bash
rmbg image.jpg custom-output.png
```

### Combining options

Process a directory, save as WebP with quality 80:

```bash
rmbg -f webp -c 80 images/
```

### Help

Display help information:

```bash
rmbg -h
```

## License

[MIT License](LICENSE)
