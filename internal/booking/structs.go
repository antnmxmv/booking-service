package booking

import (
	"fmt"
	"time"

	"github.com/antnmxmv/booking-service/internal/payment"
)

type ReservationStatus string

const (
	CreatedReservationStatus  ReservationStatus = "created"
	CanceledReservationStatus ReservationStatus = "canceled"
	FinishedReservationStatus ReservationStatus = "finished"
)

type Reservation struct {
	ID                    string               `json:"id"`
	UserID                string               `json:"user_id"`
	HotelID               string               `json:"hotel_id"`
	RoomTypes             RoomsRequest         `json:"rooms"`
	PaymentType           payment.SourceType   `json:"payment_type"`
	PaymentRequestDetails payment.OrderDetails `json:"payment_request,omitempty"`
	PaymentOrder          payment.Order        `json:"payment_order,omitempty"`
	StartDate             time.Time            `json:"start_date"`
	EndDate               time.Time            `json:"end_date"`
	Cost                  int                  `json:"cost"`
	Status                ReservationStatus    `json:"status"`
	AppliedDiscountIDs    []string             `json:"applied_discount_ids"`
	LastUpdateTime        time.Time            `json:"last_update"`
}

type RoomAvailability struct {
	Date      time.Time `json:"date"`
	Type      string    `json:"type"`
	FreeCount uint      `json:"free_count"`
}

type ReservationRequest struct {
	ID             string
	UserID         string
	HotelID        string
	RoomsRequest   RoomsRequest
	PaymentType    payment.SourceType
	PaymentDetails payment.OrderDetails
	StartDate      time.Time
	EndDate        time.Time
}

type RoomRequest struct {
	RoomType string `json:"type"`
	Count    uint   `json:"count"`
}

type RoomsRequest []RoomRequest

type wrappedError struct {
	err  error
	text string
}

func (e wrappedError) Unwrap() error {
	return e.err
}

func (e wrappedError) Error() string {
	return fmt.Sprintf("%s: %s", e.err.Error(), e.text)
}

func WrapError(err error, text string) error {
	return wrappedError{err: err, text: text}
}
