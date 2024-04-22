package payment

import (
	"context"
	"errors"
	"sync"
)

// Provider is payment source decorators factory.
// payment sources in real life could have more options. card payment use sms
// for approval or not use it, or it may be different types of merchants
type Provider struct {
	sources   map[SourceType]Source
	updatesCh chan Order
}

func NewPaymentProvider(sources ...Source) *Provider {
	res := &Provider{
		sources:   make(map[SourceType]Source, len(sources)),
		updatesCh: make(chan Order),
	}

	// merge source channels

	wg := sync.WaitGroup{}

	wg.Add(len(sources))

	for _, source := range sources {
		source := source
		go func() {
			for update := range source.subscribe() {
				res.updatesCh <- update
			}
			wg.Done()
		}()
		res.sources[source.name()] = source
	}
	go func() {
		wg.Wait()
		close(res.updatesCh)
	}()

	return res
}

// GetSources returns array of included payment provider types
func (p *Provider) GetSources() map[SourceType]Source {
	return p.sources
}

func (p *Provider) UnmarshalDetailsJSON(pType SourceType, msg []byte) (OrderDetails, error) {
	if provider, ok := p.sources[pType]; ok {
		return provider.unmarshalDetailsJSON(msg)
	}
	return nil, errors.New("payment provider not supported")
}

// CreateOrder creates payment order using reservationID as identifier
func (p *Provider) CreateOrder(reservationID string, amount int, sourceType SourceType, details OrderDetails) (Order, error) {
	order, err := p.sources[sourceType].createOrder(reservationID, amount, details)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (p *Provider) CancelOrder(reservationID string, sourceType SourceType) (Order, error) {
	return p.sources[sourceType].cancelOrder(reservationID)
}

func (p *Provider) SubscribeOnStatusUpdates() <-chan Order {
	return p.updatesCh
}

func (p *Provider) Start(ctx context.Context) error {
	return nil
}

func (p *Provider) Stop(ctx context.Context) error {
	return nil
}
