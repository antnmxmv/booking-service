package jobs

import "github.com/antnmxmv/booking-service/internal/booking"

type PriceServiceFacade interface {
	GetPrice(reservationRequest booking.Reservation) (discounts []string, finalCost int, err error)
}

type PriceJob struct {
	p  PriceServiceFacade
	ch chan booking.JobResponse
}

func NewPriceJob(p PriceServiceFacade) *PriceJob {
	// it will be synchronious
	ch := make(chan booking.JobResponse)
	close(ch)
	return &PriceJob{
		p:  p,
		ch: ch,
	}
}

func (p *PriceJob) Name() booking.ReservationStatus {
	return "price_calculation"
}

func (p *PriceJob) Run(r *booking.Reservation) (*bool, error) {
	discountIDs, cost, err := p.p.GetPrice(*r)
	if err != nil {
		return nil, err
	}
	r.AppliedDiscountIDs = discountIDs
	r.Cost = cost
	res := true
	return &res, nil
}

func (p *PriceJob) Cancel(_ *booking.Reservation) (*bool, error) {
	// no need to cancel anything, but here could be request to
	// discount service to reduce some user metrics
	res := true
	return &res, nil
}

func (p *PriceJob) Subscribe() (<-chan booking.JobResponse, error) {
	return p.ch, nil
}
