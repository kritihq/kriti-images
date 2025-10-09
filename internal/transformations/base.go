package transformations

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/disintegration/gift"
	"github.com/gofiber/fiber/v2/log"
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

type TransformationCtx struct {
	Filters          []gift.Filter
	DestinationImage *DestinationImage
}

func GetContextFromString(optionsStr string, srcImg image.Image, srcFormat string) (*TransformationCtx, error) {
	options := strings.Split(optionsStr, ",")

	destination := DestinationImage{
		BgColor: color.Transparent,
		Width:   srcImg.Bounds().Dx(),
		Height:  srcImg.Bounds().Dy(),
		Format:  srcFormat,
		Quality: 100,
	}

	trValues := make(map[TransformationOption]string)
	for _, optStr := range options {
		transformation, values, err := processOption(optStr)
		if err != nil {
			return nil, err
		}

		switch transformation {
		case Background:
			destination.BgColor, err = parseBackgroundColor(values)
			if err != nil {
				return nil, fmt.Errorf("invalid background color: %w", err)
			}
		case Width:
			destination.Width, err = parseIntValue(values, 1, 10000)
			if err != nil {
				return nil, fmt.Errorf("invalid width: %w", err)
			}
		case Height:
			destination.Height, err = parseIntValue(values, 1, 10000)
			if err != nil {
				return nil, fmt.Errorf("invalid height: %w", err)
			}
		case Format:
			destination.Format, err = parseFormatValue(values)
			if err != nil {
				return nil, fmt.Errorf("invalid format: %w", err)
			}
		case Quality:
			destination.Quality, err = parseIntValue(values, 1, 100)
			if err != nil {
				return nil, fmt.Errorf("invalid quality: %w", err)
			}
		default:
			trValues[transformation] = values
		}
	}

	filters, err := createFilters(trValues, &destination)
	if err != nil {
		return nil, fmt.Errorf("failed to create filters: %w", err)
	}

	return &TransformationCtx{Filters: filters, DestinationImage: &destination}, nil
}

func processOption(optStr string) (TransformationOption, string, error) {
	parts := strings.Split(optStr, "=")
	if len(parts) != 2 {
		return -1, "", fmt.Errorf("invalid option format: %s", optStr)
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	switch key {
	case "flip":
		return Flip, value, nil
	case "blur":
		return Blur, value, nil
	case "brightness":
		return Brightness, value, nil
	case "contrast":
		return Contrast, value, nil
	case "fit":
		return Fit, value, nil
	case "gamma":
		return Gamma, value, nil
	case "rotate":
		return Rotate, value, nil
	case "saturation":
		return Saturation, value, nil
	case "sharpen":
		return Sharpen, value, nil
	case "background":
		return Background, value, nil
	case "width":
		return Width, value, nil
	case "height":
		return Height, value, nil
	case "format":
		return Format, value, nil
	case "quality":
		return Quality, value, nil
	case "radius":
		return BorderRadius, value, nil
	default:
		return -1, "", fmt.Errorf("unknown option: %s", key)
	}
}

func createFilters(transformationsAndValues map[TransformationOption]string, destination *DestinationImage) ([]gift.Filter, error) {
	filters := make([]gift.Filter, 0)

	// Check if we have dimensions but no fit parameter
	hasDimensions := destination.Width > 0 || destination.Height > 0
	_, hasFit := transformationsAndValues[Fit]

	// If we have dimensions but no explicit fit, add default "contain" behavior
	if hasDimensions && !hasFit {
		fitFilter, err := createFitFilter("crop", destination)
		if err != nil {
			return nil, fmt.Errorf("failed to create default fit filter: %w", err)
		}
		if fitFilter != nil {
			filters = append(filters, fitFilter)
		}
	}

	for t, values := range transformationsAndValues {
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
			strength := parseFloatValue(values, 1, 250, 1)
			filters = append(filters, gift.GaussianBlur(strength))
		case Brightness:
			strengthPct := parseFloatValue(values, -100, 100, 0)
			filters = append(filters, gift.Brightness(strengthPct))
		case Contrast:
			strengthPct := parseFloatValue(values, -100, 100, 0)
			filters = append(filters, gift.Contrast(strengthPct))
		case Fit:
			fitFilter, err := createFitFilter(values, destination)
			if err != nil {
				return nil, fmt.Errorf("failed to create fit filter: %w", err)
			}
			if fitFilter != nil {
				filters = append(filters, fitFilter)
			}
		case Gamma:
			strength := parseFloatValue(values, 0, 2.0, 0)
			filters = append(filters, gift.Gamma(strength))
		case Rotate:
			rotateAngle, err := parseRotateAngle(values)
			if err != nil {
				return nil, fmt.Errorf("invalid rotate value: %w", err)
			}
			filters = append(filters, gift.Rotate(rotateAngle, image.Transparent, gift.LinearInterpolation))
		case Saturation:
			strengthPct := parseFloatValue(values, -100, 500, 0)
			filters = append(filters, gift.Saturation(strengthPct))
		case Sharpen:
			strength := parseFloatValue(values, 0.5, 1.5, 0.5)
			filters = append(filters, gift.UnsharpMask(1.0, strength, 0.0))
		case BorderRadius:
			radiusFilter, err := createBorderRadiusFilter(values)
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
