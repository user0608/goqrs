package pdfqr

import (
	"bytes"
	"errors"
	"fmt"
	"image/jpeg"
	"io"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/signintech/gopdf"
)

type DocumentConfigs struct {
	ItemWith int
	QrSize   int
	QrXPos   int
	QrYPos   int
}

var ErrImageType = errors.New("tipo de imagen incorrecto, solo se acepta jpeg")

const default_qr_with = 74

func CreateDocument(templateImage io.ReadSeeker, codes []string, conf DocumentConfigs) (io.Reader, error) {
	img, err := jpeg.Decode(templateImage)
	if err != nil {
		return nil, ErrImageType
	}
	templateImage.Seek(0, io.SeekStart)
	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	itemhight := rebaseMesure(width, conf.ItemWith, height)
	rows, err := xPositons(float64(conf.ItemWith))
	if err != nil {
		return nil, err
	}
	cols, err := yPositons(float64(itemhight))
	if err != nil {
		return nil, err
	}
	pages := numPages(len(codes), len(rows), len(cols))
	if pages == 0 {
		return nil, errors.New("medidas inválidas, número de páginas = 0")
	}
	templateHolder, err := gopdf.ImageHolderByReader(templateImage)
	if err != nil {
		return nil, fmt.Errorf("err creando template image holder: %w", err)
	}
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{
		Unit:     gopdf.UnitMM,
		PageSize: *gopdf.PageSizeA4,
	})
	var counter = 0
LOOP:
	for i := 0; i < pages; i++ {
		pdf.AddPage()
		for _, y := range cols {
			for _, x := range rows {
				rect := &gopdf.Rect{W: float64(conf.ItemWith), H: float64(itemhight)}
				if err := pdf.ImageByHolder(templateHolder, x, y, rect); err != nil {
					return nil, fmt.Errorf("err adding image holder to document %d: %w", counter, err)
				}
				qrSize := rebaseMesure(width, conf.ItemWith, conf.QrSize)
				qrRect := &gopdf.Rect{W: qrSize, H: qrSize}
				qrImage, err := rqimg(codes[counter])
				if err != nil {
					return nil, err
				}
				qrHolder, err := gopdf.ImageHolderByReader(qrImage)
				if err != nil {
					return nil, fmt.Errorf("err creating qr holder %d: %w", counter, err)
				}
				qrX := x + rebaseMesure(width, conf.ItemWith, conf.QrXPos)
				qrY := y + rebaseMesure(width, conf.ItemWith, conf.QrYPos)
				if err := pdf.ImageByHolder(qrHolder, qrX, qrY, qrRect); err != nil {
					return nil, err
				}
				counter++
				if counter == len(codes) {
					break LOOP
				}
			}
		}
	}
	var doc bytes.Buffer
	if err := pdf.Write(&doc); err != nil {
		return nil, err
	}
	return &doc, nil
}
func rqimg(content string) (io.Reader, error) {
	qrcode, err := qr.Encode(content, qr.Q, qr.Auto)
	if err != nil {
		return nil, fmt.Errorf("err encoding qr: %w", err)
	}
	qrcode, err = barcode.Scale(qrcode, default_qr_with, default_qr_with)
	if err != nil {
		return nil, fmt.Errorf("err scaling qr: %w", err)
	}
	var imgbuffer bytes.Buffer
	if err = jpeg.Encode(&imgbuffer, qrcode, &jpeg.Options{Quality: 100}); err != nil {
		return nil, fmt.Errorf("err jpg encode qr: %w", err)
	}
	return &imgbuffer, nil
}

func numPages(numitems, rows, cols int) int {
	if rows == 0 || cols == 0 {
		return 0
	}
	var itemsForPage = rows * cols
	return (numitems + itemsForPage - 1) / itemsForPage
}
func rebaseMesure(ref, targetRef, measure int) float64 {
	return (float64(targetRef) * float64(measure)) / float64(ref)
}
