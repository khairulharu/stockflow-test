package test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
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

	mock.ExpectQuery(countQuery).
		WithArgs("").
		WillReturnRows(countRows(2))
	mock.ExpectQuery(searchQuery).
		WithArgs("", 10, 0).
		WillReturnRows(productRows().
			AddRow(int64(2), "SKU-2", "Teh Botol", int64(5000), int64(0), false, createdAt, nil).
			AddRow(int64(1), "SKU-1", "Kopi Susu", int64(15000), int64(10), true, createdAt, lastMovement))

	responses, total, err := uc.SearchProduct(context.Background(), &model.SearchProductRequest{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(responses) != 2 {
		t.Fatalf("len(responses) = %d, want 2", len(responses))
	}

	first, second := responses[0], responses[1]
	if first.ID != 2 || first.SKU != "SKU-2" || first.Name != "Teh Botol" ||
		first.Price != 5000 || first.Stock != 0 || first.IsActive ||
		!first.CreatedAt.Equal(createdAt) {
		t.Errorf("first response mismatch: %+v", first)
	}
	if first.LastMovementAt != nil {
		t.Errorf("first.LastMovement = %v, want nil (produk belum punya movement)", first.LastMovementAt)
	}

	if second.ID != 1 || second.SKU != "SKU-1" || second.Name != "Kopi Susu" ||
		second.Price != 15000 || second.Stock != 10 || !second.IsActive ||
		!second.CreatedAt.Equal(createdAt) {
		t.Errorf("second response mismatch: %+v", second)
	}
	if second.LastMovementAt == nil || !second.LastMovementAt.Equal(lastMovement) {
		t.Errorf("second.LastMovement = %v, want %v", second.LastMovementAt, lastMovement)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestProductUseCase_SearchProduct_PassesSearchAndPaging(t *testing.T) {
	uc, mock := newUseCase(t)

	// page 2, limit 5 -> OFFSET 5
	mock.ExpectQuery(countQuery).
		WithArgs("kopi").
		WillReturnRows(countRows(7))
	mock.ExpectQuery(searchQuery).
		WithArgs("kopi", 5, 5).
		WillReturnRows(productRows().
			AddRow(int64(6), "KOP-6", "Kopi Hitam", int64(12000), int64(4), true, createdAt, nil).
			AddRow(int64(7), "KOP-7", "Kopi Tubruk", int64(13000), int64(2), true, createdAt, nil))

	responses, total, err := uc.SearchProduct(context.Background(), &model.SearchProductRequest{
		Search: "kopi", Page: 2, Limit: 5,
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

func TestProductUseCase_SearchProduct_EscapesLikeWildcards(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string // nilai yang harus sampai ke repository / query
	}{
		{name: "percent and underscore", input: `50%_off`, want: `50\%\_off`},
		{name: "only underscore", input: `_`, want: `\_`},
		{name: "only percent", input: `%`, want: `\%`},
		{name: "backslash", input: `a\b`, want: `a\\b`},
		{name: "plain text unchanged", input: `kopi`, want: `kopi`},
		{name: "empty stays empty", input: ``, want: ``},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc, mock := newUseCase(t)

			mock.ExpectQuery(countQuery).
				WithArgs(tc.want).
				WillReturnRows(countRows(1))
			mock.ExpectQuery(searchQuery).
				WithArgs(tc.want, 10, 0).
				WillReturnRows(productRows().
					AddRow(int64(1), "A1", "Kopi", int64(15000), int64(10), true, createdAt, nil))

			_, _, err := uc.SearchProduct(context.Background(), &model.SearchProductRequest{
				Search: tc.input, Page: 1, Limit: 10,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// kalau argumen yang dikirim ke DB beda dari tc.want, sqlmock gagal di sini
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet expectations: %v", err)
			}
		})
	}
}

func TestProductUseCase_SearchProduct_EmptyResult(t *testing.T) {
	uc, mock := newUseCase(t)

	mock.ExpectQuery(countQuery).
		WithArgs("nothing").
		WillReturnRows(countRows(0))

	responses, total, err := uc.SearchProduct(context.Background(), &model.SearchProductRequest{
		Search: "nothing", Page: 1, Limit: 10,
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
				mock.ExpectQuery(countQuery).
					WithArgs("").
					WillReturnError(errors.New("db down"))
			},
		},
		{
			name: "select query fails",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(countQuery).
					WithArgs("").
					WillReturnRows(countRows(1))
				mock.ExpectQuery(searchQuery).
					WithArgs("", 10, 0).
					WillReturnError(errors.New("db down"))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc, mock := newUseCase(t)
			tc.setup(mock)

			responses, total, err := uc.SearchProduct(context.Background(), &model.SearchProductRequest{Page: 1, Limit: 10})

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

func TestProductResponse_LastMovementAt_JSON(t *testing.T) {
	b, err := json.Marshal(model.ProductResponse{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"last_movement_at":null`) {
		t.Errorf("json = %s, want key last_movement_at with null", b)
	}
}
