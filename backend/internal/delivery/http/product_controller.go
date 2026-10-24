package http

import (
	"stockflow/internal/model"
	"stockflow/internal/usecase"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type ProductController struct {
	ProductUseCase *usecase.ProductUseCase
}

func NewProductController(productUseCase *usecase.ProductUseCase) *ProductController {
	return &ProductController{
		ProductUseCase: productUseCase,
	}
}

func (p *ProductController) SearchProducts(ctx *fiber.Ctx) error {
	page := ctx.QueryInt("page", 1)
	limit := ctx.QueryInt("limit", 10)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	request := &model.SearchProductRequest{
		Search: strings.TrimSpace(ctx.Query("search", "")),
		Page:   page,
		Limit:  limit,
	}

	responses, total, err := p.ProductUseCase.SearchProduct(ctx.UserContext(), request)
	if err != nil {
		return err
	}

	return ctx.JSON(model.PageResponse[model.ProductResponse]{
		Data:  responses,
		Page:  request.Page,
		Limit: request.Limit,
		Total: total,
	})

}

func (p *ProductController) Create(ctx *fiber.Ctx) error {
	var request model.CreateProductRequest

	if err := ctx.BodyParser(&request); err != nil {
		return fiber.ErrBadRequest
	}

	if err := p.ProductUseCase.CreateProduct(ctx.UserContext(), &request); err != nil {
		return err
	}

	return ctx.SendStatus(fiber.StatusCreated)
}
