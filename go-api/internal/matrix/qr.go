package matrix

import (
	"errors"
	"fmt"
	"math"
)

// Por debajo de esta magnitud un valor es ruido de coma flotante.
const epsilon = 1e-12

var ErrEmptyMatrix = errors.New("la matriz no puede estar vacía")

func Dimensions(a [][]float64) (rows, cols int, err error) {
	if len(a) == 0 {
		return 0, 0, ErrEmptyMatrix
	}

	cols = len(a[0])
	if cols == 0 {
		return 0, 0, errors.New("las filas de la matriz no pueden estar vacías")
	}

	for i, row := range a {
		if len(row) != cols {
			return 0, 0, fmt.Errorf("la fila %d tiene %d valores, se esperaban %d", i, len(row), cols)
		}
		for _, value := range row {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return 0, 0, fmt.Errorf("la fila %d trae un valor que no es un número finito", i)
			}
		}
	}

	return len(a), cols, nil
}

// QRDecompose devuelve la factorización QR reducida de a (a = Q * R) usando
// reflexiones de Householder. Para una entrada de m x n, Q es m x k y R es k x n
// con k = min(m, n).
func QRDecompose(a [][]float64) (q, r [][]float64, err error) {
	m, n, err := Dimensions(a)
	if err != nil {
		return nil, nil, err
	}

	k := m
	if n < k {
		k = n
	}

	r = clone(a)
	q = identity(m)

	v := make([]float64, m)
	for j := 0; j < k; j++ {
		norm := 0.0
		for i := j; i < m; i++ {
			norm += r[i][j] * r[i][j]
		}
		norm = math.Sqrt(norm)
		if norm <= epsilon {
			continue
		}

		// Se elige el signo que evita cancelación catastrófica.
		alpha := -norm
		if r[j][j] < 0 {
			alpha = norm
		}

		for i := j; i < m; i++ {
			v[i] = r[i][j]
		}
		v[j] -= alpha

		beta := 0.0
		for i := j; i < m; i++ {
			beta += v[i] * v[i]
		}
		if beta <= epsilon {
			continue
		}
		beta = 2 / beta

		// r = (I - beta*v*v^T) * r, solo sobre el bloque que queda por debajo.
		for c := j; c < n; c++ {
			dot := 0.0
			for i := j; i < m; i++ {
				dot += v[i] * r[i][c]
			}
			dot *= beta
			for i := j; i < m; i++ {
				r[i][c] -= dot * v[i]
			}
		}

		// q = q * (I - beta*v*v^T), acumulando la reflexión.
		for row := 0; row < m; row++ {
			dot := 0.0
			for i := j; i < m; i++ {
				dot += q[row][i] * v[i]
			}
			dot *= beta
			for i := j; i < m; i++ {
				q[row][i] -= dot * v[i]
			}
		}
	}

	// R ya es triangular superior por construcción; se limpian los valores por
	// debajo de la diagonal que sobreviven como error de redondeo.
	for i := 0; i < k; i++ {
		for c := 0; c < i && c < n; c++ {
			r[i][c] = 0
		}
	}

	reducedQ := make([][]float64, m)
	for i := 0; i < m; i++ {
		reducedQ[i] = clean(q[i][:k])
	}

	reducedR := make([][]float64, k)
	for i := 0; i < k; i++ {
		reducedR[i] = clean(r[i][:n])
	}

	return reducedQ, reducedR, nil
}

func clone(a [][]float64) [][]float64 {
	out := make([][]float64, len(a))
	for i, row := range a {
		out[i] = append([]float64(nil), row...)
	}
	return out
}

func identity(n int) [][]float64 {
	out := make([][]float64, n)
	for i := range out {
		out[i] = make([]float64, n)
		out[i][i] = 1
	}
	return out
}

func clean(row []float64) []float64 {
	out := make([]float64, len(row))
	for i, value := range row {
		if math.Abs(value) < epsilon {
			value = 0
		}
		out[i] = value
	}
	return out
}
