package collection

import (
	"goqrs/models"
	"goqrs/security"
	"goqrs/services"
	"goqrs/utils"

	"github.com/ksaucedo002/answer"
	"github.com/labstack/echo/v4"
)

func HandelCreate(s services.CollectionService) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := security.UserName(c.Request().Context())
		var collection = models.Collection{
			AccountUsername: username,
		}
		if err := utils.JSON(c, &collection); err != nil {
			return answer.ErrorResponse(c, err)
		}
		if err := s.Create(c.Request().Context(), &collection); err != nil {
			return answer.ErrorResponse(c, err)
		}
		return answer.Message(c, answer.SUCCESS_CREATE)
	}
}
func HandelUpdate(s services.CollectionService) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := security.UserName(c.Request().Context())
		var collection models.Collection
		if err := utils.JSON(c, &collection); err != nil {
			return answer.ErrorResponse(c, err)
		}
		if err := s.Update(c.Request().Context(), &collection, username); err != nil {
			return answer.ErrorResponse(c, err)
		}
		return answer.Message(c, answer.SUCCESS_OPERATION)
	}
}
