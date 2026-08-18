package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"interseguro/go-api/internal/domain"
)

func TestMapDomainError_EmptyMatrix(t *testing.T) {
	he := mapDomainError(domain.ErrEmptyMatrix)
	if he.Status != fiber.StatusBadRequest || he.Code != "MATRIX_EMPTY" {
		t.Fatalf("unexpected mapping: %+v", he)
	}
}

func TestMapDomainError_LinearDependent(t *testing.T) {
	he := mapDomainError(domain.ErrLinearDependent)
	if he.Status != fiber.StatusUnprocessableEntity || he.Code != "MATRIX_LINEARLY_DEPENDENT" {
		t.Fatalf("unexpected mapping: %+v", he)
	}
}

func TestMapDomainError_ValidationError(t *testing.T) {
	ve := &domain.ValidationError{
		Code:     "MATRIX_NOT_RECTANGULAR",
		Message:  "matrix must be rectangular",
		Row:      1,
		Expected: 2,
		Got:      3,
	}
	he := mapDomainError(ve)
	if he.Status != fiber.StatusBadRequest || he.Code != "MATRIX_NOT_RECTANGULAR" {
		t.Fatalf("unexpected mapping: %+v", he)
	}
	if he.Details["row"] != 1 || he.Details["expected"] != 2 || he.Details["got"] != 3 {
		t.Fatalf("unexpected details: %v", he.Details)
	}
	if he.Error() != "matrix must be rectangular" {
		t.Fatalf("unexpected error message: %s", he.Error())
	}
}

func TestMapDomainError_Unknown(t *testing.T) {
	he := mapDomainError(errors.New("boom"))
	if he.Status != fiber.StatusInternalServerError || he.Code != "INTERNAL" {
		t.Fatalf("unexpected mapping: %+v", he)
	}
	if he.Error() == "" {
		t.Fatal("HTTPError must implement error")
	}
}

func TestNewBadRequest(t *testing.T) {
	he := newBadRequest("invalid JSON body")
	if he.Status != fiber.StatusBadRequest || he.Code != "BAD_REQUEST" {
		t.Fatalf("unexpected: %+v", he)
	}
}

// decodeErrorResponse lee y decodifica el body JSON de una respuesta.
func decodeErrorResponse(t *testing.T, body io.Reader) ErrorResponse {
	t.Helper()
	raw, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	var resp ErrorResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("invalid JSON error body: %v", err)
	}
	return resp
}

func TestErrorHandler_HTTPErrorPassthrough(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler})
	app.Get("/x", func(c *fiber.Ctx) error {
		return &HTTPError{Status: fiber.StatusConflict, ErrorResponse: ErrorResponse{Error: "conflicto", Code: "CONFLICT"}}
	})

	resp, _ := app.Test(httptest.NewRequest(http.MethodGet, "/x", nil), -1)
	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
	body := decodeErrorResponse(t, resp.Body)
	if body.Code != "CONFLICT" || body.Error != "conflicto" {
		t.Fatalf("unexpected body: %+v", body)
	}
}

func TestErrorHandler_FiberError(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler})
	app.Get("/x", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusNotFound, "route not found")
	})

	resp, _ := app.Test(httptest.NewRequest(http.MethodGet, "/x", nil), -1)
	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
	body := decodeErrorResponse(t, resp.Body)
	if body.Code != "HTTP_ERROR" {
		t.Fatalf("expected HTTP_ERROR code, got %s", body.Code)
	}
}

func TestErrorHandler_UnknownError(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler})
	app.Get("/x", func(c *fiber.Ctx) error {
		return errors.New("boom interno")
	})

	resp, _ := app.Test(httptest.NewRequest(http.MethodGet, "/x", nil), -1)
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
	body := decodeErrorResponse(t, resp.Body)
	if body.Code != "INTERNAL" {
		t.Fatalf("expected INTERNAL code, got %s", body.Code)
	}
}
