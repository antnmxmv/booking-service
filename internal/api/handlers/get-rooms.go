package handlers

import (
	"errors"
	"net/http"

	"github.com/antnmxmv/booking-service/internal/booking"
	"github.com/gin-gonic/gin"
)

type getRoomsHandler struct {
	s *booking.BookingService
}

func NewGetRoomsHandler(bookingService *booking.BookingService) gin.HandlerFunc {
	return (&getRoomsHandler{s: bookingService}).handlerFn
}

func (h *getRoomsHandler) handlerFn(ctx *gin.Context) {
	setHeaders(ctx)

	hotelID := ctx.Param("hotelID")
	if hotelID == "" {
		ctx.JSON(http.StatusBadRequest, gin.Error{Err: errors.New("bad request")})
		return
	}

	res, err := h.s.GetAvailableRoomTypes(hotelID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.Error{Err: errors.New(http.StatusText(http.StatusInternalServerError))})
		return
	}

	ctx.JSON(http.StatusOK, res)
}
