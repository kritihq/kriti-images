# Kriti Images

A high-performance image transformation service built in Go, providing a URL-based API for real-time image processing. An open-source alternative to Cloudflare Images and ImageKit.

## 🚀 Features

- **URL-based transformations** - Transform images through simple URL parameters
- **Multiple formats** - Support for JPEG, PNG, and WebP
- **Rich transformations** - Resize, crop, rotate, blur, adjust brightness/contrast, and more
- **Smart resizing modes** - Contains, cover, crop, pad, squeeze, and scale-down options
- **Color adjustments** - Brightness, contrast, saturation, gamma correction
- **Background colors** - Support for hex, RGB, and named colors
- **High performance** - Built with Go and optimized for speed
- **CDN-friendly** - Proper caching headers for optimal CDN integration

## 📖 Quick Example

Transform an image by modifying the URL:

```bash
# Original image
GET /cgi/images/tr:quality=100/image1.jpg

# Resize to 300px width, auto height
GET /cgi/images/tr:width=300/image1.jpg

# Resize with specific fit mode and background
GET /cgi/images/tr:width=400,height=300,fit=pad,background=blue/image1.jpg

# Multiple transformations
GET /cgi/images/tr:width=500,brightness=20,contrast=30,format=webp/image1.jpg

# Blur with rotation
GET /cgi/images/tr:blur=10,rotate=45,background=white/image1.jpg
```

## 🛠 Supported Transformations

### Resize & Cropping
- `width` - Set image width (1-10000px)
- `height` - Set image height (1-10000px)
- `fit` - Resize behavior: `contain`, `cover`, `crop`, `pad`, `squeeze`, `scaledown`

### Image Adjustments
- `brightness` - Adjust brightness (-100 to 100)
- `contrast` - Adjust contrast (-100 to 100)
- `saturation` - Adjust color saturation (-100 to 500)
- `gamma` - Gamma correction (0.1 to 2.0)
- `blur` - Gaussian blur (1 to 250)
- `sharpen` - Unsharp mask sharpening (0.5 to 1.5)

### Rotation & Flipping
- `rotate` - Rotate image (0-360° or shortcuts: `90`, `cw`, `180`, `270`, `ccw`)
- `flip` - Flip image (`h` for horizontal, `v` for vertical, `hv` for both)

### Format & Quality
- `format` - Output format (`jpeg`, `png`, `webp`)
- `quality` - JPEG/WebP quality (1-100, higher = better quality)
- `background` - Background color (hex: `#ff0000`, named: `red`, rgb: `rgb(255,0,0)`)

## 🏗 Build & Run

### Prerequisites
- Go 1.24.6 or later
- Git

### Development Setup

1. **Prepare sample images**
   ```bash
   mkdir -p web/static/assets
   # Add your test images to web/static/assets/
   ```

2. **Run the server**
   ```bash
   go run main.go
   ```

3. **Access the service**
   - API: `http://localhost:8080/cgi/images/tr:<transformations>/<image-name>`
   - Demo page: `http://localhost:8080/demo`

### Production Build

```bash
# Build binary
go build -o kriti-images main.go

# Run with production settings
./kriti-images
```

## 🔧 Configuration

The service uses external configuration files with the [Viper](https://github.com/spf13/viper) library. Configuration can be provided in YAML or TOML format.

### Configuration Files

Create a `config.yaml` or `config.toml` file in the project root.

### Configuration Options

- **server.port** - Server port (default: 8080)
- **server.enable_print_routes** - Enable route debugging (default: false)
- **server.read_timeout** - Request read timeout (default: 30s)
- **server.write_timeout** - Response write timeout (default: 30s)
- **images.base_path** - Image source directory (default: "web/static/assets")
- **images.max_image_dimension** - Maximum image dimension, any source image beyond will not be processed (default: 8192 (8K))
- **images.max_file_size_in_bytes** - Maximum image file size, any source image beyond will not be processed (default: 52428800 (50MB))
- **limiter.max** - Rate limit per minute (default: 100)
- **limiter.expiration** - Rate limit window (default: 1m)

## 🌐 API Reference

### Base URL Structure
```
/cgi/images/tr:<transformations>/<image-path>
```

### Transformation Syntax
Multiple transformations are separated by commas:
```
tr:width=300,height=200,format=webp,quality=80
```

## 🎨 Interactive Demo

Visit `/demo` to see an interactive demonstration of all available transformations with live examples.

## 🚦 Health & Monitoring

- **Health Check**: `GET /health/ready` - Returns 200 when service is ready
- **Liveness Check**: `GET /health/live` - Returns 200 when service is alive
- **Metrics**: `GET /metrics` - Prometheus-compatible metrics


## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.
