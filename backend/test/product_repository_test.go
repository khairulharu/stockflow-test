package test

import (
	"context"
	"errors"
	"testing"

	"stockflow/internal/model"
	"stockflow/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
)

var errBoom = errors.New("boom")

func newMockRepo(t *testing.T) (*repository.ProductRepository, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	return repository.NewProductRepository(db), mock
}

func TestProductRepository_Search_ReturnsRows(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(countQuery).
		WithArgs("").
		WillReturnRows(countRows(2))
	mock.ExpectQuery(searchQuery).
		WithArgs("", 10, 0).
		WillReturnRows(productRows().
			AddRow(int64(2), "SKU-2", "Teh Botol", int64(5000), int64(0), false, createdAt, nil).
			AddRow(int64(1), "SKU-1", "Kopi Susu", int64(15000), int64(10), true, createdAt, lastMovement))

	products, total, err := repo.Search(context.Background(), &model.SearchProductRequest{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(products) != 2 {
		t.Fatalf("len(products) = %d, want 2", len(products))
	}

	first, second := products[0], products[1]
	if first.ID != 2 || first.SKU != "SKU-2" || first.Name != "Teh Botol" ||
		first.Price != 5000 || first.Stock != 0 || first.IsActive {
		t.Errorf("first product mismatch: %+v", first)
	}
	if first.LastMovementAt.Valid {
		t.Errorf("first.LastMovementAt should be NULL, got %v", first.LastMovementAt.Time)
	}

	if second.ID != 1 || second.SKU != "SKU-1" || second.Name != "Kopi Susu" ||
		second.Price != 15000 || second.Stock != 10 || !second.IsActive {
		t.Errorf("second product mismatch: %+v", second)
	}
	if !second.CreatedAt.Equal(createdAt) {
		t.Errorf("CreatedAt = %v, want %v", second.CreatedAt, createdAt)
	}
	if !second.LastMovementAt.Valid || !second.LastMovementAt.Time.Equal(lastMovement) {
		t.Errorf("LastMovementAt = %+v, want %v", second.LastMovementAt, lastMovement)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestProductRepository_Search_SearchAndPagination(t *testing.T) {
	tests := []struct {
		name       string
		request    model.SearchProductRequest
		wantSearch string
		wantLimit  int
		wantOffset int
	}{
		{
			name:       "empty search",
			request:    model.SearchProductRequest{Page: 1, Limit: 10},
			wantSearch: "",
			wantLimit:  10,
			wantOffset: 0,
		},
		{
			name:       "with search",
			request:    model.SearchProductRequest{Search: "kopi", Page: 1, Limit: 10},
			wantSearch: "kopi",
			wantLimit:  10,
			wantOffset: 0,
		},
		{
			name:       "page 3 limit 5 gives offset 10",
			request:    model.SearchProductRequest{Page: 3, Limit: 5},
			wantSearch: "",
			wantLimit:  5,
			wantOffset: 10,
		},
		{
			// escape dilakukan di usecase, repository tidak boleh mengubah input
			name:       "search is passed as-is",
			request:    model.SearchProductRequest{Search: `50\%\_off`, Page: 1, Limit: 10},
			wantSearch: `50\%\_off`,
			wantLimit:  10,
			wantOffset: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock := newMockRepo(t)

			mock.ExpectQuery(countQuery).
				WithArgs(tc.wantSearch).
				WillReturnRows(countRows(25))
			mock.ExpectQuery(searchQuery).
				WithArgs(tc.wantSearch, tc.wantLimit, tc.wantOffset).
				WillReturnRows(productRows().
					AddRow(int64(1), "A1", "Kopi", int64(15000), int64(10), true, createdAt, nil))

			request := tc.request
			products, total, err := repo.Search(context.Background(), &request)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if total != 25 {
				t.Errorf("total = %d, want 25", total)
			}
			if len(products) != 1 {
				t.Errorf("len(products) = %d, want 1", len(products))
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet expectations: %v", err)
			}
		})
	}
}

func TestProductRepository_Search_NoResult_SkipsDataQuery(t *testing.T) {
	repo, mock := newMockRepo(t)

	// hanya COUNT yang diharapkan; kalau repo tetap menjalankan SELECT data,
	// sqlmock mengembalikan error "unexpected query" dan test gagal.
	mock.ExpectQuery(countQuery).
		WithArgs("nothing").
		WillReturnRows(countRows(0))

	products, total, err := repo.Search(context.Background(), &model.SearchProductRequest{
		Search: "nothing", Page: 1, Limit: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 {
		t.Errorf("total = %d, want 0", total)
	}
	if products == nil {
		t.Error("products is nil, want empty slice (supaya JSON jadi [] bukan null)")
	}
	if len(products) != 0 {
		t.Errorf("len(products) = %d, want 0", len(products))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestProductRepository_Search_CountError(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(countQuery).WithArgs("").WillReturnError(errBoom)

	products, total, err := repo.Search(context.Background(), &model.SearchProductRequest{Page: 1, Limit: 10})
	if !errors.Is(err, errBoom) {
		t.Fatalf("err = %v, want %v", err, errBoom)
	}
	if products != nil || total != 0 {
		t.Errorf("got products=%v total=%d, want nil and 0", products, total)
	}
}

func TestProductRepository_Search_SelectError(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(countQuery).WithArgs("").WillReturnRows(countRows(1))
	mock.ExpectQuery(searchQuery).
		WithArgs("", 10, 0).
		WillReturnError(errBoom)

	products, total, err := repo.Search(context.Background(), &model.SearchProductRequest{Page: 1, Limit: 10})
	if !errors.Is(err, errBoom) {
		t.Fatalf("err = %v, want %v", err, errBoom)
	}
	if products != nil || total != 0 {
		t.Errorf("got products=%v total=%d, want nil and 0", products, total)
	}
}

func TestProductRepository_Search_ScanError(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(countQuery).WithArgs("").WillReturnRows(countRows(1))
	mock.ExpectQuery(searchQuery).
		WithArgs("", 10, 0).
		// id berisi string non-angka, jadi Scan ke int64 gagal
		WillReturnRows(productRows().
			AddRow("not-a-number", "A1", "Kopi", int64(15000), int64(10), true, createdAt, nil))

	products, total, err := repo.Search(context.Background(), &model.SearchProductRequest{Page: 1, Limit: 10})
	if err == nil {
		t.Fatal("expected scan error, got nil")
	}
	if products != nil || total != 0 {
		t.Errorf("got products=%v total=%d, want nil and 0", products, total)
	}
}

func TestProductRepository_Search_RowsIterationError(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(countQuery).WithArgs("").WillReturnRows(countRows(2))
	mock.ExpectQuery(searchQuery).
		WithArgs("", 10, 0).
		WillReturnRows(productRows().
			AddRow(int64(1), "A1", "Kopi", int64(15000), int64(10), true, createdAt, nil).
			AddRow(int64(2), "A2", "Teh", int64(5000), int64(3), true, createdAt, nil).
			RowError(1, errBoom)) // error muncul saat iterasi row kedua

	products, total, err := repo.Search(context.Background(), &model.SearchProductRequest{Page: 1, Limit: 10})
	if !errors.Is(err, errBoom) {
		t.Fatalf("err = %v, want %v", err, errBoom)
	}
	if products != nil || total != 0 {
		t.Errorf("got products=%v total=%d, want nil and 0", products, total)
	}
}
