package collection

import (
	"context"
	"errors"
	"fmt"
	"goqrs/database"
	"goqrs/handlers/events"
	"goqrs/models"
	"goqrs/pdfqr"
	"goqrs/security"
	"goqrs/services"
	"goqrs/xstorage"
	"io"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/ksaucedo002/answer"
	"github.com/ksaucedo002/answer/errores"
	"github.com/labstack/echo/v4"
)

func HandlerPruebaDocument() echo.HandlerFunc {
	randomuuids := func(numitems int) []string {
		var items = make([]string, numitems)
		for i := 0; i < numitems; i++ {
			items[i] = uuid.NewString()
		}
		return items
	}
	var binder = &echo.DefaultBinder{}
	return func(c echo.Context) error {
		uuids := randomuuids(10)
		var sts models.TemlateDetails
		if err := binder.BindBody(c, &sts); err != nil {
			return answer.ErrorResponse(c,
				errores.NewBadRequestf(err, "los valores del formulario son incorrectos"),
			)
		}
		file, err := c.FormFile("template")
		if err != nil {
			return answer.ErrorResponse(c,
				errores.NewBadRequestf(nil, "no se encontro el template"),
			)
		}
		templateFile, err := file.Open()
		if err != nil {
			return answer.ErrorResponse(c,
				errores.NewBadRequestf(nil, "no se pudo abrir el template"),
			)
		}
		defer templateFile.Close()
		doc, err := createDocument(templateFile, uuids, sts)
		if err != nil {
			return answer.ErrorResponse(c, err)
		}
		return c.Stream(http.StatusOK, "application/pdf", doc)
	}
}

func createDocument(f io.ReadSeeker, codes []string, s models.TemlateDetails) (io.Reader, error) {
	if s.ItemWidth == 0 {
		return nil, errores.NewBadRequestf(nil, "item_width no puede ser 0")
	}
	if s.QqSize == 0 {
		return nil, errores.NewBadRequestf(nil, "qr_size no puede ser 0")
	}

	doc, err := pdfqr.CreateDocument(f, codes, pdfqr.DocumentConfigs{
		ItemWith: s.ItemWidth,
		QrSize:   s.QqSize,
		QrXPos:   s.QqXPos,
		QrYPos:   s.QqYPos,
	})
	if err != nil {
		if errors.Is(err, pdfqr.ErrImageType) {
			return nil, errores.NewBadRequestf(nil, err.Error())
		}
		return nil, err
	}
	return doc, nil
}

func HandleProcessDocument(publicher events.EventPublicher, service services.DocumentService) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("collection_id")
		username := security.UserName(c.Request().Context())
		claims, _ := security.JwtClaims(c.Request().Context())
		go func() {
			publicher.Publish(username, events.EventData{
				EventName: events.DOCUMENT_PROCESSING,
				Data:      map[string]string{"collection_id": id},
			})
			ctx := security.Context(context.Background(), claims)
			ctx = database.Context(ctx)
			docuuid, err := service.GenerateDocument(ctx, id)
			if err != nil {
				publicher.Publish(username, events.EventData{
					EventName: events.DOCUMENT_PROCESSED,
					Data:      map[string]string{"collection_id": id, "error": err.Error()},
				})
			}
			publicher.Publish(username, events.EventData{
				EventName: events.DOCUMENT_PROCESSED,
				Data:      map[string]string{"collection_id": id, "doc_uuid": docuuid, "error": ""},
			})
		}()
		return answer.Message(c, answer.SUCCESS_OPERATION)
	}
}

func DownloadDocumentPDF(s xstorage.StorageService) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := security.UserName(c.Request().Context())
		id := c.Param("document_uuid")
		document, err := s.Find(fmt.Sprintf("%s/%s.pdf", username, id))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return answer.ErrorResponse(
					c,
					errores.NewNotFoundf(nil, "no se encontro el recurso"),
				)
			}
			return answer.ErrorResponse(c, err)
		}
		doc, err := io.ReadAll(document)
		if err != nil {
			return answer.ErrorResponse(c, err)
		}
		c.Response().Writer.Header().Set("Content-Length", fmt.Sprint(len(doc)))
		return c.Blob(http.StatusOK, "application/pdf", doc)
	}
}
