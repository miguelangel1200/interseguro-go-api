// Package domain contiene las entidades y reglas de negocio puras de la API.
// No depende de Fiber, HTTP ni de ningún framework externo.
package domain

// Matrix es una matriz rectangular de números de punto flotante.
// Se representa como una lista de filas: Matrix[fila][columna].
type Matrix [][]float64

// MaxMatrixDimension es el límite de filas/columnas permitido para prevenir
// cargas excesivas (la factorización QR es O(n³)).
const MaxMatrixDimension = 100

// QRResult es el resultado de la factorización QR de una matriz A:
// A = Q·R, donde Q es ortogonal (m×n) y R triangular superior (n×n).
type QRResult struct {
	Original Matrix `json:"original"`
	Rotated  Matrix `json:"rotated"`
	Q        Matrix `json:"Q"`
	R        Matrix `json:"R"`
}

// NewMatrix valida y normaliza una matriz de entrada.
// Devuelve un error si la matriz está vacía, no es rectangular o supera las
// dimensiones máximas permitidas.
func NewMatrix(rows [][]float64) (Matrix, error) {
	if len(rows) == 0 {
		return nil, ErrEmptyMatrix
	}
	cols := len(rows[0])
	if cols == 0 {
		return nil, ErrEmptyMatrix
	}
	if len(rows) > MaxMatrixDimension || cols > MaxMatrixDimension {
		return nil, ErrMatrixTooLarge
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
