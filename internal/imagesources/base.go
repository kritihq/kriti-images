// package imagesources handles various sources from where an image can be retrieved.
// e.g. local disk, AWS S3 or Cloudflare R2
package imagesources

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
)

// TODO: convert to application configs
const (
	MaxImageDimension = 8192             // 8K max
	MaxImageFileSize  = 50 * 1024 * 1024 // 50MB
)

// ImageSource represents an source to retrieve images from.
type ImageSource interface {
	// GetImage retrieves the image with name `fileName` from the source.
	// If the image is present it is returned as `image.Image` along with its
	// format i.e. extension (JPEG, PNG or WEBP).
	//
	// In case of any error or no image found, `error` is returned and other
	// return values are null and empty.
	GetImage(fileName string) (image.Image, string, error)
}

// TODO: add other S3 compatible sources

// ImageSourceLocal represents the machine's local disk as an image source.
type ImageSourceLocal struct {
	BasePath string // base path of the mounted disk
}

func (i *ImageSourceLocal) GetImage(fileName string) (image.Image, string, error) {
	// Ensure the path is safe and doesn't contain directory traversal
	cleanPath := filepath.Clean(fileName)
	if filepath.IsAbs(cleanPath) || strings.Contains(cleanPath, "..") {
		return nil, "", fmt.Errorf("invalid image path")
	}

	fullPath := filepath.Join(i.BasePath, cleanPath)

	_, err := os.Stat(fullPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to stat image: %w", err)
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	fileStat, _ := file.Stat()
	if err := validateImageSize(fileStat.Size()); err != nil {
		return nil, "", err
	}

	img, format, err := image.Decode(file)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode image: %w", err)
	}

	if err := validateImageDimensions(img.Bounds().Dx(), img.Bounds().Dy()); err != nil {
		return nil, "", err
	}

	return img, format, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Util functions for all image sources
///////////////////////////////////////////////////////////////////////////////////////////////

// validateImageDimensions returns error if the image dimensions exceed max allowed dimensions
func validateImageDimensions(width, height int) error {
	if width > MaxImageDimension || height > MaxImageDimension {
		return fmt.Errorf("image dimensions too large: max allowed is %dx%d", MaxImageDimension, MaxImageDimension)
	}
	return nil
}

// validateImageSize returns error if size of the image file is more than max allowed file size
func validateImageSize(fileSize int64) error {
	if fileSize > MaxImageFileSize {
		return fmt.Errorf("image file too large: max allowed is %d bytes", MaxImageFileSize)
	}
	return nil
}
