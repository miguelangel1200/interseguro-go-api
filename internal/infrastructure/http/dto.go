// Package http contiene los adaptadores HTTP de entrada (controllers) y sus
// DTOs de request/response, así como el manejo de errores de la API.
package http

import "interseguro/go-api/internal/domain"

// ProcessRequest es el DTO de entrada del endpoint POST /process.
type ProcessRequest struct {
	Matrix domain.Matrix `json:"matrix"`
}

// ProcessResponse es el DTO de salida del endpoint POST /process.
type ProcessResponse struct {
	Message    string                 `json:"message"`
	Result     ProcessResultDTO       `json:"result"`
	Statistics domain.Statistics      `json:"statistics,omitempty"`
	Error      *ErrorResponse         `json:"error,omitempty"`
}

// ProcessResultDTO describe la factorización QR dentro de la respuesta.
type ProcessResultDTO struct {
	Original domain.Matrix `json:"original"`
	Rotated  domain.Matrix `json:"rotated"`
	Q        domain.Matrix `json:"Q"`
	R        domain.Matrix `json:"R"`
}

// StatisticsResponseDTO es el DTO devuelto cuando el servicio de estadísticas
// respondió correctamente.
type StatisticsResponseDTO struct {
	Message    string            `json:"message"`
	Result     ProcessResultDTO  `json:"result"`
	Statistics domain.Statistics `json:"statistics"`
}
