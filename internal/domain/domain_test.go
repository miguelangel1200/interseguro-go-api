package domain

import (
	"errors"
	"testing"
)

func TestRotate90_Empty(t *testing.T) {
	var m Matrix
	got := m.Rotate90()
	if len(got) != 0 {
		t.Fatalf("expected empty matrix, got %v", got)
	}
}

func TestRotate90_SingleRow(t *testing.T) {
	// Matriz 1x3 -> 3x1.
	m := Matrix{{1, 2, 3}}
	got := m.Rotate90()
	want := Matrix{{1}, {2}, {3}}
	if !matricesEqual(got, want) {
		t.Fatalf("rotate90 single row: got %v, want %v", got, want)
	}
}

func TestValidationError_Error(t *testing.T) {
	ve := &ValidationError{
		Code:     "MATRIX_NOT_RECTANGULAR",
		Message:  "matrix must be rectangular",
		Row:      1,
		Expected: 2,
		Got:      3,
	}
	if got := ve.Error(); got != "matrix must be rectangular" {
		t.Fatalf("unexpected message: %q", got)
	}
}

func TestErrors_Sentinels(t *testing.T) {
	if ErrEmptyMatrix.Error() == "" {
		t.Fatal("ErrEmptyMatrix must have a message")
	}
	if ErrLinearDependent.Error() == "" {
		t.Fatal("ErrLinearDependent must have a message")
	}
	if ErrMatrixTooLarge.Error() == "" {
		t.Fatal("ErrMatrixTooLarge must have a message")
	}
}

func TestNewMatrix_TooLarge(t *testing.T) {
	// Matriz 101x1 supera el límite de filas.
	rows := make([][]float64, MaxMatrixDimension+1)
	for i := range rows {
		rows[i] = []float64{1}
	}
	if _, err := NewMatrix(rows); !errors.Is(err, ErrMatrixTooLarge) {
		t.Fatalf("expected ErrMatrixTooLarge for rows, got %v", err)
	}

	// Matriz 1x101 supera el límite de columnas.
	wide := make([][]float64, 1)
	wide[0] = make([]float64, MaxMatrixDimension+1)
	if _, err := NewMatrix(wide); !errors.Is(err, ErrMatrixTooLarge) {
		t.Fatalf("expected ErrMatrixTooLarge for cols, got %v", err)
	}

	// Una matriz exactamente en el límite es válida.
	ok := make([][]float64, MaxMatrixDimension)
	for i := range ok {
		ok[i] = make([]float64, MaxMatrixDimension)
	}
	if _, err := NewMatrix(ok); err != nil {
		t.Fatalf("expected valid matrix at limit, got %v", err)
	}
}
