package inmemory

import (
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/antnmxmv/booking-service/internal/booking"
)

type Storage struct {
	reservations     map[string]*booking.Reservation
	roomAvailability []*RoomAvailability
	mux              sync.RWMutex
}

func NewStorage() *Storage {
	return &Storage{
		reservations:     map[string]*booking.Reservation{},
		roomAvailability: []*RoomAvailability{},
	}
}

func (s *Storage) getStartIndex(date time.Time) int {
	return sort.Search(len(s.roomAvailability), func(i int) bool {
		return !s.roomAvailability[i].Date.Before(date)
	})
}

func (s *Storage) WithRoomAvailability(roomAvailability []*RoomAvailability) *Storage {
	s.roomAvailability = make([]*RoomAvailability, len(roomAvailability))
	copy(s.roomAvailability, roomAvailability)
	sort.Slice(s.roomAvailability, func(i, j int) bool {
		return s.roomAvailability[i].Date.Before(s.roomAvailability[j].Date)
	})
	return s
}

func (s *Storage) WithReservations(reservations []*booking.Reservation) *Storage {
	for _, r := range reservations {
		s.reservations[r.ID] = r
	}
	return s
}

func (s *Storage) Repository() booking.Repository {
	return s
}

func (s *Storage) CreateReservation(reservation *booking.ReservationRequest) (*booking.Reservation, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, ok := s.reservations[reservation.ID]; ok {
		return nil, booking.ErrDuplicate
	}

	needRoomTypes := make(map[string]uint, len(reservation.RoomsRequest))

	for _, roomRequest := range reservation.RoomsRequest {
		needRoomTypes[roomRequest.RoomType] += roomRequest.Count
	}

	daysToBook := int(reservation.EndDate.Sub(reservation.StartDate).Hours()/24) + 1
	totalRoomDays := len(reservation.RoomsRequest) * daysToBook

	startIndex := s.getStartIndex(reservation.StartDate)

	for i := startIndex; i < len(s.roomAvailability) && !s.roomAvailability[i].Date.After(reservation.EndDate); i++ {
		if s.roomAvailability[i].HotelID != reservation.HotelID {
			continue
		}
		if needRoomTypes[s.roomAvailability[i].RoomType] > s.roomAvailability[i].Quota {
			return nil, booking.WrapError(booking.ErrAlreadyBooked, s.roomAvailability[i].Date.String())
		}
		totalRoomDays--
	}

	if totalRoomDays > 0 {
		return nil, booking.ErrNotWorkingDays
	}

	for i := startIndex; i < len(s.roomAvailability) && !s.roomAvailability[i].Date.After(reservation.EndDate); i++ {
		if !s.roomAvailability[i].Date.After(reservation.EndDate) &&
			s.roomAvailability[i].HotelID == reservation.HotelID &&
			needRoomTypes[s.roomAvailability[i].RoomType] > 0 {
			s.roomAvailability[i].Quota -= needRoomTypes[s.roomAvailability[i].RoomType]
		}
	}

	result := &booking.Reservation{
		ID:                    reservation.ID,
		HotelID:               reservation.HotelID,
		RoomTypes:             reservation.RoomsRequest,
		StartDate:             reservation.StartDate,
		PaymentType:           reservation.PaymentType,
		PaymentRequestDetails: reservation.PaymentDetails,
		EndDate:               reservation.EndDate,
		Status:                booking.CreatedReservationStatus,
		LastUpdateTime:        time.Now(),
	}

	s.reservations[result.ID] = result

	return result, nil
}

func (s *Storage) CancelReservation(reservationID string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	reservation := s.reservations[reservationID]

	reservation.Status = booking.CanceledReservationStatus

	startIndex := s.getStartIndex(reservation.StartDate)

	roomTypeCount := make(map[string]uint, len(reservation.RoomTypes))
	for _, roomType := range reservation.RoomTypes {
		roomTypeCount[roomType.RoomType] = uint(roomType.Count)
	}

	for i := startIndex; i < len(s.roomAvailability) && !s.roomAvailability[i].Date.After(reservation.EndDate); i++ {
		if s.roomAvailability[i].HotelID != reservation.HotelID {
			continue
		}
		if count, ok := roomTypeCount[s.roomAvailability[i].RoomType]; ok {
			s.roomAvailability[i].Quota += count
		}
	}

	return nil
}

func (s *Storage) UpdateReservation(update *booking.Reservation) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	r, ok := s.reservations[update.ID]
	if !ok {
		return errors.New("not found")
	}
	for i, rt := range update.RoomTypes {
		if r.RoomTypes[i] != rt {
			return errors.New("room types can not be updated. please create new reservation")
		}
	}

	s.reservations[update.ID] = update

	return nil
}

func (s *Storage) GetRoomsByDates(hotelID string, startDate, endDate time.Time) ([]*booking.RoomAvailability, error) {
	res := make([]*booking.RoomAvailability, 0, int(endDate.Sub(startDate).Hours())/24)
	s.mux.RLock()
	defer s.mux.RUnlock()

	for i := len(s.roomAvailability) - 1; i >= 0; i-- {
		if s.roomAvailability[i].Date.Before(startDate) {
			break
		}
		if !s.roomAvailability[i].Date.After(endDate) &&
			s.roomAvailability[i].HotelID == hotelID &&
			s.roomAvailability[i].Quota > 0 {
			res = append(res, &booking.RoomAvailability{
				Date:      s.roomAvailability[i].Date,
				Type:      s.roomAvailability[i].RoomType,
				FreeCount: s.roomAvailability[i].Quota,
			})
		}
	}

	return res, nil
}

func (s *Storage) GetReservationByID(id string) (*booking.Reservation, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	res, ok := s.reservations[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return res, nil
}

func (s *Storage) GetReservationsByUserID(userID string) ([]*booking.Reservation, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	res := []*booking.Reservation{}

	for _, reservation := range s.reservations {
		if reservation.UserID == userID {
			res = append(res, reservation)
		}
	}

	return res, nil
}

func (s *Storage) GetNotFinishedReservations() ([]*booking.Reservation, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()
	res := []*booking.Reservation{}

	for _, reservation := range s.reservations {
		if reservation.Status != booking.CanceledReservationStatus &&
			reservation.Status != booking.FinishedReservationStatus {
			res = append(res, reservation)
		}
	}

	return res, nil
}
