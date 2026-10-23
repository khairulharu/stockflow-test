package test

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"stockflow/internal/model"
	"stockflow/internal/repository"
)

var errBoom = errors.New("boom")

// newMockRepo bikin repository yang konek ke sqlmock.
// QueryMatcherEqual: SQL harus sama persis (whitespace diabaikan),
// jadi nomor placeholder $1, $2, ... ikut ke-verifikasi.
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

	mock.ExpectQuery(countAllQuery).
		WillReturnRows(countRows(2))
	mock.ExpectQuery(selectBase+" "+orderBy+" LIMIT $1 OFFSET $2").
		WithArgs(10, 0).
		WillReturnRows(productRows().
			AddRow(int64(2), "SKU-2", "Teh Botol", int64(5000), int64(0), false, createdAt).
			AddRow(int64(1), "SKU-1", "Kopi Susu", int64(15000), int64(10), true, createdAt))

	products, total, err := repo.Search(context.Background(), &model.SearchProductRequest{Page: 1, Size: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	if len(products) != 2 {
		t.Fatalf("len(products) = %d, want 2", len(products))
	}

	// urutan harus sama dengan urutan row dari DB
	first, second := products[0], products[1]
	if first.ID != 2 || first.SKU != "SKU-2" || first.Name != "Teh Botol" ||
		first.Price != 5000 || first.Stock != 0 || first.IsActive {
		t.Errorf("first product mismatch: %+v", first)
	}
	if second.ID != 1 || second.SKU != "SKU-1" || second.Name != "Kopi Susu" ||
		second.Price != 15000 || second.Stock != 10 || !second.IsActive {
		t.Errorf("second product mismatch: %+v", second)
	}
	if !second.CreatedAt.Equal(createdAt) {
		t.Errorf("CreatedAt = %v, want %v", second.CreatedAt, createdAt)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestProductRepository_Search_FilterAndPagination(t *testing.T) {
	tests := []struct {
		name       string
		request    model.SearchProductRequest
		countQuery string
		dataQuery  string
		filterArgs []driver.Value // args yang dipakai filter (tanpa limit & offset)
		limit      int
		offset     int
	}{
		{
			name:       "no filter",
			request:    model.SearchProductRequest{Page: 1, Size: 10},
			countQuery: countAllQuery,
			dataQuery:  selectBase + " " + orderBy + " LIMIT $1 OFFSET $2",
			limit:      10,
			offset:     0,
		},
		{
			name:       "sku only",
			request:    model.SearchProductRequest{SKU: "A1", Page: 1, Size: 10},
			countQuery: countAllQuery + " WHERE sku ILIKE $1",
			dataQuery:  selectBase + " WHERE sku ILIKE $1 " + orderBy + " LIMIT $2 OFFSET $3",
			filterArgs: []driver.Value{"%A1%"},
			limit:      10,
			offset:     0,
		},
		{
			name:       "name only",
			request:    model.SearchProductRequest{Name: "kopi", Page: 1, Size: 10},
			countQuery: countAllQuery + " WHERE name ILIKE $1",
			dataQuery:  selectBase + " WHERE name ILIKE $1 " + orderBy + " LIMIT $2 OFFSET $3",
			filterArgs: []driver.Value{"%kopi%"},
			limit:      10,
			offset:     0,
		},
		{
			name:       "sku and name",
			request:    model.SearchProductRequest{SKU: "A1", Name: "kopi", Page: 1, Size: 10},
			countQuery: countAllQuery + " WHERE sku ILIKE $1 AND name ILIKE $2",
			dataQuery:  selectBase + " WHERE sku ILIKE $1 AND name ILIKE $2 " + orderBy + " LIMIT $3 OFFSET $4",
			filterArgs: []driver.Value{"%A1%", "%kopi%"},
			limit:      10,
			offset:     0,
		},
		{
			name:       "page 3 size 5 gives offset 10",
			request:    model.SearchProductRequest{Page: 3, Size: 5},
			countQuery: countAllQuery,
			dataQuery:  selectBase + " " + orderBy + " LIMIT $1 OFFSET $2",
			limit:      5,
			offset:     10,
		},
		{
			name:       "like wildcard from user input is escaped",
			request:    model.SearchProductRequest{Name: `50%_off`, Page: 1, Size: 10},
			countQuery: countAllQuery + " WHERE name ILIKE $1",
			dataQuery:  selectBase + " WHERE name ILIKE $1 " + orderBy + " LIMIT $2 OFFSET $3",
			filterArgs: []driver.Value{`%50\%\_off%`},
			limit:      10,
			offset:     0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock := newMockRepo(t)

			mock.ExpectQuery(tc.countQuery).
				WithArgs(tc.filterArgs...).
				WillReturnRows(countRows(25))

			dataArgs := append(append([]driver.Value{}, tc.filterArgs...), tc.limit, tc.offset)
			mock.ExpectQuery(tc.dataQuery).
				WithArgs(dataArgs...).
				WillReturnRows(productRows().
					AddRow(int64(1), "A1", "Kopi", int64(15000), int64(10), true, createdAt))

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

	// hanya COUNT yang diharapkan; kalau repo tetap jalanin SELECT data,
	// sqlmock akan return error "unexpected query" dan test gagal.
	mock.ExpectQuery(countAllQuery + " WHERE name ILIKE $1").
		WithArgs("%nothing%").
		WillReturnRows(countRows(0))

	products, total, err := repo.Search(context.Background(), &model.SearchProductRequest{
		Name: "nothing", Page: 1, Size: 10,
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

	mock.ExpectQuery(countAllQuery).WillReturnError(errBoom)

	products, total, err := repo.Search(context.Background(), &model.SearchProductRequest{Page: 1, Size: 10})
	if !errors.Is(err, errBoom) {
		t.Fatalf("err = %v, want %v", err, errBoom)
	}
	if products != nil || total != 0 {
		t.Errorf("got products=%v total=%d, want nil and 0", products, total)
	}
}

func TestProductRepository_Search_SelectError(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(countAllQuery).WillReturnRows(countRows(1))
	mock.ExpectQuery(selectBase+" "+orderBy+" LIMIT $1 OFFSET $2").
		WithArgs(10, 0).
		WillReturnError(errBoom)

	products, total, err := repo.Search(context.Background(), &model.SearchProductRequest{Page: 1, Size: 10})
	if !errors.Is(err, errBoom) {
		t.Fatalf("err = %v, want %v", err, errBoom)
	}
	if products != nil || total != 0 {
		t.Errorf("got products=%v total=%d, want nil and 0", products, total)
	}
}

func TestProductRepository_Search_ScanError(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(countAllQuery).WillReturnRows(countRows(1))
	mock.ExpectQuery(selectBase+" "+orderBy+" LIMIT $1 OFFSET $2").
		WithArgs(10, 0).
		// id berisi string non-angka, jadi Scan ke int64 gagal
		WillReturnRows(productRows().
			AddRow("not-a-number", "A1", "Kopi", int64(15000), int64(10), true, createdAt))

	products, total, err := repo.Search(context.Background(), &model.SearchProductRequest{Page: 1, Size: 10})
	if err == nil {
		t.Fatal("expected scan error, got nil")
	}
	if products != nil || total != 0 {
		t.Errorf("got products=%v total=%d, want nil and 0", products, total)
	}
}

func TestProductRepository_Search_RowsIterationError(t *testing.T) {
	repo, mock := newMockRepo(t)

	mock.ExpectQuery(countAllQuery).WillReturnRows(countRows(2))
	mock.ExpectQuery(selectBase+" "+orderBy+" LIMIT $1 OFFSET $2").
		WithArgs(10, 0).
		WillReturnRows(productRows().
			AddRow(int64(1), "A1", "Kopi", int64(15000), int64(10), true, createdAt).
			AddRow(int64(2), "A2", "Teh", int64(5000), int64(3), true, createdAt).
			RowError(1, errBoom)) // error muncul saat iterasi row kedua

	products, total, err := repo.Search(context.Background(), &model.SearchProductRequest{Page: 1, Size: 10})
	if !errors.Is(err, errBoom) {
		t.Fatalf("err = %v, want %v", err, errBoom)
	}
	if products != nil || total != 0 {
		t.Errorf("got products=%v total=%d, want nil and 0", products, total)
	}
}
