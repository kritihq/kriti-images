package routes

import (
	"errors"
	"fmt"
	"image/color"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/kritihq/kriti-images/internal/utils"
	"github.com/kritihq/kriti-images/pkg/kritiimages"
)

func BindRouteTransformation(server *fiber.App, k *kritiimages.KritiImages) {
	server.Get(`/cgi/images/tr\::options?/:image`, func(c *fiber.Ctx) error {
		optionsStr := c.Params("options", "")
		imagePath, err := url.PathUnescape(c.Params("image", ""))
		if err != nil {
			log.Warn("failed to unescape image path, using original value", "path", imagePath)
			imagePath = c.Params("image", "")
		}
		log.Infow("new request", "options", optionsStr, "path", imagePath)

		if optionsStr == "" {
			return c.Status(http.StatusBadRequest).SendString("options parameter is required")
		}

		if imagePath == "" {
			return c.Status(http.StatusBadRequest).SendString("image parameter is required")
		}

		// Parse transformation context
		options, dest, err := getContextFromString(optionsStr)
		if err != nil {
			log.Errorw("failed to transform image", "options", optionsStr, "path", imagePath, "error", err.Error())
			return c.Status(http.StatusInternalServerError).SendString(fmt.Sprintf("failed to process the request; %s", err.Error()))
		}

		buffer, err := k.Transform(c.Context(), imagePath, dest, options)
		if errors.Is(err, kritiimages.ErrSourceImageNotFound) {
			return c.Status(http.StatusNotFound).SendString("image not found")
		} else if errors.Is(err, kritiimages.ErrTransformationsNotFound) {
			return c.Status(http.StatusBadRequest).SendString("invalid transformation requested")
		} else if errors.Is(err, kritiimages.ErrInvalidImageFormat) {
			return c.Status(http.StatusBadRequest).SendString("invalid image format requested")
		} else if err != nil {
			return c.Status(http.StatusInternalServerError).SendString("failed to transform image")
		}

		format := dest.Format
		switch strings.ToLower(format) {
		case "jpg", "jpeg":
			c.Set("Content-Type", "image/jpeg")
		case "png":
			c.Set("Content-Type", "image/png")
		case "webp":
			c.Set("Content-Type", "image/webp")
		default:
			return c.Status(http.StatusBadRequest).SendString("invalid image format requested")
		}

		// Set CDN-friendly caching headers
		c.Set("Cache-Control", "public, max-age=31536000, immutable") // 1 year cache
		c.Set("Expires", time.Now().Add(time.Hour*24*365).UTC().Format(http.TimeFormat))
		c.Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))

		// Add Vary header to ensure CDN caches different versions properly
		c.Set("Vary", "Accept")

		// Security headers for CDN
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("Content-Security-Policy", "default-src 'none'")

		// Add CDN-specific headers
		c.Set("X-Robots-Tag", "noindex, nofollow")
		c.Set("Access-Control-Allow-Origin", "*")

		return c.Status(http.StatusOK).Send(buffer.Bytes())
	})
}

// getContextFromString converts url path portion containing transformations
// into map of TransformationOption and provided values for that transformation.
// e.g. blur=90,width=100 is converted to {blur: 90, width: 100}.
//
// return meaningful errors, they are sent as response as is
func getContextFromString(optionsStr string) (map[kritiimages.TransformationOption]string, *kritiimages.DestinationImage, error) {
	options := strings.Split(optionsStr, ",")

	destination := kritiimages.DestinationImage{
		BgColor: color.Transparent,
		Quality: 100,
	}

	trValues := make(map[kritiimages.TransformationOption]string)
	for _, optStr := range options {
		transformation, values, err := processOption(optStr)
		if err != nil {
			return nil, nil, err
		}

		switch transformation {
		case kritiimages.Background:
			destination.BgColor, err = utils.ParseBackgroundColor(values)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid background color: %w", err)
			}
		case kritiimages.Width:
			destination.Width, err = utils.ParseIntValue(values, 1, 10000)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid width: %w", err)
			}
		case kritiimages.Height:
			destination.Height, err = utils.ParseIntValue(values, 1, 10000)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid height: %w", err)
			}
		case kritiimages.Format:
			destination.Format, err = utils.ParseFormatValue(values)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid format: %w", err)
			}
		case kritiimages.Quality:
			destination.Quality, err = utils.ParseIntValue(values, 1, 100)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid quality: %w", err)
			}
		default:
			trValues[transformation] = values
		}
	}

	return trValues, &destination, nil
}

// processOption returns given string to TransformationOption enum and its value
//
// return meaningful errors, they are sent as response as is
func processOption(optStr string) (kritiimages.TransformationOption, string, error) {
	parts := strings.Split(optStr, "=")
	if len(parts) != 2 {
		return -1, "", fmt.Errorf("invalid option format: %s", optStr)
	}

	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	switch key {
	case "flip":
		return kritiimages.Flip, value, nil
	case "blur":
		return kritiimages.Blur, value, nil
	case "brightness":
		return kritiimages.Brightness, value, nil
	case "contrast":
		return kritiimages.Contrast, value, nil
	case "fit":
		return kritiimages.Fit, value, nil
	case "gamma":
		return kritiimages.Gamma, value, nil
	case "rotate":
		return kritiimages.Rotate, value, nil
	case "saturation":
		return kritiimages.Saturation, value, nil
	case "sharpen":
		return kritiimages.Sharpen, value, nil
	case "background":
		return kritiimages.Background, value, nil
	case "width":
		return kritiimages.Width, value, nil
	case "height":
		return kritiimages.Height, value, nil
	case "format":
		return kritiimages.Format, value, nil
	case "quality":
		return kritiimages.Quality, value, nil
	case "radius":
		return kritiimages.BorderRadius, value, nil
	default:
		return -1, "", fmt.Errorf("unknown option: %s", key)
	}
}
