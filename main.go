package main

import (
	"fmt"
	"net"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/template/html/v2"

	"github.com/kritihq/kriti-images/internal/imagesources"
	"github.com/kritihq/kriti-images/internal/server/routes"
)

// TODO: convert to application configs
const (
	Port           = 8080
	ImagesBasePath = "web/static/assets"
)

func main() {
	server := fiber.New(fiber.Config{
		EnablePrintRoutes:     false, // toggle only for debugging
		GETOnly:               false,
		DisableStartupMessage: true,
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          30 * time.Second,
		Views:                 html.New("./web/templates", ".html"),
	})

	src := imagesources.ImageSourceLocal{BasePath: ImagesBasePath}
	server.Use(limiter.New(limiter.Config{
		Max:               100,
		Expiration:        1 * time.Minute,
		LimiterMiddleware: limiter.SlidingWindow{},
		Next: func(c *fiber.Ctx) bool {
			return c.IP() == "127.0.0.1" // for testing; skip rate limiter when localhost
		},
	}))
	server.Use(helmet.New())
	server.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,HEAD,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization,Cache-Control,If-None-Match",
	}))
	server.Use(etag.New())
	server.Get("/metrics", monitor.New())
	server.Use(healthcheck.New(healthcheck.Config{
		LivenessEndpoint:  "/health/live",
		ReadinessEndpoint: "/health/ready",
	}))
	server.Use(logger.New())

	log.Info("registering routes")
	routes.BindRoutesBase(server, &src)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", Port))
	if err != nil {
		panic(err)
	}

	_ = listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close() // TODO: handle error

	log.Infow("starting server", "port", Port)
	if err := server.Listen(fmt.Sprintf(":%d", Port)); err != nil {
		log.Errorw("failed to start server", "error", err.Error())
	}
}
