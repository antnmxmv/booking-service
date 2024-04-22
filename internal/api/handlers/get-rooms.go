package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/antnmxmv/booking-service/internal/booking"
	"github.com/go-chi/chi"
)

type getRoomsHandler struct {
	s *booking.BookingService
}

func NewGetRoomsHandler(bookingService *booking.BookingService) http.HandlerFunc {
	return (&getRoomsHandler{s: bookingService}).handlerFn
}

func (h *getRoomsHandler) handlerFn(w http.ResponseWriter, r *http.Request) {
	setHeaders(w)

	hotelID := chi.URLParam(r, "hotelID")
	if hotelID == "" {
		errorJSON(w, "bad request", http.StatusBadRequest)
		return
	}

	res, err := h.s.GetAvailableRoomTypes(hotelID)
	if err != nil {
		errorJSON(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	outputBytes, err := json.Marshal(res)

	if err != nil {
		errorJSON(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Write(outputBytes)
}
