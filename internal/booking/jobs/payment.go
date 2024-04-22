package jobs

import (
	"context"

	"github.com/antnmxmv/booking-service/internal/booking"
	"github.com/antnmxmv/booking-service/internal/payment"
)

// PaymentJob garantees that payment order status changes will affect the reservation state
type PaymentJob struct {
	p        *payment.Provider
	repo     booking.Repository
	updateCh chan booking.JobResponse
	doneCh   chan struct{}
}

func NewPaymentJob(p *payment.Provider, repo booking.Repository) *PaymentJob {
	return &PaymentJob{
		p:        p,
		repo:     repo,
		updateCh: make(chan booking.JobResponse),
		doneCh:   make(chan struct{}),
	}
}

func (p *PaymentJob) Name() booking.ReservationStatus {
	return "payment"
}

func (p *PaymentJob) Run(req *booking.Reservation) (*bool, error) {
	var isSucceeded *bool

	paymentOrder, err := p.p.CreateOrder(req.ID, req.Cost, req.PaymentType, req.PaymentRequestDetails)
	if err == nil {
		req.PaymentOrder = paymentOrder

		if paymentOrder.Status() == payment.PaymentStatusPending {
			// payment order in 'created' status means that we are waiting for users action
			// return nil
			return nil, nil
		}
		// if order immediately succeeded, we tell orchestrator that payment job is done
		boolValue := paymentOrder.Status() == payment.PaymentStatusSuccess
		isSucceeded = &boolValue
		// if order creation failed provide details of unsuccessful operation to orchestrator
		// assuming that he could change payment type later
	}

	return isSucceeded, err
}

func (p *PaymentJob) Cancel(req *booking.Reservation) (*bool, error) {
	var done *bool
	order, err := p.p.CancelOrder(req.ID, req.PaymentType)
	if err == nil {
		if order.Status() == payment.PaymentStatusPending {
			return nil, nil
		}
		boolValue := order.Status() == payment.PaymentStatusCanceled
		done = &boolValue
		req.PaymentOrder = order
	}
	return done, err
}

func (p *PaymentJob) Subscribe() (<-chan booking.JobResponse, error) {
	return p.updateCh, nil
}

// conumeIncomingChanges changes reservation status when payment order update happens
func (p *PaymentJob) conumeIncomingChanges() error {
	go func() {
		updatesCh := p.p.SubscribeOnStatusUpdates()
		for {
			select {
			case paymentStatusUpdate := <-updatesCh:
				p.updateCh <- booking.JobResponse{
					ReservationID: paymentStatusUpdate.ReservationID(),
					IsSucceeded:   paymentStatusUpdate.Status() == payment.PaymentStatusSuccess,
					UpdateData: func(r *booking.Reservation) {
						r.PaymentOrder = paymentStatusUpdate
					},
					JobName: p.Name(),
				}
			case <-p.doneCh:
				return
			}
		}
	}()
	return nil
}

func (p *PaymentJob) Start(_ context.Context) error {
	if err := p.conumeIncomingChanges(); err != nil {
		return err
	}

	return nil
}

func (p *PaymentJob) Stop(_ context.Context) error {
	close(p.doneCh)
	return nil
}
