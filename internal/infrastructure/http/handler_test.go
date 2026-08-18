package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"interseguro/go-api/internal/application/ports"
	"interseguro/go-api/internal/application/services"
	"interseguro/go-api/internal/domain"
)

// mockStatisticsRepo implementa el puerto de salida StatisticsRepository para
// las pruebas de integración del adaptador HTTP.
type mockStatisticsRepo struct {
	stats domain.Statistics
	err   error
}

func (m *mockStatisticsRepo) SendStatistics(_ context.Context, _ domain.QRResult, _ string) (domain.Statistics, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.stats, nil
}

var _ ports.StatisticsRepository = (*mockStatisticsRepo)(nil)

// signTestToken genera un token JWT válido para las pruebas (mismo issuer y
// audience que exige la verificación).
func signTestToken(secret string) string {
	claims := jwt.RegisteredClaims{
		Subject:  "test",
		Issuer:   "interseguro",
		Audience: jwt.ClaimStrings{"interseguro-api"},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(secret))
	return signed
}

func testApp(t *testing.T, repo ports.StatisticsRepository) *fiber.App {
	t.Helper()
	authCfg := AuthConfig{Secret: "test-secret", Issuer: "interseguro", Audience: "interseguro-api"}
	service := services.NewMatrixService(repo)
	handler := NewMatrixHandler(service)

	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler})
	app.Post("/process", JWTAuth(authCfg), handler.Process)
	return app
}

func doPost(app *fiber.App, path, body, token string) *http.Response {
	req := httptest.NewRequest("POST", path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, _ := app.Test(req, -1)
	return resp
}

func TestProcess_RequiresToken(t *testing.T) {
	app := testApp(t, &mockStatisticsRepo{stats: domain.Statistics{}})
	resp := doPost(app, "/process", `{"matrix":[[1,2],[3,4]]}`, "")
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestProcess_InvalidToken(t *testing.T) {
	app := testApp(t, &mockStatisticsRepo{stats: domain.Statistics{}})
	resp := doPost(app, "/process", `{"matrix":[[1,2],[3,4]]}`, "not-a-token")
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestProcess_ValidToken_Success(t *testing.T) {
	repo := &mockStatisticsRepo{stats: domain.Statistics{"ok": true}}
	app := testApp(t, repo)
	token := signTestToken("test-secret")
	resp := doPost(app, "/process", `{"matrix":[[1,2],[3,4]]}`, token)

	if resp.StatusCode != fiber.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(body))
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	var body map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body["message"] != "processed successfully" {
		t.Fatalf("unexpected message: %v", body["message"])
	}
}

func TestProcess_NotRectangular(t *testing.T) {
	app := testApp(t, &mockStatisticsRepo{stats: domain.Statistics{}})
	token := signTestToken("test-secret")
	resp := doPost(app, "/process", `{"matrix":[[1,2],[3]]}`, token)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
