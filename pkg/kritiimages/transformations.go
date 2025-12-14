package kritiimages

import (
	"fmt"
	"image"
	"image/color"

	"github.com/disintegration/gift"
	"github.com/gofiber/fiber/v2/log"
	"github.com/kritihq/kriti-images/internal/transformations"
	"github.com/kritihq/kriti-images/internal/utils"
)

// Supported transformations
type TransformationOption int

const (
	Flip TransformationOption = iota + 1
	Blur
	Brightness
	Contrast
	Fit
	// The gamma transformation works differently from brightness - while brightness shifts all pixel values equally, gamma correction applies a curve that affects midtones more than highlights and shadows.
	Gamma
	Rotate
	Saturation
	Sharpen
	Background
	Width
	Height
	Format
	Quality
	BorderRadius
)

type DestinationImage struct {
	BgColor color.Color
	Width   int
	Height  int
	Format  string
	Quality int // lossy quality for JPEG & WEBP, 1 to 100 higher is better
}

func getFilters(options map[TransformationOption]string, destination *DestinationImage) ([]gift.Filter, error) {
	filters := make([]gift.Filter, 0)

	// Check if we have dimensions but no fit parameter
	hasDimensions := destination.Width > 0 || destination.Height > 0
	_, hasFit := options[Fit]

	// If we have dimensions but no explicit fit, add default "contain" behavior
	if hasDimensions && !hasFit {
		fitFilter, err := transformations.CreateFitFilter("crop", destination.Width, destination.Height, destination.BgColor)
		if err != nil {
			return nil, fmt.Errorf("failed to create default fit filter: %w", err)
		}
		if fitFilter != nil {
			filters = append(filters, fitFilter)
		}
	}

	for t, values := range options {
		switch t {
		case Flip:
			switch values {
			case "h":
				filters = append(filters, gift.FlipHorizontal())
			case "v":
				filters = append(filters, gift.FlipVertical())
			case "hv", "vh":
				filters = append(filters, gift.FlipHorizontal(), gift.FlipVertical())
			}
		case Blur:
			strength := utils.ParseFloatValue(values, 1, 250, 1)
			filters = append(filters, gift.GaussianBlur(strength))
		case Brightness:
			strengthPct := utils.ParseFloatValue(values, -100, 100, 0)
			filters = append(filters, gift.Brightness(strengthPct))
		case Contrast:
			strengthPct := utils.ParseFloatValue(values, -100, 100, 0)
			filters = append(filters, gift.Contrast(strengthPct))
		case Fit:
			fitFilter, err := transformations.CreateFitFilter(values, destination.Width, destination.Height, destination.BgColor)
			if err != nil {
				return nil, fmt.Errorf("failed to create fit filter: %w", err)
			}
			if fitFilter != nil {
				filters = append(filters, fitFilter)
			}
		case Gamma:
			strength := utils.ParseFloatValue(values, 0, 2.0, 0)
			filters = append(filters, gift.Gamma(strength))
		case Rotate:
			rotateAngle, err := utils.ParseRotateAngle(values)
			if err != nil {
				return nil, fmt.Errorf("invalid rotate value: %w", err)
			}
			filters = append(filters, gift.Rotate(rotateAngle, image.Transparent, gift.LinearInterpolation))
		case Saturation:
			strengthPct := utils.ParseFloatValue(values, -100, 500, 0)
			filters = append(filters, gift.Saturation(strengthPct))
		case Sharpen:
			strength := utils.ParseFloatValue(values, 0.5, 1.5, 0.5)
			filters = append(filters, gift.UnsharpMask(1.0, strength, 0.0))
		case BorderRadius:
			radiusFilter, err := transformations.CreateBorderRadiusFilter(values)
			if err != nil {
				return nil, fmt.Errorf("failed to create border radius filter: %w", err)
			}
			if radiusFilter != nil {
				filters = append(filters, radiusFilter)
			}
		default:
			log.Warnf("unkonwn transformation option: %v", t)
		}
	}

	return filters, nil
}
