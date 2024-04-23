package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/antnmxmv/booking-service/internal/booking"
	"github.com/antnmxmv/booking-service/internal/payment"
	"github.com/gin-gonic/gin"
)

type reservationHandler struct {
	s *booking.BookingService
	p *payment.Provider
}

func NewCreateReservationHandler(bookingService *booking.BookingService, paymentObserver *payment.Provider) gin.HandlerFunc {
	return (&reservationHandler{s: bookingService, p: paymentObserver}).handlerFn
}

func (h *reservationHandler) handlerFn(ctx *gin.Context) {
	setHeaders(ctx)

	userID := ctx.GetHeader("user_id")

	req := createReservationRequest{}
	_ = ctx.ShouldBindJSON(&req)

	reservationRequest, err := h.requestToModel(req, userID)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorJSON(err.Error()))
		return
	}

	err = h.validate(reservationRequest)

	if err != nil {
		httpError := err.(httpError)
		ctx.JSON(httpError.code, errorJSON(httpError.text))
		return
	}

	res, err := h.s.CreateReservation(userID, reservationRequest)

	if err != nil {
		if errors.Is(err, booking.ErrAlreadyBooked) || errors.Is(err, booking.ErrDuplicate) {
			ctx.JSON(http.StatusConflict, errorJSON(err.Error()))
		} else if errors.Is(err, booking.ErrNotWorkingDays) ||
			errors.Is(err, booking.ErrNotListedPaymentType) {
			ctx.JSON(http.StatusBadRequest, errorJSON(err.Error()))
		} else {
			ctx.JSON(http.StatusInternalServerError, errorJSON(http.StatusText(http.StatusInternalServerError)))
		}

		return
	}

	ctx.JSON(http.StatusOK, reservationModelToResponse(res))
}

// createReservationRequest object for manual decoding
type createReservationRequest struct {
	ID             string          `json:"id"`
	HotelID        string          `json:"hotel_id"`
	RoomsRequest   []roomsRequest  `json:"rooms"`
	PaymentType    string          `json:"payment_type"`
	PaymentDetails json.RawMessage `json:"payment_details"`
	StartDate      *TimeJSON       `json:"start_date"`
	EndDate        *TimeJSON       `json:"end_date"`
}

type roomsRequest struct {
	RoomType string `json:"type"`
	Count    uint   `json:"count"`
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

	res.StartDate, res.EndDate = toDay(r.StartDate.Time), toDay(r.EndDate.Time)

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
