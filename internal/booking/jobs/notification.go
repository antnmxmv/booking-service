package jobs

import "github.com/antnmxmv/booking-service/internal/booking"

type NotificationProviderFacade interface {
	Notify(any) error
}

// NotificationJob example of service
type NotificationJob struct {
	// api NotificationProviderFacade
	ch chan booking.JobResponse
}

func NewNotificationJob() *NotificationJob {
	ch := make(chan booking.JobResponse)
	close(ch)
	return &NotificationJob{
		ch: ch,
	}
}

func (p *NotificationJob) Name() booking.ReservationStatus {
	return "notification"
}

func (p *NotificationJob) Run(_ *booking.Reservation) (*bool, error) {
	// p.api.Notify(reservation.ID)
	res := true
	return &res, nil
}

func (p *NotificationJob) Cancel(_ *booking.Reservation) (*bool, error) {
	// already notified, can not cancel
	res := false
	return &res, nil
}

func (p *NotificationJob) Subscribe() (<-chan booking.JobResponse, error) {
	return p.ch, nil
}
