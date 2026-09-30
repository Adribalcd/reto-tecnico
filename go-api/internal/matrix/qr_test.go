package matrix

import (
	"math"
	"testing"
)

func TestQRDecomposeReconstructsInput(t *testing.T) {
	cases := map[string][][]float64{
		"cuadrada":         {{12, -51, 4}, {6, 167, -68}, {-4, 24, -41}},
		"alta":             {{1, 2}, {3, 4}, {5, 6}},
		"ancha":            {{1, 2, 3}, {4, 5, 6}},
		"rango deficiente": {{0, 1}, {0, 2}, {0, 3}},
		"una fila":         {{3, -4, 5}},
		"una columna":      {{3}, {-4}, {5}},
		"identidad":        {{1, 0}, {0, 1}},
	}

	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			q, r, err := QRDecompose(input)
			if err != nil {
				t.Fatalf("QRDecompose devolvió un error: %v", err)
			}

			assertOrthonormalColumns(t, q)
			assertUpperTriangular(t, r)
			assertReconstructs(t, input, q, r)
		})
	}
}

func TestQRDecomposeShapes(t *testing.T) {
	q, r, err := QRDecompose([][]float64{{1, 2, 3}, {4, 5, 6}})
	if err != nil {
		t.Fatalf("QRDecompose devolvió un error: %v", err)
	}

	if len(q) != 2 || len(q[0]) != 2 {
		t.Fatalf("Q debe ser 2x2 para una entrada 2x3, se obtuvo %dx%d", len(q), len(q[0]))
	}
	if len(r) != 2 || len(r[0]) != 3 {
		t.Fatalf("R debe ser 2x3 para una entrada 2x3, se obtuvo %dx%d", len(r), len(r[0]))
	}
}

func TestQRDecomposeZeroMatrix(t *testing.T) {
	q, r, err := QRDecompose([][]float64{{0, 0}, {0, 0}})
	if err != nil {
		t.Fatalf("QRDecompose devolvió un error: %v", err)
	}

	if !isIdentity(q) {
		t.Fatalf("Q debe quedar como la identidad para una matriz nula, se obtuvo %v", q)
	}
	if !isZero(r) {
		t.Fatalf("R debe quedar en cero para una matriz nula, se obtuvo %v", r)
	}
}

func TestDimensionsRejectsMalformedMatrices(t *testing.T) {
	cases := map[string][][]float64{
		"vacía":              {},
		"fila vacía":         {{}},
		"irregular":          {{1, 2}, {3}},
		"valores no finitos": {{1, math.Inf(1)}},
		"no es un número":    {{math.NaN()}},
	}

	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			if _, _, err := Dimensions(input); err == nil {
				t.Fatal("se esperaba que Dimensions rechazara la entrada")
			}
		})
	}
}

func assertOrthonormalColumns(t *testing.T, q [][]float64) {
	t.Helper()

	cols := len(q[0])
	for i := 0; i < cols; i++ {
		for j := 0; j < cols; j++ {
			dot := 0.0
			for row := range q {
				dot += q[row][i] * q[row][j]
			}
			want := 0.0
			if i == j {
				want = 1.0
			}
			if math.Abs(dot-want) > 1e-9 {
				t.Fatalf("las columnas %d y %d de Q tienen producto punto %v, se esperaba %v", i, j, dot, want)
			}
		}
	}
}

func assertUpperTriangular(t *testing.T, r [][]float64) {
	t.Helper()

	for i := range r {
		for j := 0; j < i && j < len(r[i]); j++ {
			if r[i][j] != 0 {
				t.Fatalf("el valor R[%d][%d] = %v, se esperaba 0", i, j, r[i][j])
			}
		}
	}
}

func assertReconstructs(t *testing.T, a, q, r [][]float64) {
	t.Helper()

	product := multiply(q, r)
	for i := range a {
		for j := range a[i] {
			if math.Abs(product[i][j]-a[i][j]) > 1e-9 {
				t.Fatalf("Q*R[%d][%d] = %v, se esperaba %v", i, j, product[i][j], a[i][j])
			}
		}
	}
}

func multiply(a, b [][]float64) [][]float64 {
	out := make([][]float64, len(a))
	for i := range a {
		out[i] = make([]float64, len(b[0]))
		for j := range b[0] {
			sum := 0.0
			for k := range b {
				sum += a[i][k] * b[k][j]
			}
			out[i][j] = sum
		}
	}
	return out
}

func isIdentity(a [][]float64) bool {
	for i := range a {
		for j := range a[i] {
			want := 0.0
			if i == j {
				want = 1.0
			}
			if math.Abs(a[i][j]-want) > 1e-12 {
				return false
			}
		}
	}
	return true
}

func isZero(a [][]float64) bool {
	for i := range a {
		for j := range a[i] {
			if a[i][j] != 0 {
				return false
			}
		}
	}
	return true
}
