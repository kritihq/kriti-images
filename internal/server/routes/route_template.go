package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/kritihq/kriti-images/internal/template"
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

		// TODO: handle errors
		templateJSON, err := fetchTemplate(templateName)
		if err != nil {
			return c.Status(http.StatusInternalServerError).SendString("failed to fetch template")
		}
		var vars map[string]string

		// TODO: handle errors
		buffer, err := template.RenderTemplate(templateJSON, vars)
		if err != nil {
			fmt.Println(err.Error())
		}
		if errors.Is(err, kritiimages.ErrSourceImageNotFound) {
			return c.Status(http.StatusNotFound).SendString("image not found")
		} else if errors.Is(err, kritiimages.ErrTransformationsNotFound) {
			return c.Status(http.StatusBadRequest).SendString("invalid transformation requested")
		} else if errors.Is(err, kritiimages.ErrInvalidImageFormat) {
			return c.Status(http.StatusBadRequest).SendString("invalid image format requested")
		} else if err != nil {
			return c.Status(http.StatusInternalServerError).SendString("failed to transform image")
		}

		// TODO: take from URL
		// format := dest.Format
		format := "png"
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

func fetchTemplate(fileName string) (string, error) {
	templatePath := "templates/json/" + fileName + ".json"

	data, err := os.ReadFile(templatePath)
	if err != nil {
		return "", errors.New("template file not found")
	}

	var jsonCheck any
	if err := json.Unmarshal(data, &jsonCheck); err != nil {
		return "", errors.New("invalid JSON in template file")
	}

	return string(data), nil
}
