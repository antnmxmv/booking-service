package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/antnmxmv/booking-service/internal/booking"
	"github.com/antnmxmv/booking-service/internal/payment"
)

type reservationHandler struct {
	s *booking.BookingService
	p *payment.Provider
}

func NewCreateReservationHandler(bookingService *booking.BookingService, paymentObserver *payment.Provider) http.HandlerFunc {
	return (&reservationHandler{s: bookingService, p: paymentObserver}).handlerFn
}

func (h *reservationHandler) handlerFn(w http.ResponseWriter, r *http.Request) {
	setHeaders(w)

	req, err := h.parseReservationRequest(r)

	if err != nil {
		httpError := err.(httpError)
		errorJSON(w, httpError.text, httpError.code)
		return
	}

	res, err := h.s.CreateReservation(r.Header.Get("user_id"), req)

	if err != nil {
		if errors.Is(err, booking.ErrAlreadyBooked) || errors.Is(err, booking.ErrDuplicate) {
			errorJSON(w, err.Error(), http.StatusConflict)
		} else if errors.Is(err, booking.ErrNotWorkingDays) ||
			errors.Is(err, booking.ErrNotListedPaymentType) {
			errorJSON(w, err.Error(), http.StatusBadRequest)
		} else {
			errorJSON(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		return
	}

	outputBytes, err := json.Marshal(res)

	if err != nil {
		errorJSON(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Write(outputBytes)
}

// createReservationRequest object for manual decoding
type createReservationRequest struct {
	ID             string          `json:"id"`
	HotelID        string          `json:"hotel_id"`
	RoomsRequest   []roomsRequest  `json:"rooms"`
	PaymentType    string          `json:"payment_type"`
	PaymentDetails json.RawMessage `json:"payment_details"`
	StartDate      string          `json:"start_date"`
	EndDate        string          `json:"end_date"`
}

type roomsRequest struct {
	RoomType string `json:"type"`
	Count    uint   `json:"count"`
}

func (h *reservationHandler) parseReservationRequest(r *http.Request) (*booking.ReservationRequest, error) {
	req := createReservationRequest{}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, httpError{code: http.StatusInternalServerError, text: fmt.Sprintf("reading request: %s", err.Error())}
	}
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		return nil, httpError{code: http.StatusInternalServerError, text: fmt.Sprintf("unmarshal json: %s", err.Error())}
	}

	res, err := h.requestToModel(req, r.Header.Get("user_id"))

	if err != nil {
		return nil, err
	}

	return res, h.validate(res)
}

// requestToModel compress rooms request, conditional unmarshaling and manual time format
func (h *reservationHandler) requestToModel(r createReservationRequest, userID string) (*booking.ReservationRequest, error) {
	if r.HotelID == "" {
		return nil, httpError{code: http.StatusBadRequest, text: "hotel_id in wrong format"}
	}

	res := booking.ReservationRequest{
		ID:      r.ID,
		HotelID: r.HotelID,
		UserID:  userID,
	}

	var err error
	if res.StartDate, err = time.Parse(TimeLayout, r.StartDate); err != nil {
		return nil, httpError{code: http.StatusBadRequest, text: "start_date in wrong format"}
	}
	if res.EndDate, err = time.Parse(TimeLayout, r.EndDate); err != nil {
		return nil, httpError{code: http.StatusBadRequest, text: "end_date in wrong format"}
	}

	res.StartDate, res.EndDate = toDay(res.StartDate), toDay(res.EndDate)

	res.RoomsRequest = make(booking.RoomsRequest, 0, len(r.RoomsRequest))

	roomTypesSet := map[string]bool{}

	for _, r := range r.RoomsRequest {
		if roomTypesSet[r.RoomType] {
			for i := 0; i < len(res.RoomsRequest); i++ {
				if res.RoomsRequest[i].RoomType == r.RoomType {
					res.RoomsRequest[i].Count += r.Count
				}
			}
		} else {
			res.RoomsRequest = append(res.RoomsRequest, booking.RoomRequest{
				RoomType: r.RoomType,
				Count:    r.Count,
			})
			roomTypesSet[r.RoomType] = true
		}
	}

	res.PaymentType = payment.SourceType(r.PaymentType)

	res.PaymentDetails, err = h.p.UnmarshalDetailsJSON(res.PaymentType, r.PaymentDetails)
	if err != nil {
		err = httpError{code: http.StatusBadRequest, text: err.Error()}
	}

	return &res, err
}

var (
	paymentTypeError   = httpError{code: http.StatusBadRequest, text: "wrong payment type"}
	roomTypeEmptyError = httpError{code: http.StatusBadRequest, text: "room type cant be empty"}
	roomsCountError    = httpError{code: http.StatusBadRequest, text: "rooms count cant be less than 1"}
	datesOrderError    = httpError{code: http.StatusBadRequest, text: "start_date is after end_date"}
	wrongDatesError    = httpError{code: http.StatusBadRequest, text: "dates must be not before today"}
)

// validate checks
func (h *reservationHandler) validate(req *booking.ReservationRequest) error {
	if _, ok := h.p.GetSources()[req.PaymentType]; !ok {
		return paymentTypeError
	}

	if len(req.RoomsRequest) == 0 {
		return roomsCountError
	}

	for i := 0; i < len(req.RoomsRequest); i++ {
		if req.RoomsRequest[i].RoomType == "" {
			return roomTypeEmptyError
		}
		if req.RoomsRequest[i].Count == 0 {
			return roomsCountError
		}
	}

	if req.EndDate.Before(req.StartDate) {
		return datesOrderError
	}
	today := toDay(time.Now())

	if req.StartDate.Before(today) || req.EndDate.Before(today) {
		return wrongDatesError
	}

	return nil
}
