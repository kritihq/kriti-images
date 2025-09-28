package transformations

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"slices"
	"strings"

	"github.com/disintegration/gift"
)

func createFitFilter(value string, destination *DestinationImage) (gift.Filter, error) {
	// Format: just the mode name (e.g., "contain", "cover", "scaledown")
	mode := strings.TrimSpace(value)

	// Validate mode
	validModes := []string{"scaledown", "contain", "cover", "crop", "pad", "squeeze"}
	if !slices.Contains(validModes, mode) {
		return nil, fmt.Errorf("invalid fit mode: %s. Valid modes are: %s", mode, strings.Join(validModes, ", "))
	}

	width := destination.Width
	height := destination.Height

	// If no dimensions are set, return an error
	if width == 0 && height == 0 {
		return nil, fmt.Errorf("width and/or height must be specified for fit operation")
	}

	switch mode {
	case "contain":
		if width > 0 && height > 0 {
			return gift.ResizeToFit(width, height, gift.LanczosResampling), nil
		} else if width > 0 {
			return gift.Resize(width, 0, gift.LanczosResampling), nil
		} else if height > 0 {
			return gift.Resize(0, height, gift.LanczosResampling), nil
		}

	case "cover":
		if width > 0 && height > 0 {
			return gift.ResizeToFill(width, height, gift.LanczosResampling, gift.CenterAnchor), nil
		}
		return nil, fmt.Errorf("cover mode requires both width and height")

	case "squeeze":
		if width > 0 && height > 0 {
			return gift.Resize(width, height, gift.LanczosResampling), nil
		}
		return nil, fmt.Errorf("squeeze mode requires both width and height")

	case "scaledown":
		if width > 0 && height > 0 {
			return &scaleDownFilter{width: width, height: height}, nil
		} else if width > 0 {
			return &scaleDownFilter{width: width, height: 0}, nil
		} else if height > 0 {
			return &scaleDownFilter{width: 0, height: height}, nil
		}

	case "crop":
		if width > 0 && height > 0 {
			return &cropFilter{width: width, height: height}, nil
		}
		return nil, fmt.Errorf("crop mode requires both width and height")

	case "pad":
		if width > 0 && height > 0 {
			return &padFilter{width: width, height: height, bgColor: destination.BgColor}, nil
		}
		return nil, fmt.Errorf("pad mode requires both width and height")
	}

	return nil, fmt.Errorf("unsupported fit mode: %s", mode)
}

// Custom filter for scaledown mode
type scaleDownFilter struct {
	width, height int
}

func (f *scaleDownFilter) Bounds(srcBounds image.Rectangle) image.Rectangle {
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	if f.width == 0 && f.height == 0 {
		return srcBounds
	}

	var newW, newH int
	if f.width > 0 && f.height > 0 {
		// Both dimensions specified
		if srcW <= f.width && srcH <= f.height {
			return srcBounds // Don't scale up
		}
		// Calculate aspect-preserving dimensions
		scaleW := float64(f.width) / float64(srcW)
		scaleH := float64(f.height) / float64(srcH)
		scale := math.Min(scaleW, scaleH)
		newW = int(float64(srcW) * scale)
		newH = int(float64(srcH) * scale)
	} else if f.width > 0 {
		// Only width specified
		if srcW <= f.width {
			return srcBounds // Don't scale up
		}
		scale := float64(f.width) / float64(srcW)
		newW = f.width
		newH = int(float64(srcH) * scale)
	} else {
		// Only height specified
		if srcH <= f.height {
			return srcBounds // Don't scale up
		}
		scale := float64(f.height) / float64(srcH)
		newW = int(float64(srcW) * scale)
		newH = f.height
	}

	return image.Rect(0, 0, newW, newH)
}

func (f *scaleDownFilter) Draw(dst draw.Image, src image.Image, options *gift.Options) {
	bounds := f.Bounds(src.Bounds())
	if bounds == src.Bounds() {
		// No scaling needed, just copy
		gift.New().Draw(dst, src)
	} else {
		// Scale down
		resizeFilter := gift.Resize(bounds.Dx(), bounds.Dy(), gift.LanczosResampling)
		resizeFilter.Draw(dst, src, options)
	}
}

// Custom filter for crop mode
type cropFilter struct {
	width, height int
}

func (f *cropFilter) Bounds(srcBounds image.Rectangle) image.Rectangle {
	return image.Rect(0, 0, f.width, f.height)
}

func (f *cropFilter) Draw(dst draw.Image, src image.Image, options *gift.Options) {
	srcW := src.Bounds().Dx()
	srcH := src.Bounds().Dy()

	// If source is smaller than target, behave like scaledown
	if srcW <= f.width && srcH <= f.height {
		scaleDownFilter := &scaleDownFilter{width: f.width, height: f.height}
		scaleDownFilter.Draw(dst, src, options)
	} else {
		// Behave like cover
		fillFilter := gift.ResizeToFill(f.width, f.height, gift.LanczosResampling, gift.CenterAnchor)
		fillFilter.Draw(dst, src, options)
	}
}

// Custom filter for pad mode
type padFilter struct {
	width, height int
	bgColor       color.Color
}

func (f *padFilter) Bounds(srcBounds image.Rectangle) image.Rectangle {
	return image.Rect(0, 0, f.width, f.height)
}

func (f *padFilter) Draw(dst draw.Image, src image.Image, options *gift.Options) {
	// First resize to fit within bounds
	fitFilter := gift.ResizeToFit(f.width, f.height, gift.LanczosResampling)
	tempBounds := fitFilter.Bounds(src.Bounds())
	temp := image.NewRGBA(tempBounds)
	fitFilter.Draw(temp, src, options)

	// Fill destination with background color
	bgColor := f.bgColor
	if bgColor == nil {
		bgColor = color.White // Fallback to white
	}

	// Fill destination with white background
	for y := dst.Bounds().Min.Y; y < dst.Bounds().Max.Y; y++ {
		for x := dst.Bounds().Min.X; x < dst.Bounds().Max.X; x++ {
			dst.Set(x, y, bgColor)
		}
	}

	// Center the resized image
	offsetX := (f.width - tempBounds.Dx()) / 2
	offsetY := (f.height - tempBounds.Dy()) / 2

	for y := 0; y < tempBounds.Dy(); y++ {
		for x := 0; x < tempBounds.Dx(); x++ {
			dst.Set(x+offsetX, y+offsetY, temp.At(x, y))
		}
	}
}
