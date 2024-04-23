package payment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"sync"
	"time"

	"github.com/antnmxmv/booking-service/internal/config"
)

// CardSource is asynchronious payment source. It sends a random state other than 'pending' with delay
// this is example of payment source can create order for already known credentials by their id
// and for sms verification it generates link to web form
type CardSource struct {
	cnf       *config.Config
	lastId    uint
	updatesCh chan Order
	// ordersCancelingChans is map of doneCh for every pending order
	// we need it to handle order cancelation just for example
	ordersCancelingChans map[string]chan struct{}
	mux                  sync.Mutex
}

func (cp *CardSource) subscribe() <-chan Order {
	return cp.updatesCh
}

func NewCardSource(cnf *config.Config) *CardSource {
	rand.Seed(time.Now().Unix())
	return &CardSource{
		cnf:                  cnf,
		lastId:               0,
		updatesCh:            make(chan Order),
		ordersCancelingChans: make(map[string]chan struct{}),
	}
}

func (cp *CardSource) name() SourceType {
	return "card"
}

func (cp *CardSource) createOrder(reservationID string, amount int, details OrderDetails) (Order, error) {
	// type assertion
	request, ok := details.(cardOrderDetails)
	if !ok {
		return nil, fmt.Errorf("card payment source needs *cardOrderDetails, but got %s", reflect.TypeOf(details).String())
	}

	cp.mux.Lock()
	defer cp.mux.Unlock()

	defer func() { cp.lastId += 1 }()
	paymentLink := fmt.Sprintf("http://merchant-url/card/%d/order/%d", request.CardID, cp.lastId+1)

	res := &cardPaymentOrder{
		URL:           paymentLink,
		Comment:       "it will randomly become successful or failed in 5 seconds",
		PaymentStatus: PaymentStatusPending,
		RID:           reservationID,
	}
	cp.ordersCancelingChans[reservationID] = make(chan struct{})

	go func(order cardPaymentOrder) {

		if rand.Int()%2 == 0 {
			order.PaymentStatus = PaymentStatusFailed
			order.Comment = "transfer declined"
		} else {
			order.PaymentStatus = PaymentStatusSuccess
			order.Comment = "transfer accepted"
		}

		t := time.NewTimer(cp.cnf.Payment.Card.Timeout)
		select {
		case <-t.C:
			delete(cp.ordersCancelingChans, reservationID)
			cp.updatesCh <- order
		case <-cp.ordersCancelingChans[reservationID]:
			t.Stop()
		}
	}(*res)

	return res, nil
}

func (cp *CardSource) cancelOrder(reservationID string) (Order, error) {
	cp.mux.Lock()
	defer cp.mux.Unlock()
	if ch, ok := cp.ordersCancelingChans[reservationID]; ok {
		close(ch)
	}
	return cardPaymentOrder{
		RID:           reservationID,
		URL:           "",
		PaymentStatus: PaymentStatusCanceled,
		Comment:       "",
	}, nil
}

type cardPaymentOrder struct {
	RID           string        `json:"-"`
	URL           string        `json:"url"`
	PaymentStatus PaymentStatus `json:"status"`
	Comment       string        `json:"comment"`
}

func (c cardPaymentOrder) ReservationID() string {
	return c.RID
}

func (cp cardPaymentOrder) Status() PaymentStatus {
	return cp.PaymentStatus
}

type cardOrderDetails struct {
	CardID uint `json:"card_id"`
}

func (c cardOrderDetails) Type() SourceType {
	return "card"
}

func (cp *CardSource) unmarshalDetailsJSON(req []byte) (OrderDetails, error) {
	res := cardOrderDetails{}
	if err := json.Unmarshal(req, &res); err != nil {
		return res, errors.New(`please provide {\"card_id\": 0} formatted object in payment_details`)
	}
	return res, nil
}

func (cp *CardSource) Start(_ context.Context) error {
	// assume that it will connect to some message queue to get updates
	return nil
}

func (cp *CardSource) Stop(_ context.Context) error {
	// and stop subscription here
	return nil
}
