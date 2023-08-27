package collection

import (
	"goqrs/internal/dominio/models"
	"goqrs/internal/dominio/services"
	"goqrs/utils"

	"github.com/ksaucedo002/answer"
	"github.com/labstack/echo/v4"
)

func HandleAddTag(service services.CollectionService) echo.HandlerFunc {
	return func(c echo.Context) error {
		var tag models.Tag
		tag.CollectionID = c.Param("collection_id")
		if err := utils.JSON(c, &tag); err != nil {
			return answer.ErrorResponse(c, err)
		}
		if err := service.AddTag(c.Request().Context(), &tag); err != nil {
			return answer.ErrorResponse(c, err)
		}
		return answer.Message(c, answer.SUCCESS_CREATE)
	}
}

func HandleRemoveTag(service services.CollectionService) echo.HandlerFunc {
	return func(c echo.Context) error {
		var tagid = c.Param("tag_id")
		if err := service.RemoveTag(c.Request().Context(), tagid); err != nil {
			return answer.ErrorResponse(c, err)
		}
		return answer.Message(c, answer.SUCCESS_OPERATION)
	}
}
