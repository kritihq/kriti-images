package routes

import (
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/kritihq/kriti-images/internal/imagesources"
	"github.com/kritihq/kriti-images/pkg/kritiimages"
)

func TestBindRouteTransformation_SubdirectoryPath(t *testing.T) {
	app := setupTransformTestApp(t)

	tests := []struct {
		name string
		path string
	}{
		{
			name: "plain subdirectory path",
			path: "/cgi/images/tr:quality=100/garden/flower.png",
		},
		{
			name: "encoded slash path",
			path: "/cgi/images/tr:quality=100/garden%2Fflower.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test() error = %v", err)
			}

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				t.Fatalf("unexpected status code: got %d, body=%s", resp.StatusCode, string(body))
			}

			if got := resp.Header.Get("Content-Type"); got != "image/png" {
				t.Fatalf("unexpected content type: got %q", got)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("failed to read response body: %v", err)
			}
			if len(body) == 0 {
				t.Fatalf("expected non-empty image response body")
			}
		})
	}
}

func setupTransformTestApp(t *testing.T) *fiber.App {
	t.Helper()

	basePath := t.TempDir()
	filePath := filepath.Join(basePath, "garden", "flower.png")
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	img.Set(1, 0, color.RGBA{G: 255, A: 255})
	img.Set(0, 1, color.RGBA{B: 255, A: 255})
	img.Set(1, 1, color.RGBA{R: 255, G: 255, A: 255})

	f, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}
	if err := png.Encode(f, img); err != nil {
		_ = f.Close()
		t.Fatalf("failed to encode test image: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("failed to close test image file: %v", err)
	}

	validations := &imagesources.SourceImageValidations{
		MaxImageDimension:  8192,
		MaxFileSizeInBytes: 50 * 1024 * 1024,
	}
	localSource := kritiimages.NewImageSourceLocal(basePath, validations)
	httpSource := kritiimages.NewImageSourceURL(validations)
	service := kritiimages.New(map[string]kritiimages.ImageSource{
		"local": localSource,
		"http":  httpSource,
	}, localSource)

	app := fiber.New()
	BindRouteTransformation(app, service)
	return app
}
