package main

import (
	"context"
	"fmt"
	"net"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
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
)

func main() {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

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

	imgSrcDefault, imgSrcHTTP, err := getImageSource(context.Background(), cfg)
	if err != nil {
		log.Panicf("failed to configure image source, err: %w", err)
	}

	log.Info("registering routes")
	routes.BindRoutesBase(server, imgSrcDefault, imgSrcHTTP)

	if cfg.Experimental.EnableUploadAPI {
		log.Info("registering api/v0/upload")
		routes.BindAPIUpload(server, imgSrcDefault)
	}

	// Register 404 handler last, after all other routes
	server.Use(func(c *fiber.Ctx) error {
		return c.Status(404).Render("404", 0)
	})

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		log.Panicw("failed to start listener on port", "port", cfg.Server.Port, "error", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port // user can use "0" to start on random port
	defer listener.Close()

	log.Infow("starting server", "port", port)
	if err := server.Listener(listener); err != nil {
		log.Errorw("failed to start server", "error", err.Error())
	}
}

func getImageSource(ctx context.Context, cfg *config.Config) (imagesources.ImageSource, imagesources.ImageSourceHTTP, error) {
	validations := imagesources.SourceImageValidations{
		MaxImageDimension:  cfg.Images.MaxImageDimension,
		MaxFileSizeInBytes: cfg.Images.MaxImageSizeInBytes,
	}

	imageSourceHTTP := imagesources.NewImageSourceURL(&validations)

	if cfg.Images.Source == "awss3" {
		defaultSrc, err := imagesources.NewImageSourceS3(ctx, cfg.Images.AwsS3.Bucket, &validations)
		return defaultSrc, *imageSourceHTTP, err
	} else {
		return imagesources.NewImageSourceLocal(cfg.Images.Local.BasePath, &validations), *imageSourceHTTP, nil
	}
}
