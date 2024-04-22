package main

import (
	"context"
	"time"

	"github.com/antnmxmv/booking-service/data"
	"github.com/antnmxmv/booking-service/internal/api"
	"github.com/antnmxmv/booking-service/internal/booking"
	"github.com/antnmxmv/booking-service/internal/booking/jobs"
	"github.com/antnmxmv/booking-service/internal/config"
	"github.com/antnmxmv/booking-service/internal/payment"
	"github.com/antnmxmv/booking-service/internal/price"
	"github.com/antnmxmv/booking-service/internal/storage/inmemory"
	"github.com/antnmxmv/booking-service/pkg/container"
	"github.com/antnmxmv/booking-service/pkg/queue"
)

type BaseContainer struct{}

func (m BaseContainer) Start(_ context.Context) error {
	return nil
}

func (m BaseContainer) Stop(_ context.Context) error {
	return nil
}

var DefaultConfigData = config.Data{
	Server: config.Server{
		Port:  "8080",
		Debug: true,
	},
	Booking: config.Booking{
		IdleReservationTimeoutStr: "5s",
	},
}

type Config struct {
	config.Data
	BaseContainer
}

func (m *Config) GetData() config.Data {
	return m.Data
}

func runTestingApp() *container.App {
	app := container.NewApp()

	config := &Config{Data: DefaultConfigData}
	app.AddContainer(config)

	cardPaymentSource := payment.NewCardSource(time.Second * 5)
	app.AddContainer(cardPaymentSource)

	paymentProvider := payment.NewPaymentProvider(cardPaymentSource, payment.NewCashSource())
	app.AddContainer(paymentProvider)

	repository := inmemory.NewStorage().WithRoomAvailability(data.RoomAvailability)
	// building reservation strategy
	reservationOrchestrator := booking.NewReservationOrchestrator(
		repository,
		jobs.NewPriceJob(price.NewExampleProvider()),
		// without payment job
		jobs.NewNotificationJob(),
	)
	bookingService := booking.NewBookingService(
		config,
		repository,
		paymentProvider,
		reservationOrchestrator,
		queue.NewDelayedQueue[string](),
	)
	app.AddContainer(bookingService)

	controller := api.NewController(bookingService, paymentProvider, config, app.IsReady)
	app.AddContainer(controller)

	go app.Run()

	return app
}
