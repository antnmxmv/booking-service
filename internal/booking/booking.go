package booking

import (
	"context"
	"time"

	"github.com/antnmxmv/booking-service/internal/config"
	"github.com/antnmxmv/booking-service/internal/payment"
)

const roomsAvailabilityRequestWindow = time.Hour * 24 * 365

// BookingService provides business level api for making reservations
type BookingService struct {
	config config.Config
	repo   Repository
	// cancelationQueue is literaly delayed queue which handles reservation cancelation
	cancelationQueue DelayedQueue
	// reservationOrchestrator handles reservations lifecycle
	reservationOrchestrator *ReservationOrchestrator
	doneCh                  chan struct{}
}

func NewBookingService(
	cnf config.Config,
	repo Repository,
	paymentProvider *payment.Provider,
	reservationOrchestrator *ReservationOrchestrator,
	queue DelayedQueue,
) *BookingService {
	return &BookingService{
		config:                  cnf,
		repo:                    repo,
		cancelationQueue:        queue,
		reservationOrchestrator: reservationOrchestrator,
		doneCh:                  make(chan struct{}),
	}
}

func (s *BookingService) CreateReservation(userID string, request *ReservationRequest) (*Reservation, error) {

	// send to cancelation queue
	if err := s.cancelationQueue.SendMessage(request.ID, s.config.GetData().Booking.IdleReservationTimeout); err != nil {
		return nil, err
	}

	// make transaction in db, check if we have available dates
	reservation, err := s.repo.CreateReservation(request)
	if err != nil {
		return nil, err
	}

	// execute all jobs over this reservation very consistently
	if err := s.reservationOrchestrator.execute(reservation, false); err != nil {
		return nil, err
	}

	return reservation, nil
}

func (s *BookingService) ChangePaymentMethod(reservationID string, method payment.SourceType, details payment.OrderDetails) (*Reservation, error) {

	r, err := s.repo.GetReservationByID(reservationID)
	if err != nil {
		return r, err
	}

	err = s.reservationOrchestrator.rollback(r, false)

	if err := s.repo.UpdateReservation(r); err != nil {
		return r, err
	}

	if err != nil {
		return r, err
	}

	err = s.reservationOrchestrator.execute(r, true)

	if err != nil {
		return r, err
	}

	return r, nil
}

func (s *BookingService) GetUserReservations(userID string) ([]*Reservation, error) {
	return s.repo.GetReservationsByUserID(userID)
}

func (s *BookingService) GetAvailableRoomTypes(hotelID string) ([]*RoomAvailability, error) {
	return s.repo.GetRoomsByDates(hotelID, time.Now(), time.Now().Add(roomsAvailabilityRequestWindow))
}

func (s *BookingService) Start(ctx context.Context) error {
	go func() {
		queue := s.cancelationQueue.Subscribe()
		for {
			select {
			case reservationID := <-queue:
				reservation, err := s.repo.GetReservationByID(reservationID)
				if err != nil {
					break
				}
				if reservation.Status == FinishedReservationStatus ||
					reservation.Status == CanceledReservationStatus {
					break
				}
				// update status or requeue
				if time.Since(reservation.LastUpdateTime) > s.config.GetData().Booking.IdleReservationTimeout {
					if err := s.repo.CancelReservation(reservationID); err != nil {
						_ = s.cancelationQueue.SendMessage(reservationID, time.Minute)
					}
				} else {
					_ = s.cancelationQueue.SendMessage(reservationID, s.config.GetData().Booking.IdleReservationTimeout-time.Since(reservation.LastUpdateTime))
				}

			case <-s.doneCh:
				return
			}
		}
	}()

	return s.reservationOrchestrator.run(s.config.GetData().Booking.IdleReservationTimeout)
}

func (s *BookingService) Stop(ctx context.Context) error {
	close(s.doneCh)
	s.reservationOrchestrator.stop()
	return nil
}
