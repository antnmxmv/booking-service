package payment

type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusSuccess  PaymentStatus = "success"
	PaymentStatusCanceled PaymentStatus = "canceled"
	PaymentStatusFailed   PaymentStatus = "failed"
)

type Order interface {
	ReservationID() string
	Status() PaymentStatus
}

type SourceType string

type OrderDetails interface {
	Type() SourceType
}

// Source is decorator needed to generalize logic of payment strategies
// it must be able to provide order status by reservationID
type Source interface {
	name() SourceType
	// createOrder creates an order
	// if order is not in completed state ('finished' or 'created') after creation,
	// observer would continiously check it's status
	createOrder(reservationID string, amount int, details OrderDetails) (Order, error)

	cancelOrder(reservationID string) (Order, error)

	unmarshalDetailsJSON([]byte) (OrderDetails, error)

	subscribe() <-chan Order
}
