package collection

import (
	"goqrs/internal/dominio/services"

	"github.com/ksaucedo002/answer"
	"github.com/labstack/echo/v4"
)

func HandleDeleteCollection(service services.CollectionService) echo.HandlerFunc {
	return func(c echo.Context) error {
		var id = c.Param("collection_id")
		if err := service.Delete(c.Request().Context(), id); err != nil {
			return answer.ErrorResponse(c, err)
		}
		return answer.Message(c, answer.SUCCESS_OPERATION)
	}
}
