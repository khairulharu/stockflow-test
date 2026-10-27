package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v2"

	"stockflow/internal/config"
	httpd "stockflow/internal/delivery/http"
	"stockflow/internal/repository"
	"stockflow/internal/usecase"
)

func newApp(t *testing.T) (*fiber.App, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	repo := repository.NewProductRepository(db)
	uc := usecase.NewProductUseCase(repo, config.NewValidator())
	ctrl := httpd.NewProductController(uc)

	app := fiber.New()
	app.Get("/api/products", ctrl.Search)
	app.Post("/api/products", ctrl.Create)
	app.Put("/api/products/:id", ctrl.Update)
	return app, mock
}

func doRequest(t *testing.T, app *fiber.App, method, url, body string) *http.Response {
	t.Helper()

	req := httptest.NewRequest(method, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	return resp
}

func TestEndpoint_CreateProduct_Created(t *testing.T) {
	app, mock := newApp(t)

	mock.ExpectQuery(existsSKUQuery).WithArgs("ABC-001").WillReturnRows(existsRows(false))
	mock.ExpectBegin()
	mock.ExpectQuery(createProductQuery).
		WithArgs("ABC-001", "Kopi", int64(15000), 10).
		WillReturnRows(idRows(1))
	mock.ExpectExec(insertMovementQuery).
		WithArgs(int64(1), "in", 10, "").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	resp := doRequest(t, app, http.MethodPost, "/api/products",
		`{"sku":"abc-001","name":"Kopi","price":15000,"stock":10}`)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestEndpoint_CreateProduct_StatusCodes(t *testing.T) {
	tests := []struct {
		name  string
		body  string
		setup func(m sqlmock.Sqlmock)
		want  int
	}{
		{name: "invalid json", body: `{not json`, want: http.StatusBadRequest},
		{name: "negative stock", body: `{"sku":"A1","name":"Kopi","price":1,"stock":-1}`, want: http.StatusBadRequest},
		{name: "empty sku", body: `{"sku":"  ","name":"Kopi","price":1,"stock":1}`, want: http.StatusBadRequest},
		{
			name: "duplicate sku",
			body: `{"sku":"abc-001","name":"Kopi","price":1,"stock":1}`,
			setup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(existsSKUQuery).WithArgs("ABC-001").WillReturnRows(existsRows(true))
			},
			want: http.StatusConflict,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app, mock := newApp(t)
			if tc.setup != nil {
				tc.setup(mock)
			}

			resp := doRequest(t, app, http.MethodPost, "/api/products", tc.body)
			if resp.StatusCode != tc.want {
				t.Errorf("status = %d, want %d", resp.StatusCode, tc.want)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet expectations: %v", err)
			}
		})
	}
}

func TestEndpoint_SearchProducts_ResponseShapeAndClamp(t *testing.T) {
	app, mock := newApp(t)

	// page=0 -> 1, limit=1000 -> 100
	mock.ExpectQuery(countQuery).WithArgs("").WillReturnRows(countRows(1))
	mock.ExpectQuery(searchQuery).
		WithArgs("", 100, 0).
		WillReturnRows(productWithMovementRows().
			AddRow(int64(1), "A1", "Kopi", int64(15000), int64(10), true, createdAt, nil))

	resp := doRequest(t, app, http.MethodGet, "/api/products?page=0&limit=1000", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, key := range []string{"data", "page", "limit", "total"} {
		if _, ok := body[key]; !ok {
			t.Errorf("response missing key %q", key)
		}
	}
	if _, ok := body["paging"]; ok {
		t.Error("response must not contain nested \"paging\"")
	}
	if body["page"] != float64(1) || body["limit"] != float64(100) || body["total"] != float64(1) {
		t.Errorf("page/limit/total = %v/%v/%v, want 1/100/1", body["page"], body["limit"], body["total"])
	}

	items := body["data"].([]any)
	if _, ok := items[0].(map[string]any)["last_movement_at"]; !ok {
		t.Error("item missing key last_movement_at")
	}
}

func TestEndpoint_SearchProducts_EmptyResultIsArray(t *testing.T) {
	app, mock := newApp(t)

	mock.ExpectQuery(countQuery).WithArgs("zzz").WillReturnRows(countRows(0))

	resp := doRequest(t, app, http.MethodGet, "/api/products?search=zzz", "")
	var body map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&body)

	data, ok := body["data"].([]any)
	if !ok || len(data) != 0 {
		t.Errorf("data = %v, want empty array [] (bukan null)", body["data"])
	}
	if body["total"] != float64(0) {
		t.Errorf("total = %v, want 0", body["total"])
	}
}

func TestEndpoint_UpdateProduct_StatusCodes(t *testing.T) {
	tests := []struct {
		name  string
		id    string
		body  string
		setup func(m sqlmock.Sqlmock)
		want  int
	}{
		{
			name: "invalid json",
			id:   "1",
			body: `{not json`,
			want: http.StatusBadRequest,
		},
		{
			name: "invalid id param",
			id:   "abc",
			body: `{"name":"Kopi","price":100,"is_active":true}`,
			want: http.StatusBadRequest,
		},
		{
			name: "validation fails",
			id:   "1",
			body: `{"name":"","price":100,"is_active":true}`, // name kosong
			want: http.StatusBadRequest,
		},
		{
			name: "product not found",
			id:   "999", // ID yang tidak ada
			body: `{"name":"Kopi","price":100,"is_active":true}`,
			setup: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(updateProductQuery).
					WithArgs("Kopi", int64(100), true, int64(999)).
					WillReturnRows(productRows()) // return 0 rows -> ErrNoRows
			},
			want: http.StatusNotFound, // Harusnya 404!
		},
		{
			name: "success update",
			id:   "1",
			body: `{"name":"Kopi Susu","price":20000,"is_active":false}`,
			setup: func(m sqlmock.Sqlmock) {
				rows := productRows().
					AddRow(int64(1), "ABC-001", "Kopi Susu", int64(20000), int64(10), false, createdAt)
				m.ExpectQuery(updateProductQuery).
					WithArgs("Kopi Susu", int64(20000), false, int64(1)).
					WillReturnRows(rows)
			},
			want: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app, mock := newApp(t)
			// Pastikan route sudah didaftarkan di newApp
			// app.Put("/api/products/:id", ctrl.Update)

			if tc.setup != nil {
				tc.setup(mock)
			}

			resp := doRequest(t, app, http.MethodPut, "/api/products/"+tc.id, tc.body)
			if resp.StatusCode != tc.want {
				t.Errorf("status = %d, want %d", resp.StatusCode, tc.want)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet expectations: %v", err)
			}
		})
	}
}
