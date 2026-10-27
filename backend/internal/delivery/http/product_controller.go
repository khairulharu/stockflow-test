package http

import (
	"errors"
	"fmt"
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

func (p *ProductController) Search(ctx *fiber.Ctx) error {
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
		if fiberErr, ok := errors.AsType[*fiber.Error](err); ok {
			return ctx.Status(fiberErr.Code).JSON(model.ErrorResponse{
				Error: model.ErrorDetail{
					Code: fiberErr.Message,
				},
			})
		}

		return ctx.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: model.ErrorDetail{
				Code: "INTERNAL_SERVER_ERROR",
			},
		})
	}

	return ctx.JSON(model.PageResponse[model.ProductResponseWithMovement]{
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
		if fiberErr, ok := errors.AsType[*fiber.Error](err); ok {
			return ctx.Status(fiberErr.Code).JSON(model.ErrorResponse{
				Error: model.ErrorDetail{
					Code: fiberErr.Message,
				},
			})
		}

		return ctx.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: model.ErrorDetail{
				Code: "INTERNAL_SERVER_ERROR",
			},
		})
	}

	return ctx.SendStatus(fiber.StatusCreated)
}

func (p *ProductController) Update(ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id")
	if err != nil {
		return fiber.ErrBadRequest
	}
	var request model.UpdateProductRequest
	if err := ctx.BodyParser(&request); err != nil {
		return fiber.ErrBadRequest
	}
	request.ID = int64(id)

	response, err := p.ProductUseCase.UpdateProduct(ctx.UserContext(), &request)
	if err != nil {
		if fiberErr, ok := errors.AsType[*fiber.Error](err); ok {
			return ctx.Status(fiberErr.Code).JSON(model.ErrorResponse{
				Error: model.ErrorDetail{
					Code: fiberErr.Message,
				},
			})
		}

		return ctx.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: model.ErrorDetail{
				Code: "INTERNAL_SERVER_ERROR",
			},
		})
	}

	return ctx.JSON(response)
}

func (p *ProductController) Delete(ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id")
	if err != nil {
		return fiber.ErrBadRequest
	}
	fmt.Println(id)
	if err := p.ProductUseCase.DeleteProduct(ctx.UserContext(), id); err != nil {
		if fiberErr, ok := errors.AsType[*fiber.Error](err); ok {
			return ctx.Status(fiberErr.Code).JSON(model.ErrorResponse{
				Error: model.ErrorDetail{
					Code: fiberErr.Message,
				},
			})
		}

		return ctx.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Error: model.ErrorDetail{
				Code: "INTERNAL_SERVER_ERROR",
			},
		})
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
