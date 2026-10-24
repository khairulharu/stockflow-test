package usecase

import (
	"context"
	"stockflow/internal/model"
	"stockflow/internal/model/converter"
	"stockflow/internal/repository"

	"github.com/gofiber/fiber/v2"
)

type ProductUseCase struct {
	ProductRepository *repository.ProductRepository
}

func NewProductUseCase(productRepository *repository.ProductRepository) *ProductUseCase {
	return &ProductUseCase{
		ProductRepository: productRepository,
	}
}

func (p *ProductUseCase) SearchProduct(ctx context.Context, request *model.SearchProductRequest) ([]model.ProductResponse, int64, error) {
	request.Search = request.EscapedSearch()
	products, total, err := p.ProductRepository.Search(ctx, request)
	if err != nil {
		return nil, 0, fiber.ErrInternalServerError
	}

	responses := make([]model.ProductResponse, len(products))
	for i, product := range products {
		responses[i] = *converter.ProductToResponse(&product)
	}

	return responses, total, nil
}
