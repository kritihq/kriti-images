package main

import (
	"context"
	"fmt"
	"net"

	"github.com/gofiber/fiber/v2/log"

	"github.com/kritihq/kriti-images/internal/config"
	"github.com/kritihq/kriti-images/internal/server"
)

func main() {
	configs, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	server, _ := server.ConfigureAndGet(context.Background(), configs)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", configs.Server.Port))
	if err != nil {
		log.Panicw("failed to start listener on port", "port", configs.Server.Port, "error", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port // user can use "0" to start on random port
	defer listener.Close()

	log.Infow("starting server", "port", port)
	if err := server.Listener(listener); err != nil {
		log.Errorw("failed to start server", "error", err.Error())
	}
}
