package server

import (
	"context"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/kritihq/kriti-images/internal/config"
	"github.com/kritihq/kriti-images/internal/imagesources"
	"github.com/kritihq/kriti-images/internal/server/routes"
	"github.com/kritihq/kriti-images/pkg/kritiimages"
)

func ConfigureAndGet(ctx context.Context, cfg *config.Config) (*fiber.App, *kritiimages.KritiImages) {
	server := initFiberApp(cfg)

	sources := getImageSources(ctx, &cfg.Images)
	service := kritiimages.New(sources, sources[cfg.Images.Source])

	routes.BindRouteTransformation(server, service)

	// NOTE: do we need upload feature?
	// It will need auth layer to be prod ready
	if cfg.Experimental.EnableUploadAPI {
		routes.BindAPIUpload(server, service)
	}

	// Register 404 handler last, after all other routes
	server.Use(func(c *fiber.Ctx) error {
		return c.Status(404).Render("404", 0)
	})

	return server, service
}

func initFiberApp(cfg *config.Config) *fiber.App {
	server := fiber.New(fiber.Config{
		AppName:               "Kriti Images",
		EnablePrintRoutes:     cfg.Server.EnablePrintRoutes,
		DisableStartupMessage: true,
		ReadTimeout:           cfg.Server.ReadTimeout,
		WriteTimeout:          cfg.Server.WriteTimeout,
	})
	server.Use(limiter.New(limiter.Config{
		Max:               cfg.Server.Limiter.Max,
		Expiration:        cfg.Server.Limiter.Expiration,
		LimiterMiddleware: limiter.SlidingWindow{},
		Next: func(c *fiber.Ctx) bool {
			return c.IP() == "127.0.0.1" // for testing; skip rate limiter when localhost
		},
	}))
	server.Use(helmet.New(helmet.Config{
		CrossOriginResourcePolicy: cfg.Server.CrossOriginPolicy.Corp,
	}))
	server.Use(cors.New(cors.Config{
		AllowOrigins: cfg.Server.CrossOriginPolicy.CorsAllowOrigins,
		AllowMethods: cfg.Server.CrossOriginPolicy.CorsAllowMethods,
		AllowHeaders: cfg.Server.CrossOriginPolicy.CorsAllowHeaders,
	}))
	server.Use(etag.New())
	server.Get("/metrics", monitor.New())
	server.Use(healthcheck.New(healthcheck.Config{
		LivenessEndpoint:  "/health/live",
		ReadinessEndpoint: "/health/ready",
	}))
	server.Use(logger.New())

	return server
}

func getImageSources(ctx context.Context, cfg *config.ImagesConfig) map[string]kritiimages.ImageSource {
	validations := imagesources.SourceImageValidations{
		MaxImageDimension:  cfg.MaxImageDimension,
		MaxFileSizeInBytes: cfg.MaxImageSizeInBytes,
	}

	sources := make(map[string]kritiimages.ImageSource, 0)
	switch cfg.Source {
	case "awss3":
		// TODO: handle errors
		s3Client, _ := getS3Client(ctx)
		sources["awss3"], _ = kritiimages.NewImageSourceS3(ctx, cfg.AwsS3.Bucket, s3Client, &validations)
	case "local":
		sources["local"] = kritiimages.NewImageSourceLocal(cfg.Local.BasePath, &validations)
	}

	// always present, for now
	sources["http"] = kritiimages.NewImageSourceURL(&validations)
	return sources
}

func getS3Client(ctx context.Context) (*s3.Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return s3.NewFromConfig(cfg), nil
}
