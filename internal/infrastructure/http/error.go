package http

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"interseguro/go-api/internal/domain"
)

// ErrorResponse es el DTO de error normalizado que la API devuelve ante
// cualquier fallo. Mantiene una forma consistente en toda la API.
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Code    string                 `json:"code,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// HTTPError asocia una respuesta de error con un código de estado HTTP.
type HTTPError struct {
	Status int
	ErrorResponse
}

// Error implementa la interfaz error.
func (e *HTTPError) Error() string {
	return e.ErrorResponse.Error
}

// mapDomainError traduce un error de dominio a un HTTPError con su código.
func mapDomainError(err error) *HTTPError {
	switch {
	case errors.Is(err, domain.ErrEmptyMatrix):
		return &HTTPError{
			Status:        fiber.StatusBadRequest,
			ErrorResponse: ErrorResponse{Error: err.Error(), Code: "MATRIX_EMPTY"},
		}
	case errors.Is(err, domain.ErrLinearDependent):
		return &HTTPError{
			Status:        fiber.StatusUnprocessableEntity,
			ErrorResponse: ErrorResponse{Error: err.Error(), Code: "MATRIX_LINEARLY_DEPENDENT"},
		}
	default:
		// Error de validación con detalles (matriz no rectangular).
		var ve *domain.ValidationError
		if errors.As(err, &ve) {
			return &HTTPError{
				Status: fiber.StatusBadRequest,
				ErrorResponse: ErrorResponse{
					Error: ve.Message,
					Code:  ve.Code,
					Details: map[string]interface{}{
						"row":      ve.Row,
						"expected": ve.Expected,
						"got":      ve.Got,
					},
				},
			}
		}
		return &HTTPError{
			Status:        fiber.StatusInternalServerError,
			ErrorResponse: ErrorResponse{Error: "internal server error", Code: "INTERNAL"},
		}
	}
}

// newBadRequest construye un HTTPError de petición inválida.
func newBadRequest(message string) *HTTPError {
	return &HTTPError{
		Status:        fiber.StatusBadRequest,
		ErrorResponse: ErrorResponse{Error: message, Code: "BAD_REQUEST"},
	}
}

// ErrorHandler es el manejador de errores global de Fiber. Convierte cualquier
// error en una respuesta JSON estructurada con código HTTP adecuado.
func ErrorHandler(c *fiber.Ctx, err error) error {
	he, ok := err.(*HTTPError)
	if !ok {
		// Convierte errores nativos de Fiber (p. ej. rutas no encontradas).
		if fe, isFiber := err.(*fiber.Error); isFiber {
			he = &HTTPError{
				Status:        fe.Code,
				ErrorResponse: ErrorResponse{Error: fe.Message, Code: "HTTP_ERROR"},
			}
		} else {
			he = mapDomainError(err)
		}
	}
	return c.Status(he.Status).JSON(he.ErrorResponse)
}
