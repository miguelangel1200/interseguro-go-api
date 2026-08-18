package http

import (
	"github.com/gofiber/fiber/v2"

	"interseguro/go-api/internal/domain"
	"interseguro/go-api/internal/application/ports"
)

// MatrixHandler es el adaptador HTTP de entrada (controller) para el endpoint
// POST /process. Depende únicamente del puerto de entrada MatrixProcessor.
type MatrixHandler struct {
	processor ports.MatrixProcessor
}

// NewMatrixHandler construye el controller con su caso de uso.
func NewMatrixHandler(p ports.MatrixProcessor) *MatrixHandler {
	return &MatrixHandler{processor: p}
}

// Process valida la petición, invoca el caso de uso y serializa la respuesta.
func (h *MatrixHandler) Process(c *fiber.Ctx) error {
	var req ProcessRequest
	if err := c.BodyParser(&req); err != nil {
		return newBadRequest("invalid JSON body")
	}

	// Validación de dominio: construye la entidad Matrix.
	matrix, err := domain.NewMatrix(req.Matrix)
	if err != nil {
		return mapDomainError(err)
	}

	// Token JWT autenticado por el middleware, que se propaga al servicio Node.
	authToken, _ := c.Locals(ContextKeyToken).(string)

	// Invoca el caso de uso (puerto de entrada).
	result, err := h.processor.Process(c.Context(), matrix, authToken)
	if err != nil {
		return mapDomainError(err)
	}

	resp := ProcessResponse{
		Result: ProcessResultDTO{
			Original: result.QR.Original,
			Rotated:  result.QR.Rotated,
			Q:        result.QR.Q,
			R:        result.QR.R,
		},
	}

	// Decide la respuesta según la disponibilidad del servicio de estadísticas.
	if result.NodeReachable {
		resp.Message = "processed successfully"
		resp.Statistics = result.Statistics
	} else {
		resp.Message = "processed locally; node api unreachable"
		if result.NodeError != nil {
			// No se expone el detalle interno del error (p. ej. la URL del
			// servicio Node) al cliente; solo se indica que no está disponible.
			resp.Error = &ErrorResponse{
				Error: "statistics service is temporarily unavailable",
				Code:  "NODE_API_UNAVAILABLE",
			}
		}
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}
