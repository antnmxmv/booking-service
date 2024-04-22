package price

import (
	"sync"

	"github.com/antnmxmv/booking-service/internal/booking"
	"github.com/antnmxmv/booking-service/internal/booking/jobs"
)

// ExamplePriceService very simple example of discount service logic.
// all price generation logic should be in another service where managers
// will be able to create  temporary discounts of many types and conditions
type ExamplePriceService struct {
	userOrdersCount map[string]int
	mux             sync.Mutex
}

func NewExampleProvider() jobs.PriceServiceFacade {
	return &ExamplePriceService{userOrdersCount: map[string]int{}}
}

func (p *ExamplePriceService) GetPrice(reservation booking.Reservation) (discounts []string, finalCost int, err error) {
	discounts = []string{}
	totalRoomsCount := uint(0)
	// base cost
	for _, r := range reservation.RoomTypes {
		if r.RoomType == "lux" {
			finalCost += 1000
		} else {
			finalCost += 500
		}
		totalRoomsCount += r.Count
	}

	// party discount example
	if totalRoomsCount > 2 {
		finalCost -= int(float64(finalCost) * 0.01)
		discounts = append(discounts, "party")
	}

	// first order discount example (user wide, hotel wide, etc)

	p.mux.Lock()
	isFirstOrder := p.userOrdersCount[reservation.UserID] == 0
	// TODO profile service facade mock to get completed orders by user id
	p.userOrdersCount[reservation.UserID] += 1
	p.mux.Unlock()

	if isFirstOrder {
		finalCost -= int(float64(finalCost) * 0.05)
		discounts = append(discounts, "first_order")
	}

	return discounts, finalCost, nil
}
