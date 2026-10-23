package http

import (
	"math"
	"stockflow/internal/model"
	"stockflow/internal/usecase"

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
	request := &model.SearchProductRequest{
		SKU:  ctx.Query("sku", ""),
		Name: ctx.Query("name", ""),
		Page: ctx.QueryInt("page", 1),
		Size: ctx.QueryInt("size", 10),
	}

	responses, total, err := p.ProductUseCase.SearchProduct(ctx.Context(), request)
	if err != nil {
		return err
	}

	paging := &model.PageMetadata{
		Page:      request.Page,
		Size:      request.Size,
		TotalItem: total,
		TotalPage: int64(math.Ceil(float64(total) / float64(request.Size))),
	}

	return ctx.JSON(model.WebResponse[[]model.ProductResponse]{
		Data:   responses,
		Paging: paging,
	})

}
