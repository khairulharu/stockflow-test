package test

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"

	"stockflow/internal/model"
	"stockflow/internal/repository"
	"stockflow/internal/usecase"
)

// UseCase masih bergantung ke *repository.ProductRepository (struct, bukan interface),
// jadi yang di-mock adalah *sql.DB-nya lewat sqlmock.
func newUseCase(t *testing.T) (*usecase.ProductUseCase, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	repo := repository.NewProductRepository(db)
	return usecase.NewProductUseCase(repo), mock
}

func TestProductUseCase_SearchProduct_Success(t *testing.T) {
	uc, mock := newUseCase(t)

	mock.ExpectQuery(countAllQuery).
		WillReturnRows(countRows(2))
	mock.ExpectQuery(selectBase+" "+orderBy+" LIMIT $1 OFFSET $2").
		WithArgs(10, 0).
		WillReturnRows(productRows().
			AddRow(int64(2), "SKU-2", "Teh Botol", int64(5000), int64(0), false, createdAt).
			AddRow(int64(1), "SKU-1", "Kopi Susu", int64(15000), int64(10), true, createdAt))

	responses, total, err := uc.SearchProduct(context.Background(), &model.SearchProductRequest{Page: 1, Size: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}

	want := []model.ProductResponse{
		{ID: 2, SKU: "SKU-2", Name: "Teh Botol", Price: 5000, Stock: 0, IsActive: false, CreatedAt: createdAt},
		{ID: 1, SKU: "SKU-1", Name: "Kopi Susu", Price: 15000, Stock: 10, IsActive: true, CreatedAt: createdAt},
	}
	if len(responses) != len(want) {
		t.Fatalf("len(responses) = %d, want %d", len(responses), len(want))
	}
	for i := range want {
		got := responses[i]
		if got.ID != want[i].ID || got.SKU != want[i].SKU || got.Name != want[i].Name ||
			got.Price != want[i].Price || got.Stock != want[i].Stock ||
			got.IsActive != want[i].IsActive || !got.CreatedAt.Equal(want[i].CreatedAt) {
			t.Errorf("responses[%d] = %+v, want %+v", i, got, want[i])
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestProductUseCase_SearchProduct_PassesFilterAndPaging(t *testing.T) {
	uc, mock := newUseCase(t)

	// page 2, size 5 -> OFFSET 5; filter sku + name -> LIMIT $3 OFFSET $4
	mock.ExpectQuery(countAllQuery+" WHERE sku ILIKE $1 AND name ILIKE $2").
		WithArgs("%A1%", "%kopi%").
		WillReturnRows(countRows(7))
	mock.ExpectQuery(selectBase+" WHERE sku ILIKE $1 AND name ILIKE $2 "+orderBy+" LIMIT $3 OFFSET $4").
		WithArgs("%A1%", "%kopi%", 5, 5).
		WillReturnRows(productRows().
			AddRow(int64(6), "A1-6", "Kopi Hitam", int64(12000), int64(4), true, createdAt).
			AddRow(int64(7), "A1-7", "Kopi Tubruk", int64(13000), int64(2), true, createdAt))

	responses, total, err := uc.SearchProduct(context.Background(), &model.SearchProductRequest{
		SKU: "A1", Name: "kopi", Page: 2, Size: 5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 7 {
		t.Errorf("total = %d, want 7", total)
	}
	if len(responses) != 2 {
		t.Errorf("len(responses) = %d, want 2", len(responses))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestProductUseCase_SearchProduct_EmptyResult(t *testing.T) {
	uc, mock := newUseCase(t)

	mock.ExpectQuery(countAllQuery + " WHERE name ILIKE $1").
		WithArgs("%nothing%").
		WillReturnRows(countRows(0))

	responses, total, err := uc.SearchProduct(context.Background(), &model.SearchProductRequest{
		Name: "nothing", Page: 1, Size: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
	if responses == nil {
		t.Error("responses is nil, want empty slice (supaya JSON jadi [] bukan null)")
	}
	if len(responses) != 0 {
		t.Errorf("len(responses) = %d, want 0", len(responses))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestProductUseCase_SearchProduct_RepositoryError(t *testing.T) {
	tests := []struct {
		name  string
		setup func(mock sqlmock.Sqlmock)
	}{
		{
			name: "count query fails",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(countAllQuery).
					WillReturnError(errors.New("db down"))
			},
		},
		{
			name: "select query fails",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(countAllQuery).
					WillReturnRows(countRows(1))
				mock.ExpectQuery(selectBase+" "+orderBy+" LIMIT $1 OFFSET $2").
					WithArgs(10, 0).
					WillReturnError(errors.New("db down"))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc, mock := newUseCase(t)
			tc.setup(mock)

			responses, total, err := uc.SearchProduct(context.Background(), &model.SearchProductRequest{Page: 1, Size: 10})

			// error internal DB tidak boleh bocor ke client, dibungkus jadi 500
			if !errors.Is(err, fiber.ErrInternalServerError) {
				t.Fatalf("err = %v, want fiber.ErrInternalServerError", err)
			}
			if responses != nil || total != 0 {
				t.Errorf("got responses=%v total=%d, want nil and 0", responses, total)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet expectations: %v", err)
			}
		})
	}
}
