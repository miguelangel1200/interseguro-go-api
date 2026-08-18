// Package services contiene los casos de uso (servicios de aplicación) de la
// arquitectura hexagonal. Orquestan las entidades del dominio y los puertos
// de salida sin depender de frameworks concretos.
package services

import (
	"context"

	"interseguro/go-api/internal/domain"
	"interseguro/go-api/internal/application/ports"
)

// MatrixService implementa el caso de uso MatrixProcessor.
type MatrixService struct {
	repo ports.StatisticsRepository
}

// NewMatrixService construye el caso de uso con su puerto de salida.
func NewMatrixService(repo ports.StatisticsRepository) *MatrixService {
	return &MatrixService{repo: repo}
}

// Process rota la matriz, calcula su factorización QR y, si el servicio de
// estadísticas está disponible, delega el cómputo de estadísticas. Si el
// servicio remoto no responde, devuelve el resultado local con NodeReachable
// en false (comportamiento resiliente). authToken es el JWT que se propaga al
// servicio remoto.
func (s *MatrixService) Process(ctx context.Context, m domain.Matrix, authToken string) (ports.ProcessResult, error) {
	// 1. Rotación de la matriz.
	rotated := m.Rotate90()

	// 2. Factorización QR de la matriz original.
	Q, R, err := m.FactorizeQR()
	if err != nil {
		return ports.ProcessResult{}, err
	}

	result := ports.ProcessResult{
		QR: domain.QRResult{
			Original: m,
			Rotated:  rotated,
			Q:        Q,
			R:        R,
		},
		NodeReachable: true,
	}

	// 3. Envío al servicio de estadísticas (puerto de salida), propagando el token.
	stats, sendErr := s.repo.SendStatistics(ctx, result.QR, authToken)
	if sendErr != nil {
		// El fallo del servicio secundario no invalida el resultado local.
		result.NodeReachable = false
		result.NodeError = sendErr
		return result, nil
	}

	result.Statistics = stats
	return result, nil
}
