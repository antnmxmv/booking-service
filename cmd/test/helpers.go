package main

import (
	"context"

	"github.com/antnmxmv/booking-service/data"
	"github.com/antnmxmv/booking-service/internal/api"
	"github.com/antnmxmv/booking-service/internal/api/middlewares"
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

var DefaultConfigData = config.Config{
	Server: config.Server{
		Port:  "8080",
		Debug: true,
	},
	Booking: config.Booking{
		IdleReservationTimeoutStr: "5s",
	},
}

type Config struct {
	*config.Config
	BaseContainer
}

func (m *Config) GetData() config.Config {
	return *m.Config
}

func runTestingApp() *container.App {
	app := container.NewApp()

	config := &Config{Config: &DefaultConfigData}
	app.AddContainer(config)

	cardPaymentSource := payment.NewCardSource(config.Config)
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
		config.Config,
		repository,
		reservationOrchestrator,
		queue.NewDelayedQueue[string](),
	)
	app.AddContainer(bookingService)

	controller := api.NewController(config.Config, bookingService, paymentProvider, app.IsReady, middlewares.NewPrometheus(config.Config))
	app.AddContainer(controller)

	go app.Run()

	return app
}
