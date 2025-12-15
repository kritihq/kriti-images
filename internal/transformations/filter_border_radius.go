package transformations

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/disintegration/gift"
	"github.com/kritihq/kriti-images/internal/utils"
)

// createBorderRadiusFilter creates a filter that applies rounded corners to an image
func CreateBorderRadiusFilter(value string) (gift.Filter, error) {
	// Validate that value is not empty
	if value == "" {
		return nil, fmt.Errorf("border radius value cannot be empty")
	}

	// Parse the border radius value with proper validation
	radii, err := utils.ParseBorderRadiusValue(value)
	if err != nil {
		return nil, err
	}

	return &borderRadiusFilter{
		tl: radii,
		tr: radii,
		bl: radii,
		br: radii,
	}, nil
}

// borderRadiusFilter applies rounded corners to an image
type borderRadiusFilter struct {
	tl *utils.BorderRadiusValue // top-left in px
	tr *utils.BorderRadiusValue // top-right
	bl *utils.BorderRadiusValue // bottom-left
	br *utils.BorderRadiusValue // bottom-right
}

func (f *borderRadiusFilter) Bounds(srcBounds image.Rectangle) image.Rectangle {
	return srcBounds
}

func (f *borderRadiusFilter) Draw(dst draw.Image, src image.Image, options *gift.Options) {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	maxRadius := float32(math.Min(float64(width), float64(height))) / 2
	minDim := float32(math.Min(float64(width), float64(height)))
	if f.tl.IsPercent {
		f.tl.Value = (f.tl.Value / 100.0) * minDim // convert to px
		f.tl.IsPercent = false
	}
	f.tl.Value = float32(math.Min(float64(f.tl.Value), float64(maxRadius))) // clamp to max
	if f.tr.IsPercent {
		f.tr.Value = (f.tr.Value / 100.0) * minDim
		f.tr.IsPercent = false
	}
	f.tr.Value = float32(math.Min(float64(f.tr.Value), float64(maxRadius)))
	if f.bl.IsPercent {
		f.bl.Value = (f.bl.Value / 100.0) * minDim
		f.bl.IsPercent = false
	}
	f.bl.Value = float32(math.Min(float64(f.bl.Value), float64(maxRadius)))
	if f.br.IsPercent {
		f.br.Value = (f.br.Value / 100.0) * minDim
		f.br.IsPercent = false
	}
	f.br.Value = float32(math.Min(float64(f.br.Value), float64(maxRadius)))

	// Create a mask for the rounded rectangle
	mask := image.NewAlpha(bounds)

	// Fill the mask with the rounded rectangle shape
	f.drawRoundedRectMask(mask, bounds)

	// Apply the source image with the mask
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Get the alpha value from our mask
			maskAlpha := mask.AlphaAt(x, y).A

			if maskAlpha > 0 {
				// Get the source pixel
				srcColor := src.At(x, y)

				// If the mask is partially transparent, blend with transparency
				if maskAlpha < 255 {
					r, g, b, a := srcColor.RGBA()
					// Apply mask alpha
					newAlpha := uint8((uint32(maskAlpha) * (a >> 8)) >> 8)
					dst.Set(x, y, color.RGBA{
						uint8(r >> 8),
						uint8(g >> 8),
						uint8(b >> 8),
						newAlpha,
					})
				} else {
					dst.Set(x, y, srcColor)
				}
			} else {
				// Outside the rounded rectangle, set to transparent
				dst.Set(x, y, color.RGBA{0, 0, 0, 0})
			}
		}
	}
}

// drawRoundedRectMask draws a rounded rectangle mask
func (f *borderRadiusFilter) drawRoundedRectMask(mask *image.Alpha, bounds image.Rectangle) {
	width := bounds.Dx()
	height := bounds.Dy()

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			alpha := f.calculatePixelAlpha(x, y, width, height)
			mask.SetAlpha(x, y, color.Alpha{alpha})
		}
	}
}

// calculatePixelAlpha determines the alpha value for a pixel in the rounded rectangle
func (f *borderRadiusFilter) calculatePixelAlpha(x, y, width, height int) uint8 {
	fx := float32(x)
	fy := float32(y)
	fw := float32(width)
	fh := float32(height)

	// Determine which corner region we're in
	var radius float32
	var centerX, centerY float32

	// Top-left corner
	if fx < f.tl.Value && fy < f.tl.Value {
		radius = f.tl.Value
		centerX = f.tl.Value
		centerY = f.tl.Value
	} else if fx >= fw-f.tr.Value && fy < f.tr.Value {
		// Top-right corner
		radius = f.tr.Value
		centerX = fw - f.tr.Value
		centerY = f.tr.Value
	} else if fx >= fw-f.br.Value && fy >= fh-f.br.Value {
		// Bottom-right corner
		radius = f.br.Value
		centerX = fw - f.br.Value
		centerY = fh - f.br.Value
	} else if fx < f.bl.Value && fy >= fh-f.bl.Value {
		// Bottom-left corner
		radius = f.bl.Value
		centerX = f.bl.Value
		centerY = fh - f.bl.Value
	} else {
		// Not in a corner region, fully opaque
		return 255
	}

	// Calculate distance from corner center
	dx := fx - centerX
	dy := fy - centerY
	distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))

	// If we're inside the radius, it's opaque
	if distance <= radius {
		// Add some anti-aliasing for smoother edges
		if distance >= radius-1 {
			// Linear interpolation for the edge pixel
			alpha := 1.0 - (distance - (radius - 1.0))
			return uint8(alpha * 255)
		}
		return 255
	}

	// Outside the radius, transparent
	return 0
}
