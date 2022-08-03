package pdfqr

import "testing"

func TestNumPages(t *testing.T) {
	tsc := map[string]struct {
		numitems int
		rows     int
		cols     int

		want int
	}{
		"C1":  {0, 10, 10, 0},
		"C2":  {10, 0, 0, 0},
		"C3":  {21, 3, 3, 3},
		"C4":  {1, 10, 10, 1},
		"C5":  {100, 10, 2, 5},
		"C6":  {101, 10, 2, 6},
		"C7":  {0, 0, 0, 0},
		"C8":  {7, 1, 1, 7},
		"C9":  {1, 1, 1, 1},
		"C10": {1, 100, 100, 1},
	}

	for name, tc := range tsc {
		t.Run(name, func(t *testing.T) {
			if r := numPages(tc.numitems, tc.rows, tc.cols); r != tc.want {
				t.Errorf("Se esperaba :%d, se obtuvo %d", tc.want, r)
			}
		})
	}
}
