package routes

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"net/http"
	"strings"
	"time"

	"github.com/chai2010/webp"
	"github.com/disintegration/gift"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/kritihq/kriti-images/internal/imagesources"
	"github.com/kritihq/kriti-images/internal/transformations"
)

func BindRoutesBase(server *fiber.App, imageSource imagesources.ImageSource) {
	server.Get(`/cgi/images/tr\::options?/:image`, func(c *fiber.Ctx) error {
		optionsStr := c.Params("options", "")
		imagePath := c.Params("image", "")
		log.Infow("new request", "options", optionsStr, "path", imagePath)

		if optionsStr == "" {
			return c.Status(http.StatusBadRequest).SendString("Options parameter is required")
		}

		if imagePath == "" {
			return c.Status(http.StatusBadRequest).SendString("Image parameter is required")
		}

		src, srcFormat, err := imageSource.GetImage(imagePath)
		if err != nil {
			return c.Status(http.StatusNotFound).SendString("source image not found")
		}

		// Parse transformation context
		tctx, err := transformations.GetContextFromString(optionsStr, src, srcFormat)
		if err != nil {
			log.Errorw("failed to transform image", "options", optionsStr, "path", imagePath, "error", err.Error())
			return c.Status(http.StatusInternalServerError).SendString("failed to process the request")
		}

		g := gift.New(tctx.Filters...)

		// Create destination image
		dstBounds := g.Bounds(src.Bounds())
		dst := image.NewRGBA(dstBounds)

		// Apply background color if needed
		if tctx.DestinationImage.BgColor != color.Transparent {
			bounds := dst.Bounds()
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					dst.Set(x, y, tctx.DestinationImage.BgColor)
				}
			}
		}

		// Apply transformations
		g.Draw(dst, src)

		// Encode output using format from transformation context
		buffer := new(bytes.Buffer)
		format := tctx.DestinationImage.Format
		quality := tctx.DestinationImage.Quality

		switch strings.ToLower(format) {
		case "jpg", "jpeg":
			if err := jpeg.Encode(buffer, dst, &jpeg.Options{Quality: quality}); err != nil {
				return c.Status(http.StatusInternalServerError).SendString("failed to encode image")
			}
			c.Set("Content-Type", "image/jpeg")
		case "png":
			if err := png.Encode(buffer, dst); err != nil {
				return c.Status(http.StatusInternalServerError).SendString("failed to encode image")
			}
			c.Set("Content-Type", "image/png")
		case "webp":
			if err := webp.Encode(buffer, dst, &webp.Options{Quality: float32(quality)}); err != nil {
				return c.Status(http.StatusInternalServerError).SendString("failed to encode image")
			}
			c.Set("Content-Type", "image/webp")
		default:
			return c.Status(http.StatusBadRequest).SendString("unsupported format")
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

	server.Get("/demo", func(c *fiber.Ctx) error {
		sampleImage := "image1.jpg"
		return c.Render("demo", fiber.Map{
			"SampleImage": sampleImage,
		})
	})

	server.Static("/static", "web/static", fiber.Static{
		MaxAge: 86400, // 24 hours
	})

	server.Use(func(c *fiber.Ctx) error {
		return c.Status(http.StatusNotFound).Render("404", 0)
	})

}
