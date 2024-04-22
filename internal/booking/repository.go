package booking

import (
	"errors"
	"time"
)

var (
	ErrAlreadyBooked        = errors.New("rooms are not available at some date")
	ErrDuplicate            = errors.New("reservation with this id already exists")
	ErrNotWorkingDays       = errors.New("some dates in request are not listed as available to book")
	ErrNotListedPaymentType = errors.New("payment type is not allowed by hotel")
)

type Repository interface {
	CreateReservation(reservation *ReservationRequest) (*Reservation, error)

	CancelReservation(id string) error

	UpdateReservation(reservation *Reservation) error

	GetRoomsByDates(hotelID string, startDate, endDate time.Time) ([]*RoomAvailability, error)

	GetNotFinishedReservations() ([]*Reservation, error)

	GetReservationByID(id string) (*Reservation, error)

	GetReservationsByUserID(userID string) ([]*Reservation, error)
}
