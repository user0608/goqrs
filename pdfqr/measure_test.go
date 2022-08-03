package pdfqr

import (
	"errors"
	"testing"
)

func TestNumColumns(t *testing.T) {
	dts := map[string]struct {
		width float64
		cols  []float64
		err   error
	}{
		"T1": {210, []float64{}, ErrElementMaxWith},
		"T2": {200, []float64{5}, nil},
		"T3": {100, []float64{55}, nil},
		"T4": {90, []float64{10, 110}, nil},
		"T5": {0, []float64{}, ErrElementMinWith},
		"T6": {-1, []float64{}, ErrElementMinWith},
		"T7": {min_element_with - 1, []float64{}, ErrElementMinWith},
		"T8": {50, []float64{15, 80, 145}, nil},
	}
	for name, dt := range dts {
		t.Run(name, func(t *testing.T) {
			cols, err := xPositons(dt.width)
			if !errors.Is(err, dt.err) {
				t.Errorf("se esperaba err: %v, se obtuvo err: %v", dt.err, err)
			}
			if len(cols) != len(dt.cols) {
				t.Errorf("se esperaba len=%d, se obtuvo len=%d", len(dt.cols), len(cols))
				return
			}
			for i, value := range dt.cols {
				if value != cols[i] {
					t.Errorf("se esperaba [i=%d] %f, se obtuve: %f", i, value, cols[i])
				}
			}
		})
	}
}
