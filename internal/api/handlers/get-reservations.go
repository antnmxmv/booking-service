package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/antnmxmv/booking-service/internal/booking"
	"github.com/antnmxmv/booking-service/internal/payment"
)

type getReservationsHandler struct {
	s *booking.BookingService
}

func NewGetReservationsHandler(bookingService *booking.BookingService) http.HandlerFunc {
	return (&getReservationsHandler{s: bookingService}).handlerFn
}

type reservationResponse struct {
	ID                    string               `json:"id"`
	HotelID               string               `json:"hotel_id"`
	RoomTypes             booking.RoomsRequest `json:"rooms"`
	PaymentType           payment.SourceType   `json:"payment_type"`
	PaymentOrder          payment.Order        `json:"payment_order"`
	PaymentRequestDetails payment.OrderDetails `json:"payment_request"`
	StartDate             time.Time            `json:"start_date"`
	EndDate               time.Time            `json:"end_date"`
	Cost                  int                  `json:"cost"`
	Status                string               `json:"status"`
	AppliedDiscountIDs    []string             `json:"applied_discount_ids"`
	LastUpdateTime        time.Time            `json:"last_update"`
}

func (h *getReservationsHandler) handlerFn(w http.ResponseWriter, r *http.Request) {
	setHeaders(w)

	reservations, err := h.s.GetUserReservations(r.Header.Get("user_id"))
	if err != nil {
		errorJSON(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	res := make([]reservationResponse, len(reservations))

	for i, r := range reservations {
		res[i] = reservationResponse{
			ID:                    r.ID,
			HotelID:               r.HotelID,
			RoomTypes:             r.RoomTypes,
			PaymentType:           r.PaymentType,
			PaymentRequestDetails: r.PaymentRequestDetails,
			StartDate:             r.StartDate,
			EndDate:               r.EndDate,
			Cost:                  r.Cost,
			Status:                "",
			AppliedDiscountIDs:    []string{},
			LastUpdateTime:        time.Time{},
		}
	}

	outputBytes, err := json.Marshal(reservations)
	if err != nil {
		errorJSON(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Write(outputBytes)
}
