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

func BindRouteTemplate(server *fiber.App, k *kritiimages.KritiImages) {
	server.Get(`/cgi/images/tmpl\::options?/:template`, func(c *fiber.Ctx) error {
		optionsStr := c.Params("options", "")
		templateName := c.Params("template", "")
		log.Infow("new request", "options", optionsStr, "template", templateName)

		if optionsStr == "" {
			return c.Status(http.StatusBadRequest).SendString("options parameter is required")
		}

		if templateName == "" {
			return c.Status(http.StatusBadRequest).SendString("template name parameter is required")
		}

		vars, dest, err := getVars(optionsStr)
		if err != nil {
			log.Errorw("failed to render template", "options", optionsStr, "template", templateName, "error", err.Error())
			return c.Status(http.StatusInternalServerError).SendString(fmt.Sprintf("failed to process the request; %s", err.Error()))
		}

		// TODO: handle errors
		buffer, err := k.RenderTemplate(c.Context(), templateName+".json", vars)
		if err != nil {
			fmt.Println(err.Error())
		}
		if errors.Is(err, kritiimages.ErrSourceImageNotFound) {
			return c.Status(http.StatusNotFound).SendString("image not found")
		} else if errors.Is(err, kritiimages.ErrTransformationsNotFound) {
			return c.Status(http.StatusBadRequest).SendString("invalid template requested")
		} else if errors.Is(err, kritiimages.ErrInvalidImageFormat) {
			return c.Status(http.StatusBadRequest).SendString("invalid image format requested")
		} else if err != nil {
			return c.Status(http.StatusInternalServerError).SendString("failed render template")
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

// getVars converts url path portion containing variables
// into map of <string, string> required for template substitution.
// e.g. text=Lorem%20Ipsum,image_path=image.jpg is converted to {text=Lorem Ipsum, image_path=image.jpg}
func getVars(optionsStr string) (map[string]string, *kritiimages.DestinationImage, error) {
	options := strings.Split(optionsStr, ",")

	// set default values
	destination := kritiimages.DestinationImage{
		BgColor: color.Transparent,
		Quality: 100,
		Format:  "png", // always for templates
	}

	vars := make(map[string]string)
	for _, optStr := range options {
		parts := strings.Split(optStr, "=")
		if len(parts) != 2 {
			return vars, nil, fmt.Errorf("invalid option format: %s", optStr)
		}

		key := strings.TrimSpace(parts[0])
		value, err := url.PathUnescape(strings.TrimSpace(parts[1]))
		if err != nil {
			log.Warn("failed to unescape template variable, using original value", "variable", key)
			value = parts[1]
		}

		switch key {
		case "format":
			destination.Format, err = utils.ParseFormatValue(value)
			if err != nil {
				return vars, nil, fmt.Errorf("invalid format: %w", err)
			}
		case "quality":
			destination.Quality, err = utils.ParseIntValue(value, 1, 100)
			if err != nil {
				return vars, nil, fmt.Errorf("invalid quality: %w", err)
			}
		default:
			vars[key] = value
		}
	}

	return vars, &destination, nil
}
