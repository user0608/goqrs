package tickets

import (
	"goqrs/services"

	"github.com/ksaucedo002/answer"
	"github.com/labstack/echo/v4"
)

func HandleCunsultTickt(service services.TicketService) echo.HandlerFunc {
	return func(c echo.Context) error {
		uuid := c.Param("ticket_uuid")
		ticket, err := service.ConsultTicketByID(c.Request().Context(), uuid)
		if err != nil {
			return answer.ErrorResponse(c, err)
		}
		return answer.OK(c, ticket)
	}
}

func HandleClaimTicket(service services.TicketService) echo.HandlerFunc {
	return func(c echo.Context) error {
		uuid := c.Param("ticket_uuid")
		if err := service.ClaimTicket(c.Request().Context(), uuid); err != nil {
			return answer.ErrorResponse(c, err)
		}
		return answer.Message(c, answer.SUCCESS_OPERATION)
	}
}

func HandleInvalidTicket(service services.TicketService) echo.HandlerFunc {
	return func(c echo.Context) error {
		uuid := c.Param("ticket_uuid")
		if err := service.InvalidTicket(c.Request().Context(), uuid); err != nil {
			return answer.ErrorResponse(c, err)
		}
		return answer.Message(c, answer.SUCCESS_OPERATION)
	}
}
