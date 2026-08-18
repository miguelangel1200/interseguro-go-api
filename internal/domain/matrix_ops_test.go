package domain

import (
	"errors"
	"math"
	"testing"
)

func TestNewMatrix_Valid(t *testing.T) {
	m, err := NewMatrix([][]float64{{1, 2}, {3, 4}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m) != 2 || len(m[0]) != 2 {
		t.Fatalf("unexpected dimensions: %dx%d", len(m), len(m[0]))
	}
}

func TestNewMatrix_Empty(t *testing.T) {
	if _, err := NewMatrix([][]float64{}); !errors.Is(err, ErrEmptyMatrix) {
		t.Fatalf("expected ErrEmptyMatrix, got %v", err)
	}
	if _, err := NewMatrix([][]float64{{}}); !errors.Is(err, ErrEmptyMatrix) {
		t.Fatalf("expected ErrEmptyMatrix for empty rows, got %v", err)
	}
}

func TestNewMatrix_NotRectangular(t *testing.T) {
	_, err := NewMatrix([][]float64{{1, 2}, {3}})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if ve.Code != "MATRIX_NOT_RECTANGULAR" {
		t.Fatalf("unexpected code: %s", ve.Code)
	}
}

func TestRotate90_Square(t *testing.T) {
	m := Matrix{{1, 2}, {3, 4}}
	got := m.Rotate90()
	want := Matrix{{3, 1}, {4, 2}}
	if !matricesEqual(got, want) {
		t.Fatalf("rotate90: got %v, want %v", got, want)
	}
	// La matriz original no debe modificarse.
	if !matricesEqual(m, Matrix{{1, 2}, {3, 4}}) {
		t.Fatalf("original matrix was modified: %v", m)
	}
}

func TestRotate90_Rectangular(t *testing.T) {
	m := Matrix{{1, 2, 3}, {4, 5, 6}}
	got := m.Rotate90()
	// 2x3 -> 3x2.
	want := Matrix{{4, 1}, {5, 2}, {6, 3}}
	if !matricesEqual(got, want) {
		t.Fatalf("rotate90 rectangular: got %v, want %v", got, want)
	}
}

func TestFactorizeQR_2x2(t *testing.T) {
	m := Matrix{{1, 2}, {3, 4}}
	Q, R, err := m.FactorizeQR()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// R debe ser triangular superior (R[1][0] == 0).
	if math.Abs(R[1][0]) > 1e-12 {
		t.Fatalf("R is not upper triangular: %v", R)
	}

	// Ortogonalidad de Q: QᵀQ ≈ I.
	if err := checkOrthogonal(Q); err != nil {
		t.Fatalf("Q not orthogonal: %v", err)
	}

	// Reconstrucción: A ≈ Q·R.
	if err := checkReconstruction(m, Q, R); err != nil {
		t.Fatalf("reconstruction failed: %v", err)
	}
}

func TestFactorizeQR_3x2(t *testing.T) {
	m := Matrix{{2, 1}, {-1, 3}, {0, 2}}
	Q, R, err := m.FactorizeQR()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(Q) != 3 || len(Q[0]) != 2 {
		t.Fatalf("Q dimensions: got %dx%d, want 3x2", len(Q), len(Q[0]))
	}
	if len(R) != 2 || len(R[0]) != 2 {
		t.Fatalf("R dimensions: got %dx%d, want 2x2", len(R), len(R[0]))
	}
	if math.Abs(R[1][0]) > 1e-12 {
		t.Fatalf("R is not upper triangular: %v", R)
	}
	if err := checkOrthogonal(Q); err != nil {
		t.Fatalf("Q not orthogonal: %v", err)
	}
	if err := checkReconstruction(m, Q, R); err != nil {
		t.Fatalf("reconstruction failed: %v", err)
	}
}

func TestFactorizeQR_LinearDependent(t *testing.T) {
	// Columnas dependientes: la segunda es el doble de la primera.
	m := Matrix{{1, 2}, {2, 4}}
	if _, _, err := m.FactorizeQR(); !errors.Is(err, ErrLinearDependent) {
		t.Fatalf("expected ErrLinearDependent, got %v", err)
	}
}

// --- Helpers de prueba ---

func matricesEqual(a, b Matrix) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if len(a[i]) != len(b[i]) {
			return false
		}
		for j := range a[i] {
			if math.Abs(a[i][j]-b[i][j]) > 1e-9 {
				return false
			}
		}
	}
	return true
}

// checkOrthogonal verifica que Qᵀ·Q ≈ I.
func checkOrthogonal(Q Matrix) error {
	cols := len(Q[0])
	for i := 0; i < cols; i++ {
		for j := 0; j < cols; j++ {
			dot := 0.0
			for k := 0; k < len(Q); k++ {
				dot += Q[k][i] * Q[k][j]
			}
			expected := 0.0
			if i == j {
				expected = 1.0
			}
			if math.Abs(dot-expected) > 1e-9 {
				return errors.New("QᵀQ deviates from identity")
			}
		}
	}
	return nil
}

// checkReconstruction verifica que A ≈ Q·R.
func checkReconstruction(A, Q, R Matrix) error {
	m := len(A)
	n := len(A[0])
	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			sum := 0.0
			for k := 0; k < n; k++ {
				sum += Q[i][k] * R[k][j]
			}
			if math.Abs(sum-A[i][j]) > 1e-9 {
				return errors.New("A ≠ Q·R")
			}
		}
	}
	return nil
}
