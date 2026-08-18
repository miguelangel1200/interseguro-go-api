// Package domain contiene las entidades y reglas de negocio puras de la API.
// No depende de Fiber, HTTP ni de ningún framework externo.
package domain

// Matrix es una matriz rectangular de números de punto flotante.
// Se representa como una lista de filas: Matrix[fila][columna].
type Matrix [][]float64

// QRResult es el resultado de la factorización QR de una matriz A:
// A = Q·R, donde Q es ortogonal (m×n) y R triangular superior (n×n).
type QRResult struct {
	Original Matrix `json:"original"`
	Rotated  Matrix `json:"rotated"`
	Q        Matrix `json:"Q"`
	R        Matrix `json:"R"`
}

// NewMatrix valida y normaliza una matriz de entrada.
// Devuelve un error si la matriz está vacía o no es rectangular.
func NewMatrix(rows [][]float64) (Matrix, error) {
	if len(rows) == 0 {
		return nil, ErrEmptyMatrix
	}
	cols := len(rows[0])
	if cols == 0 {
		return nil, ErrEmptyMatrix
	}
	for i, row := range rows {
		if len(row) != cols {
			return nil, &ValidationError{
				Code:    "MATRIX_NOT_RECTANGULAR",
				Message: "matrix must be rectangular",
				Row:     i,
				Expected: cols,
				Got:     len(row),
			}
		}
	}
	return Matrix(rows), nil
}
