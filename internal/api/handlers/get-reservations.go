package handlers

import (
	"net/http"

	"github.com/antnmxmv/booking-service/internal/booking"
	"github.com/gin-gonic/gin"
)

type getReservationsHandler struct {
	s *booking.BookingService
}

func NewGetReservationsHandler(bookingService *booking.BookingService) gin.HandlerFunc {
	return (&getReservationsHandler{s: bookingService}).handlerFn
}

func (h *getReservationsHandler) handlerFn(ctx *gin.Context) {
	setHeaders(ctx)

	reservations, err := h.s.GetUserReservations(ctx.GetHeader("user_id"))

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorJSON(http.StatusText(http.StatusInternalServerError)))
		return
	}

	res := make([]reservationResponse, len(reservations))

	for i := range reservations {
		res[i] = reservationModelToResponse(reservations[i])
	}

	ctx.JSON(http.StatusOK, res)

	return
}
