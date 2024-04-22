package payment

// CashSource is example of synchronious payment source
type CashSource struct {
}

// subscribe returns closed channel
func (cp *CashSource) subscribe() <-chan Order {
	ch := make(chan Order)
	close(ch)
	return ch
}

func (cp *CashSource) name() SourceType {
	return "cash"
}

func NewCashSource() *CashSource {
	return &CashSource{}
}

// createOrder creates order in completed state 'success' state
func (cp *CashSource) createOrder(reservationID string, amount int, _ OrderDetails) (Order, error) {
	return &cashPaymentOrder{
		PaymentStatus: PaymentStatusSuccess,
		RID:           reservationID,
	}, nil
}

// cash payment order can be canceled even if succeeded
func (cp *CashSource) cancelOrder(reservationID string) (Order, error) {
	return &cardPaymentOrder{
		PaymentStatus: PaymentStatusCanceled,
		RID:           reservationID,
	}, nil
}

func (cp *CashSource) unmarshalDetailsJSON([]byte) (OrderDetails, error) {
	return nil, nil
}

type cashPaymentOrder struct {
	PaymentStatus PaymentStatus `json:"status"`
	RID           string        `json:"-"`
}

func (cp *cashPaymentOrder) ReservationID() string {
	return cp.RID
}

func (cp *cashPaymentOrder) Status() PaymentStatus {
	return cp.PaymentStatus
}
