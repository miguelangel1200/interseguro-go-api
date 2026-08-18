package domain

import "testing"

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
}
