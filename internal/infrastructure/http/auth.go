package http

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// ContextKeyToken es la clave bajo la que se guarda el token JWT autenticado
// en el contexto de la petición.
const ContextKeyToken = "auth.token"

// AuthConfig contiene los parámetros de verificación del token JWT.
type AuthConfig struct {
	Secret   string
	Issuer   string
	Audience string
}

// JWTAuth construye un middleware de autenticación que verifica el token
// `Authorization: Bearer <token>` (HS256) y lo guarda en el contexto.
func JWTAuth(cfg AuthConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			return &HTTPError{
				Status: fiber.StatusUnauthorized,
				ErrorResponse: ErrorResponse{
					Error: "missing bearer token",
					Code:  "UNAUTHORIZED",
				},
			}
		}

		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		if _, err := verifyJWT(token, cfg); err != nil {
			return &HTTPError{
				Status: fiber.StatusUnauthorized,
				ErrorResponse: ErrorResponse{
					Error: "invalid or expired token",
					Code:  "UNAUTHORIZED",
				},
			}
		}

		// Guarda el token para que el handler pueda propagarlo al servicio Node.
		c.Locals(ContextKeyToken, token)
		return c.Next()
	}
}

// verifyJWT valida la firma, issuer y audience del token.
func verifyJWT(token string, cfg AuthConfig) (*jwt.RegisteredClaims, error) {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fiber.ErrUnauthorized
		}
		return []byte(cfg.Secret), nil
	}, jwt.WithIssuer(cfg.Issuer), jwt.WithAudience(cfg.Audience))
	if err != nil {
		return nil, err
	}
	return claims, nil
}
