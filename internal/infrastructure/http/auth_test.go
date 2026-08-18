package http

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func tokenWith(secret string, mutate func(c *jwt.RegisteredClaims)) string {
	claims := jwt.RegisteredClaims{
		Subject:  "test",
		Issuer:   "i",
		Audience: jwt.ClaimStrings{"a"},
	}
	if mutate != nil {
		mutate(&claims)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		panic(err)
	}
	return signed
}

func TestVerifyJWT_Valid(t *testing.T) {
	cfg := AuthConfig{Secret: "s", Issuer: "i", Audience: "a"}
	if _, err := verifyJWT(tokenWith("s", nil), cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVerifyJWT_WrongSecret(t *testing.T) {
	cfg := AuthConfig{Secret: "s", Issuer: "i", Audience: "a"}
	if _, err := verifyJWT(tokenWith("otro-secreto", nil), cfg); err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestVerifyJWT_WrongIssuer(t *testing.T) {
	cfg := AuthConfig{Secret: "s", Issuer: "i", Audience: "a"}
	if _, err := verifyJWT(tokenWith("s", func(c *jwt.RegisteredClaims) { c.Issuer = "otro" }), cfg); err == nil {
		t.Fatal("expected error for wrong issuer")
	}
}

func TestVerifyJWT_WrongAudience(t *testing.T) {
	cfg := AuthConfig{Secret: "s", Issuer: "i", Audience: "a"}
	if _, err := verifyJWT(tokenWith("s", func(c *jwt.RegisteredClaims) { c.Audience = jwt.ClaimStrings{"otra"} }), cfg); err == nil {
		t.Fatal("expected error for wrong audience")
	}
}

func TestVerifyJWT_Expired(t *testing.T) {
	cfg := AuthConfig{Secret: "s", Issuer: "i", Audience: "a"}
	if _, err := verifyJWT(tokenWith("s", func(c *jwt.RegisteredClaims) {
		c.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Hour))
	}), cfg); err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestVerifyJWT_NonHMAC(t *testing.T) {
	cfg := AuthConfig{Secret: "s", Issuer: "i", Audience: "a"}
	claims := jwt.RegisteredClaims{Subject: "x"}
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	signed, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign none token: %v", err)
	}
	if _, err := verifyJWT(signed, cfg); err == nil {
		t.Fatal("expected error for non-HMAC signing method")
	}
}

// decodeMap lee y decodifica un body JSON genérico.
func decodeMap(t *testing.T, body io.Reader) map[string]interface{} {
	t.Helper()
	raw, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	return out
}

func TestJWTAuth_StoresTokenInLocals(t *testing.T) {
	cfg := AuthConfig{Secret: "s", Issuer: "i", Audience: "a"}
	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler})
	app.Get("/p", JWTAuth(cfg), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"token": c.Locals(ContextKeyToken)})
	})

	valid := tokenWith("s", nil)
	req := httptest.NewRequest(http.MethodGet, "/p", nil)
	req.Header.Set("Authorization", "Bearer "+valid)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body := decodeMap(t, resp.Body)
	if body["token"] != valid {
		t.Fatalf("token not stored in locals: %v", body)
	}
}

func TestJWTAuth_RejectsWithoutHeader(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler})
	app.Get("/p", JWTAuth(AuthConfig{Secret: "s", Issuer: "i", Audience: "a"}), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	resp, _ := app.Test(httptest.NewRequest(http.MethodGet, "/p", nil), -1)
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	body := decodeMap(t, resp.Body)
	if body["code"] != "UNAUTHORIZED" {
		t.Fatalf("unexpected body: %v", body)
	}
}

func TestJWTAuth_RejectsNonBearerHeader(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler})
	app.Get("/p", JWTAuth(AuthConfig{Secret: "s", Issuer: "i", Audience: "a"}), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/p", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	resp, _ := app.Test(req, -1)
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}
