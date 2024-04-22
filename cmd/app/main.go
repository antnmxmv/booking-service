package main

import (
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

func main() {
	// place where we import data
	repository := inmemory.NewStorage().WithRoomAvailability(data.RoomAvailability)

	app := container.NewApp()

	config := config.NewConfig()
	app.AddContainer(config)

	cardPaymentSource := payment.NewCardSource(time.Second * 500)
	app.AddContainer(cardPaymentSource)

	paymentProvider := payment.NewPaymentProvider(cardPaymentSource, payment.NewCashSource())
	app.AddContainer(paymentProvider)
	paymentJob := jobs.NewPaymentJob(paymentProvider, repository)
	app.AddContainer(paymentJob)

	// building reservation strategy
	reservationOrchestrator := booking.NewReservationOrchestrator(
		repository,
		jobs.NewPriceJob(price.NewExampleProvider()),
		paymentJob,
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

	app.Run()
}
