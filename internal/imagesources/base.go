// package imagesources handles various sources from where an image can be retrieved.
// e.g. local disk, AWS S3 or Cloudflare R2
package imagesources

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/chai2010/webp"
)

type SourceImageValidations struct {
	MaxImageDimension  int
	MaxFileSizeInBytes int64
}

// ImageSource represents an source to retrieve images from.
type ImageSource interface {
	// GetImage retrieves the image with name `fileName` from the source.
	// If the image is present it is returned as `image.Image` along with its
	// format i.e. extension (JPEG, PNG or WEBP).
	//
	// In case of any error or no image found, `error` is returned and other
	// return values are null and empty.
	GetImage(ctx context.Context, fileName string) (image.Image, string, error)

	// UploadImage uploads the image with name `fileName` to the source.
	// If the image is present it is returned as `image.Image` along with its
	// format i.e. extension (JPEG, PNG or WEBP).
	//
	// In case of any error or no image found, `error` is returned and other
	// return values are null and empty.
	UploadImage(ctx context.Context, fileName string, file image.Image) error
}

// TODO: add other S3 compatible sources

// ImageSourceLocal represents the machine's local disk as an image source.
type ImageSourceLocal struct {
	SourceImageValidations
	BasePath string // base path of the mounted disk
}

func NewImageSourceLocal(basePath string, validations *SourceImageValidations) *ImageSourceLocal {
	return &ImageSourceLocal{
		BasePath:               basePath,
		SourceImageValidations: *validations,
	}
}

func (i *ImageSourceLocal) GetImage(ctx context.Context, fileName string) (image.Image, string, error) {
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
	if err := validateImageSize(fileStat.Size(), i.MaxFileSizeInBytes); err != nil {
		return nil, "", err
	}

	img, format, err := image.Decode(file)
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode image: %w", err)
	}

	if err := validateImageDimensions(img.Bounds().Dx(), img.Bounds().Dy(), i.MaxImageDimension); err != nil {
		return nil, "", err
	}

	return img, format, nil
}

func (i *ImageSourceLocal) UploadImage(ctx context.Context, fileName string, file image.Image) error {
	// Ensure the path is safe and doesn't contain directory traversal
	cleanPath := filepath.Clean(fileName)
	if filepath.IsAbs(cleanPath) || strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid image path")
	}

	// Validate image dimensions
	if err := validateImageDimensions(file.Bounds().Dx(), file.Bounds().Dy(), i.MaxImageDimension); err != nil {
		return err
	}

	// Ensure the directory exists
	fullPath := filepath.Join(i.BasePath, cleanPath)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the file
	outFile, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	// Determine format from file extension
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".jpg", ".jpeg":
		if err := jpeg.Encode(outFile, file, &jpeg.Options{Quality: 85}); err != nil {
			return fmt.Errorf("failed to encode JPEG: %w", err)
		}
	case ".png":
		if err := png.Encode(outFile, file); err != nil {
			return fmt.Errorf("failed to encode PNG: %w", err)
		}
	case ".webp":
		if err := webp.Encode(outFile, file, &webp.Options{Quality: 85}); err != nil {
			return fmt.Errorf("failed to encode WebP: %w", err)
		}
	default:
		return fmt.Errorf("unsupported image format: %s", ext)
	}

	// Validate file size after encoding
	fileInfo, err := outFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	if err := validateImageSize(fileInfo.Size(), i.MaxFileSizeInBytes); err != nil {
		// Remove the file if it exceeds size limit
		os.Remove(fullPath)
		return err
	}

	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////
// Util functions for all image sources
///////////////////////////////////////////////////////////////////////////////////////////////

// validateImageDimensions returns error if the image dimensions exceed max allowed dimensions
func validateImageDimensions(width, height, max int) error {
	if width > max || height > max {
		return fmt.Errorf("image dimensions too large: max allowed is %dx%d", max, max)
	}
	return nil
}

// validateImageSize returns error if size of the image file is more than max allowed file size
func validateImageSize(fileSize, max int64) error {
	if fileSize > max {
		return fmt.Errorf("image file too large: max allowed is %d bytes", max)
	}
	return nil
}
