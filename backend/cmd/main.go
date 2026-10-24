package main

import (
	"fmt"
	"log"
	"stockflow/internal/config"
)

func main() {
	viperConfig := config.NewViper()
	db := config.NewDatabase(viperConfig)
	app := config.NewFiber(viperConfig)
	validate := config.NewValidator()

	config.Bootstrap(&config.BootstrapConfig{
		DB:       db,
		App:      app,
		Validate: validate,
	})

	webPort := viperConfig.GetInt("web.port")
	err := app.Listen(fmt.Sprintf(":%d", webPort))
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
