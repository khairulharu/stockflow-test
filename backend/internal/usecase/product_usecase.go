package usecase

import (
	"context"
	"stockflow/internal/model"
	"stockflow/internal/model/converter"
	"stockflow/internal/repository"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type ProductUseCase struct {
	Validate          *validator.Validate
	ProductRepository *repository.ProductRepository
}

func NewProductUseCase(productRepository *repository.ProductRepository, validate *validator.Validate) *ProductUseCase {
	return &ProductUseCase{
		Validate:          validate,
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

func (p *ProductUseCase) CreateProduct(ctx context.Context, request *model.CreateProductRequest) error {

	request.SKU = strings.ToUpper(strings.TrimSpace(request.SKU))
	request.Name = strings.TrimSpace(request.Name)
	request.Note = strings.TrimSpace(request.Note)

	if err := p.Validate.Struct(request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	exist, err := p.ProductRepository.ExistsBySKU(ctx, request.SKU)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	if exist {
		return fiber.NewError(fiber.StatusConflict, "sku already exists")
	}

	if err := p.ProductRepository.Create(ctx, request); err != nil {
		return fiber.ErrInternalServerError
	}

	return nil
}
