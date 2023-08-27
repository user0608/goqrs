package collection

import (
	"goqrs/internal/dominio/services"
	"goqrs/security"

	"github.com/ksaucedo002/answer"
	"github.com/labstack/echo/v4"
)

func HandleFindAll(s services.CollectionService) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := security.UserName(c.Request().Context())
		collections, err := s.GetAll(c.Request().Context(), username)
		if err != nil {
			return answer.ErrorResponse(c, err)
		}
		return answer.OK(c, collections)
	}
}

func HandleFindByID(s services.CollectionService) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := security.UserName(c.Request().Context())
		id := c.Param("collection_id")
		collection, err := s.GetByID(c.Request().Context(), username, id)
		if err != nil {
			return answer.ErrorResponse(c, err)
		}
		return answer.OK(c, collection)
	}
}

func HandleFindCollectionTags(s services.CollectionService) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("collection_id")
		tags, err := s.GetTags(c.Request().Context(), id)
		if err != nil {
			return answer.ErrorResponse(c, err)
		}
		return answer.OK(c, tags)
	}
}
