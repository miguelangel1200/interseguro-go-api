package domain

import "errors"

// Errores de dominio de la API.
var (
	// ErrEmptyMatrix indica que la matriz de entrada está vacía.
	ErrEmptyMatrix = errors.New("matrix must not be empty")

	// ErrLinearDependent indica que las columnas de la matriz son linealmente
	// dependientes y por tanto la factorización QR no es posible.
	ErrLinearDependent = errors.New("matrix columns are linearly dependent, QR decomposition is not possible")
)

// ValidationError representa un error de validación del dominio con detalles
// estructurados que el adaptador HTTP puede exponer.
type ValidationError struct {
	Code     string
	Message  string
	Row      int
	Expected int
	Got      int
}

// Error implementa la interfaz error.
func (e *ValidationError) Error() string {
	return e.Message
}
