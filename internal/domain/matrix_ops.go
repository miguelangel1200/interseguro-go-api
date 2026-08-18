package domain

import "math"

// Rotate90 rota la matriz 90 grados en el sentido de las agujas del reloj.
// Devuelve una nueva matriz sin modificar la original. Complejidad O(n*m).
func (m Matrix) Rotate90() Matrix {
	rows := len(m)
	if rows == 0 {
		return Matrix{}
	}
	cols := len(m[0])

	rotated := make(Matrix, cols)
	for c := 0; c < cols; c++ {
		rotated[c] = make([]float64, rows)
		for r := 0; r < rows; r++ {
			// Fila c de la nueva matriz = columna c de la original invertida.
			rotated[c][rows-1-r] = m[r][c]
		}
	}
	return rotated
}

// FactorizeQR realiza la factorización QR mediante el proceso de
// Gram-Schmidt modificado (numéricamente más estable). Devuelve Q (m×n) y
// R (n×n) tales que A = Q·R, con Q en representación estándar Q[fila][columna].
func (m Matrix) FactorizeQR() (Matrix, Matrix, error) {
	rows := len(m)
	cols := len(m[0])

	Q := newMatrix(rows, cols)
	R := newMatrix(cols, cols)

	// Columnas de la matriz de entrada: col[k] es la columna k.
	colsV := make([][]float64, cols)
	for k := 0; k < cols; k++ {
		colsV[k] = make([]float64, rows)
		for i := 0; i < rows; i++ {
			colsV[k][i] = m[i][k]
		}
	}

	// u[j] es la copia de la columna j que se irá ortonormalizando.
	u := make([][]float64, cols)
	for j := 0; j < cols; j++ {
		u[j] = make([]float64, rows)
		copy(u[j], colsV[j])
	}

	qCols := make([][]float64, cols)

	for j := 0; j < cols; j++ {
		for k := 0; k < j; k++ {
			dot := dotProduct(u[j], qCols[k])
			for i := 0; i < rows; i++ {
				u[j][i] -= dot * qCols[k][i]
			}
		}

		norm := vectorNorm(u[j])
		if norm < 1e-12 {
			return nil, nil, ErrLinearDependent
		}

		qCols[j] = make([]float64, rows)
		for i := 0; i < rows; i++ {
			qCols[j][i] = u[j][i] / norm
		}

		for k := j; k < cols; k++ {
			R[j][k] = dotProduct(qCols[j], colsV[k])
		}
	}

	// Transpone qCols a la representación estándar Q[fila][columna].
	for j := 0; j < cols; j++ {
		for i := 0; i < rows; i++ {
			Q[i][j] = qCols[j][i]
		}
	}

	// Fuerza ceros bajo la diagonal de R.
	for i := 0; i < cols; i++ {
		for j := 0; j < i; j++ {
			R[i][j] = 0
		}
	}

	return Q, R, nil
}

func newMatrix(rows, cols int) Matrix {
	m := make(Matrix, rows)
	for i := range m {
		m[i] = make([]float64, cols)
	}
	return m
}

func dotProduct(a, b []float64) float64 {
	var sum float64
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum
}

func vectorNorm(v []float64) float64 {
	var sum float64
	for _, x := range v {
		sum += x * x
	}
	return math.Sqrt(sum)
}
