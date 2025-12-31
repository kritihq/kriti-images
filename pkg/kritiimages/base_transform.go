package kritiimages

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"strings"

	"github.com/chai2010/webp"
	"github.com/disintegration/gift"
)

var (
	ErrTransformationsNotFound = errors.New("failed to get transformations")
)

// Transform transforms an image from a given source into a desired output format.
// It takes a context.Context, a path string, a destination image pointer, and a map of transformation options.
// Returns a bytes.Buffer pointer and an error.
func (k *KritiImages) Transform(ctx context.Context, path string, dest *DestinationImage, options map[TransformationOption]string) (*bytes.Buffer, error) {
	source := k.getImageSource(path)
	img, imgFormat, err := source.GetImage(ctx, path)
	if err != nil {
		return nil, ErrSourceImageNotFound
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
		return nil, errors.Join(ErrTransformationsNotFound, err)
	}
	g := gift.New(filters...)

	// create destination image
	dstBounds := g.Bounds(img.Bounds())
	dst := image.NewRGBA(dstBounds)

	// apply background color if needed
	if dest.BgColor != color.Transparent {
		bounds := dst.Bounds()
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				dst.Set(x, y, dest.BgColor)
			}
		}
	}

	// apply transformations
	g.Draw(dst, img)

	// encode output using format from transformation context
	return k.formatTo(dst, dest.Format, dest.Quality)
}

func (k *KritiImages) getImageSource(path string) ImageSource {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return k.ImageSources["http"]
	} else {
		return k.DefaultImageSource
	}
}

func (k *KritiImages) formatTo(image image.Image, format string, quality int) (*bytes.Buffer, error) {
	out := new(bytes.Buffer)

	switch strings.ToLower(format) {
	case "jpg", "jpeg":
		if err := jpeg.Encode(out, image, &jpeg.Options{Quality: quality}); err != nil {
			return nil, errors.Join(ErrFailedToEncodeImage, err)
		}
	case "png":
		if err := png.Encode(out, image); err != nil {
			return nil, errors.Join(ErrFailedToEncodeImage, err)
		}
	case "webp":
		if err := webp.Encode(out, image, &webp.Options{Quality: float32(quality)}); err != nil {
			return nil, errors.Join(ErrFailedToEncodeImage, err)
		}
	default:
		return nil, ErrInvalidImageFormat
	}

	return out, nil
}
