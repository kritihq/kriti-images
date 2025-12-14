package routes

import (
	"fmt"
	"image"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/kritihq/kriti-images/pkg/kritiimages"
)

func BindAPIUpload(server *fiber.App, k *kritiimages.KritiImages) {
	// TODO: uploads only happen on default sources, for now

	server.Post("/api/v0/images", func(c *fiber.Ctx) error {
		// Get the uploaded file
		file, err := c.FormFile("image")
		if err != nil {
			log.Errorw("failed to get uploaded file", "error", err.Error())
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "No image file provided",
			})
		}

		// Validate file type
		contentType := file.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "image/") {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "File must be an image",
			})
		}

		// Get filename from form or use original filename
		filename := c.FormValue("filename")
		if filename == "" {
			filename = file.Filename
		}

		// Validate filename extension
		ext := strings.ToLower(filepath.Ext(filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Unsupported file format. Only JPG, PNG, and WebP are allowed",
			})
		}

		// Open the uploaded file
		src, err := file.Open()
		if err != nil {
			log.Errorw("failed to open uploaded file", "error", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to process uploaded file",
			})
		}
		defer src.Close()

		// Decode the image
		img, format, err := image.Decode(src)
		if err != nil {
			log.Errorw("failed to decode image", "error", err.Error())
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid image file",
			})
		}

		// Upload the image using the image source
		if err := k.DefaultSource.UploadImage(c.Context(), filename, img); err != nil {
			log.Errorw("failed to upload image", "filename", filename, "error", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to save image: %s", err.Error()),
			})
		}

		log.Infow("image uploaded successfully", "filename", filename, "format", format, "size", fmt.Sprintf("%dx%d", img.Bounds().Dx(), img.Bounds().Dy()))

		return c.Status(http.StatusCreated).JSON(fiber.Map{
			"message":  "Image uploaded successfully",
			"filename": filename,
			"format":   format,
			"size": fiber.Map{
				"width":  img.Bounds().Dx(),
				"height": img.Bounds().Dy(),
			},
		})
	})

	server.Put("/api/v0/images", func(c *fiber.Ctx) error {
		// Get the uploaded file
		file, err := c.FormFile("image")
		if err != nil {
			log.Errorw("failed to get uploaded file", "error", err.Error())
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "No image file provided",
			})
		}

		// Validate file type
		contentType := file.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "image/") {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "File must be an image",
			})
		}

		// Get filename from form or use original filename
		filename := c.FormValue("filename")
		if filename == "" {
			filename = file.Filename
		}

		// Validate filename extension
		ext := strings.ToLower(filepath.Ext(filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Unsupported file format. Only JPG, PNG, and WebP are allowed",
			})
		}

		// Check if the image exists (for PUT, we might want to verify it exists)
		_, _, err = k.DefaultSource.GetImage(c.Context(), filename)
		if err != nil {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": "Image not found",
			})
		}

		// Open the uploaded file
		src, err := file.Open()
		if err != nil {
			log.Errorw("failed to open uploaded file", "error", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to process uploaded file",
			})
		}
		defer src.Close()

		// Decode the image
		img, format, err := image.Decode(src)
		if err != nil {
			log.Errorw("failed to decode image", "error", err.Error())
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid image file",
			})
		}

		// Upload the image using the image source (this will overwrite the existing file)
		if err := k.DefaultSource.UploadImage(c.Context(), filename, img); err != nil {
			log.Errorw("failed to update image", "filename", filename, "error", err.Error())
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to update image: %s", err.Error()),
			})
		}

		log.Infow("image updated successfully", "filename", filename, "format", format, "size", fmt.Sprintf("%dx%d", img.Bounds().Dx(), img.Bounds().Dy()))

		return c.Status(http.StatusOK).JSON(fiber.Map{
			"message":  "Image updated successfully",
			"filename": filename,
			"format":   format,
			"size": fiber.Map{
				"width":  img.Bounds().Dx(),
				"height": img.Bounds().Dy(),
			},
		})
	})
}
