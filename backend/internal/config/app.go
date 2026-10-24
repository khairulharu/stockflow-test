package config

import (
	"database/sql"
	"stockflow/internal/delivery/http"
	"stockflow/internal/delivery/http/route"
	"stockflow/internal/repository"
	"stockflow/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type BootstrapConfig struct {
	DB  *sql.DB
	App *fiber.App
}

func Bootstrap(config *BootstrapConfig) {
	productRepository := repository.NewProductRepository(config.DB)

	productUseCase := usecase.NewProductUseCase(productRepository)

	productController := http.NewProductController(productUseCase)

	routeConfig := route.RouteConfig{
		App:               config.App,
		ProductController: productController,
	}
	routeConfig.Setup()
}
