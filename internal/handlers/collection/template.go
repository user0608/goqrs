package collection

import (
	"errors"
	"fmt"

	"goqrs/internal/dominio/models"
	"goqrs/internal/dominio/services"
	"goqrs/security"
	"goqrs/xstorage"
	"net/http"
	"os"
	"path"

	"github.com/ksaucedo002/answer"
	"github.com/ksaucedo002/answer/errores"
	"github.com/labstack/echo/v4"
)

func HandleUploadTempleate(s services.CollectionService) echo.HandlerFunc {
	var binder = &echo.DefaultBinder{}
	return func(c echo.Context) error {
		collectionid := c.Param("collection_id")
		var stts models.TemlateDetails
		if err := binder.BindBody(c, &stts); err != nil {
			return answer.ErrorResponse(c,
				errores.NewBadRequestf(err, "los valores del formulario son incorrectos"),
			)
		}
		ff, err := c.FormFile("template")
		if err != nil {
			return answer.ErrorResponse(c,
				errores.NewBadRequestf(nil, "no se encontro el template"),
			)
		}
		if ff.Header.Get("Content-Type") == "image/jpeg" {
			ff.Filename = "img.jpg"
		}
		if path.Ext(ff.Filename) != ".jpg" {
			return answer.ErrorResponse(c,
				errores.NewBadRequestf(nil, "solo hay soporte para archivos .jpg"),
			)
		}
		f, err := ff.Open()
		if err != nil {
			return answer.ErrorResponse(c,
				errores.NewBadRequestf(nil, "no se pudo abrir el template"),
			)
		}
		defer f.Close()
		templateuuid, err := s.SaveTemplate(c.Request().Context(), collectionid, &stts, f)
		if err != nil {
			return answer.ErrorResponse(c, err)
		}
		return answer.OK(c, echo.Map{"template_uuid": templateuuid})
	}
}
func ImageTemplate(s xstorage.StorageService) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := security.UserName(c.Request().Context())
		id := c.Param("collection_id")
		img, err := s.Find(fmt.Sprintf("%s/%s.jpg", username, id))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return answer.ErrorResponse(
					c,
					errores.NewNotFoundf(nil, "no se encontro el recurso"),
				)
			}
			return answer.ErrorResponse(c, err)
		}
		return c.Stream(http.StatusOK, "image/jpeg", img)
	}
}
