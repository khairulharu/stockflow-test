package route

import (
	"stockflow/internal/delivery/http"

	"github.com/gofiber/fiber/v2"
)

type RouteConfig struct {
	App               *fiber.App
	ProductController *http.ProductController
}

func (r *RouteConfig) Setup() {
	r.ProductRoute()
}

func (r *RouteConfig) ProductRoute() {
	r.App.Get("/api/products", r.ProductController.SearchProducts)
	r.App.Post("/api/products", r.ProductController.Create)
}
