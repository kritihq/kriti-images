package kritiimages

import (
	"errors"
	"image/color"
)

var (
	ErrInvalidImageFormat  = errors.New("unsupported image format")
	ErrFailedToEncodeImage = errors.New("failed to encode image to provided format")
	ErrSourceImageNotFound = errors.New("source image not found")
	ErInvalidImageSources  = errors.New("invalid imagesource instance provided")
)

// DestinationImage represents the desired output image properties.
type DestinationImage struct {
	BgColor color.Color
	Width   int
	Height  int
	Format  string
	Quality int // lossy quality for JPEG & WEBP, 1 to 100, higher is better
}

// New creates a new instance of KritiImages.
// It requires a map of ImageSource instances and a default ImageSource instance along with default TemplateSource instance.
//
// Program will panic if the provided map of ImageSource instances is empty or if the default ImageSource instance is nil.
func New(sources map[string]ImageSource, defaultSrc ImageSource, templSource TemplateSource) *KritiImages {
	if len(sources) == 0 {
		panic("no imagesources provided")
	} else if defaultSrc == nil {
		panic("default imagesource can not be nil")
	}

	if templSource == nil {
		panic("default templatesource can not be nil")
	}

	return &KritiImages{
		DefaultImageSource:     defaultSrc,
		ImageSources:           sources,
		DefaultTemplateSources: templSource,
	}
}

// KritiImages represents a collection of ImageSource instances.
// It provides methods to transform images from various sources into a desired output format.
type KritiImages struct {
	DefaultImageSource     ImageSource
	ImageSources           map[string]ImageSource
	DefaultTemplateSources TemplateSource
}

// refer to base_transform & base_template files
