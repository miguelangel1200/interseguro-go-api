package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	"interseguro/go-api/internal/application/ports"
	"interseguro/go-api/internal/domain"
)

func decodeProcessBody(t *testing.T, resp *http.Response) map[string]interface{} {
	t.Helper()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	return out
}

func TestProcess_InvalidJSONBody(t *testing.T) {
	app := testApp(t, &mockStatisticsRepo{stats: domain.Statistics{}})
	token := signTestToken("test-secret")
	resp := doPost(app, "/process", "no-es-json", token)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if body := decodeProcessBody(t, resp); body["code"] != "BAD_REQUEST" {
		t.Fatalf("unexpected body: %v", body)
	}
}

func TestProcess_EmptyMatrix(t *testing.T) {
	app := testApp(t, &mockStatisticsRepo{stats: domain.Statistics{}})
	token := signTestToken("test-secret")
	resp := doPost(app, "/process", `{"matrix":[]}`, token)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	body := decodeProcessBody(t, resp)
	if body["code"] != "MATRIX_EMPTY" {
		t.Fatalf("expected MATRIX_EMPTY, got %v", body)
	}
}

func TestProcess_LinearDependent(t *testing.T) {
	app := testApp(t, &mockStatisticsRepo{stats: domain.Statistics{}})
	token := signTestToken("test-secret")
	resp := doPost(app, "/process", `{"matrix":[[1,2],[2,4]]}`, token)
	if resp.StatusCode != fiber.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", resp.StatusCode)
	}
	body := decodeProcessBody(t, resp)
	if body["code"] != "MATRIX_LINEARLY_DEPENDENT" {
		t.Fatalf("expected MATRIX_LINEARLY_DEPENDENT, got %v", body)
	}
}

func TestProcess_MatrixTooLarge(t *testing.T) {
	app := testApp(t, &mockStatisticsRepo{stats: domain.Statistics{}})
	token := signTestToken("test-secret")

	// Matriz 101x1 (JSON generado dinámicamente para no inflar el test).
	rows := make([]string, domain.MaxMatrixDimension+1)
	for i := range rows {
		rows[i] = "[1]"
	}
	payload := `{"matrix":[` + strings.Join(rows, ",") + `]}`
	resp := doPost(app, "/process", payload, token)
	if resp.StatusCode != fiber.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", resp.StatusCode)
	}
	body := decodeProcessBody(t, resp)
	if body["code"] != "MATRIX_TOO_LARGE" {
		t.Fatalf("expected MATRIX_TOO_LARGE, got %v", body)
	}
}

func TestProcess_NodeUnreachable(t *testing.T) {
	repo := &mockStatisticsRepo{err: errors.New("node api down")}
	app := testApp(t, repo)
	token := signTestToken("test-secret")
	resp := doPost(app, "/process", `{"matrix":[[1,2],[3,4]]}`, token)

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeProcessBody(t, resp)
	if body["message"] != "processed locally; node api unreachable" {
		t.Fatalf("unexpected message: %v", body["message"])
	}
	errInfo, ok := body["error"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected error object, got %v", body["error"])
	}
	if errInfo["code"] != "NODE_API_UNAVAILABLE" {
		t.Fatalf("unexpected error code: %v", errInfo)
	}
	// El detalle interno (p. ej. la URL del servicio Node) no debe filtrarse.
	if errInfo["error"] == "node api down" {
		t.Fatalf("internal error detail leaked to the client: %v", errInfo["error"])
	}
	if errInfo["error"] != "statistics service is temporarily unavailable" {
		t.Fatalf("unexpected masked error message: %v", errInfo["error"])
	}
	// El resultado local debe incluir las 4 matrices.
	result, ok := body["result"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected result object, got %v", body["result"])
	}
	for _, key := range []string{"original", "rotated", "Q", "R"} {
		if _, ok := result[key]; !ok {
			t.Fatalf("result missing %q: %v", key, result)
		}
	}
}

func TestProcess_Success_IncludesStatistics(t *testing.T) {
	repo := &mockStatisticsRepo{stats: domain.Statistics{"mean": 2.5}}
	app := testApp(t, repo)
	token := signTestToken("test-secret")
	resp := doPost(app, "/process", `{"matrix":[[1,2],[3,4]]}`, token)

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeProcessBody(t, resp)
	if body["message"] != "processed successfully" {
		t.Fatalf("unexpected message: %v", body["message"])
	}
	stats, ok := body["statistics"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected statistics, got %v", body["statistics"])
	}
	if stats["mean"] != 2.5 {
		t.Fatalf("unexpected statistics: %v", stats)
	}
}

func TestProcess_ResultMatricesShape(t *testing.T) {
	app := testApp(t, &mockStatisticsRepo{stats: domain.Statistics{}})
	token := signTestToken("test-secret")
	resp := doPost(app, "/process", `{"matrix":[[1,2],[3,4]]}`, token)

	body := decodeProcessBody(t, resp)
	result, ok := body["result"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected result, got %v", body["result"])
	}

	// A = [[1,2],[3,4]] rotada = [[3,1],[4,2]].
	rotated, _ := result["rotated"].([]interface{})
	if len(rotated) != 2 {
		t.Fatalf("unexpected rotated: %v", result["rotated"])
	}
	firstRow, _ := rotated[0].([]interface{})
	if firstRow[0].(float64) != 3 || firstRow[1].(float64) != 1 {
		t.Fatalf("unexpected rotated first row: %v", firstRow)
	}
}

// Asegura que mockStatisticsRepo siga cumpliendo el puerto (guardia de compilación).
var _ ports.StatisticsRepository = (*mockStatisticsRepo)(nil)
