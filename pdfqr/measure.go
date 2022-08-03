package pdfqr

import (
	"errors"
	"math"
)

const min_space_beetwen_x float64 = 5  //5mm
const min_space_beetwen_y float64 = 10 //5mm
const a4_width float64 = 210           //210mm
const a4_height float64 = 297          //297mm
const min_element_with float64 = 10    //10mm
var ErrElementMaxWith = errors.New("error la dimencion del objeto es mayor a la del contenedor")
var ErrElementMinWith = errors.New("error  la dimencion del objeto es menor al minimo, min=10mm")
var ErrElementSpace = errors.New("error el espacio de separación es menor al mínimo requerido")

// Segments recibe una longitud x y w, calcula la cantidad
// maxima de elementos de longitud w que pueden
// ser contenidos dentro de x.
// Retornando  las pociciones iniciales de w sobre x
// mingap = separación mínima entre segmentos w.
// x <= w  - mingap, si no error
func segments(x float64, w float64, mingap float64) ([]float64, error) {
	if w < min_element_with {
		return nil, ErrElementMinWith
	}
	n := math.Trunc((x - mingap) / (w + mingap))
	if n <= 0 {
		return nil, ErrElementMaxWith
	}
	r := math.Trunc((x - n*w) / (n + 1))
	if r < mingap {
		return nil, ErrElementMaxWith
	}
	cols := make([]float64, int(n))
	var aux = r
	for i := 0; i < int(n); i++ {
		cols[i] = aux
		aux += w + r
	}
	return cols, nil
}

// NumColumns, retorna un slice con las posiciones o condenada x
// para cada columna, el numero de elementos del slice son la
// cantidad de columnas
func xPositons(w float64) ([]float64, error) {
	return segments(a4_width, w, min_space_beetwen_x)
}
func yPositons(h float64) ([]float64, error) {
	return segments(a4_height, h, min_space_beetwen_y)
}
