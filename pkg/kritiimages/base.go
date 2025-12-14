package kritiimages

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"strings"

	"github.com/chai2010/webp"
	"github.com/disintegration/gift"
)

type KritiImages struct {
	DefaultSource ImageSource
	Sources       map[string]ImageSource
}

func New(sources map[string]ImageSource, defaultSrc ImageSource) *KritiImages {
	return &KritiImages{DefaultSource: defaultSrc, Sources: sources}
}

func (k *KritiImages) Transform(ctx context.Context, path string, dest *DestinationImage, options map[TransformationOption]string) (*bytes.Buffer, error) {
	source := k.getImageSource(path)
	img, imgFormat, err := source.GetImage(ctx, path)
	if err != nil {
		return nil, errors.New("source image not found")
	}

	// set default values if not present
	if dest.Width <= 0 {
		dest.Width = img.Bounds().Dx()
	}
	if dest.Height <= 0 {
		dest.Height = img.Bounds().Dy()
	}
	if dest.Format == "" {
		dest.Format = imgFormat
	}
	if dest.Quality <= 0 {
		dest.Quality = 100
	}

	filters, err := getFilters(options, dest)
	if err != nil {
		return nil, fmt.Errorf("failed to get filters: %w", err)
	}
	g := gift.New(filters...)

	// Create destination image
	dstBounds := g.Bounds(img.Bounds())
	dst := image.NewRGBA(dstBounds)

	// Apply background color if needed
	if dest.BgColor != color.Transparent {
		bounds := dst.Bounds()
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				dst.Set(x, y, dest.BgColor)
			}
		}
	}

	// Apply transformations
	g.Draw(dst, img)

	// Encode output using format from transformation context
	return k.formatTo(dst, dest.Format, dest.Quality)
}

func (k *KritiImages) getImageSource(path string) ImageSource {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return k.Sources["http"]
	} else {
		return k.DefaultSource
	}
}

func (k *KritiImages) formatTo(image image.Image, format string, quality int) (*bytes.Buffer, error) {
	out := new(bytes.Buffer)

	switch strings.ToLower(format) {
	case "jpg", "jpeg":
		if err := jpeg.Encode(out, image, &jpeg.Options{Quality: quality}); err != nil {
			return nil, errors.New("failed to encode image")
		}
	case "png":
		if err := png.Encode(out, image); err != nil {
			return nil, errors.New("failed to encode image")
		}
	case "webp":
		if err := webp.Encode(out, image, &webp.Options{Quality: float32(quality)}); err != nil {
			return nil, errors.New("failed to encode image")
		}
	default:
		return nil, errors.New("unsupported format")
	}

	return out, nil
}
