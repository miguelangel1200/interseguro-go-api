package services

import (
	"context"
	"errors"
	"testing"

	"interseguro/go-api/internal/application/ports"
	"interseguro/go-api/internal/domain"
)

// fakeStatsRepo implementa el puerto de salida StatisticsRepository con
// respuestas controladas (mock de datos) para aislar el caso de uso.
type fakeStatsRepo struct {
	stats domain.Statistics
	err   error
}

func (f *fakeStatsRepo) SendStatistics(_ context.Context, _ domain.QRResult, _ string) (domain.Statistics, error) {
	return f.stats, f.err
}

var _ ports.StatisticsRepository = (*fakeStatsRepo)(nil)

func TestMatrixService_Process_Success(t *testing.T) {
	repo := &fakeStatsRepo{stats: domain.Statistics{"ok": true}}
	svc := NewMatrixService(repo)

	m := domain.Matrix{{1, 2}, {3, 4}}
	res, err := svc.Process(context.Background(), m, "jwt-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !res.NodeReachable {
		t.Fatal("expected NodeReachable = true")
	}
	if res.NodeError != nil {
		t.Fatalf("unexpected node error: %v", res.NodeError)
	}
	if res.Statistics["ok"] != true {
		t.Fatalf("statistics not propagated: %v", res.Statistics)
	}
	if res.QR.Original[0][0] != 1 || res.QR.Original[1][1] != 4 {
		t.Fatalf("original matrix not preserved: %v", res.QR.Original)
	}
	// Rotación de [[1,2],[3,4]] = [[3,1],[4,2]].
	if res.QR.Rotated[0][0] != 3 || res.QR.Rotated[1][1] != 2 {
		t.Fatalf("unexpected rotated matrix: %v", res.QR.Rotated)
	}
	// Q debe ser ortogonal y R triangular superior.
	if len(res.QR.Q) != 2 || len(res.QR.R) != 2 {
		t.Fatalf("unexpected QR dimensions: Q=%v R=%v", res.QR.Q, res.QR.R)
	}
}

func TestMatrixService_Process_NodeUnavailable(t *testing.T) {
	repo := &fakeStatsRepo{err: errors.New("node api down")}
	svc := NewMatrixService(repo)

	res, err := svc.Process(context.Background(), domain.Matrix{{1, 2}, {3, 4}}, "token")
	if err != nil {
		t.Fatalf("expected resilient result without error, got %v", err)
	}
	if res.NodeReachable {
		t.Fatal("expected NodeReachable = false")
	}
	if res.NodeError == nil {
		t.Fatal("expected NodeError to be set")
	}
	// El resultado local (QR) debe seguir disponible.
	if len(res.QR.Q) != 2 || len(res.QR.R) != 2 {
		t.Fatalf("QR result missing when node is down: %+v", res)
	}
}

func TestMatrixService_Process_QRFailure(t *testing.T) {
	svc := NewMatrixService(&fakeStatsRepo{})

	// Columnas linealmente dependientes: QR no es posible.
	m := domain.Matrix{{1, 2}, {2, 4}}
	if _, err := svc.Process(context.Background(), m, "token"); err == nil {
		t.Fatal("expected error for linearly dependent matrix")
	}
}
