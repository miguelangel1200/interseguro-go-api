// Package ports define los puertos (interfaces) de la arquitectura hexagonal.
// Los puertos de entrada describen los casos de uso de la aplicación y los
// puertos de salida las dependencias externas que la aplicación necesita.
package ports

import (
	"context"

	"interseguro/go-api/internal/domain"
)

// MatrixProcessor es un puerto de entrada (caso de uso): define la operación
// de procesar una matriz (rotar + factorizar QR) y delegar el cálculo de
// estadísticas a un servicio externo.
type MatrixProcessor interface {
	Process(ctx context.Context, m domain.Matrix, authToken string) (ProcessResult, error)
}

// ProcessResult agrupa el resultado del caso de uso: la factorización QR más
// las estadísticas opcionales devueltas por el servicio remoto.
type ProcessResult struct {
	QR         domain.QRResult
	Statistics domain.Statistics
	// NodeReachable indica si el servicio de estadísticas respondió.
	NodeReachable bool
	// NodeError guarda el error de comunicación si el servicio no respondió.
	NodeError error
}

// StatisticsRepository es un puerto de salida (adaptador secundario): define
// cómo enviar el resultado de la factorización a un servicio de estadísticas
// y recuperar el cómputo resultante. authToken es el token JWT que se propaga
// al servicio remoto para autenticar la llamada.
type StatisticsRepository interface {
	SendStatistics(ctx context.Context, res domain.QRResult, authToken string) (domain.Statistics, error)
}
