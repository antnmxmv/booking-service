package main

import (
	"sync/atomic"

	"github.com/antnmxmv/booking-service/data"
	"github.com/antnmxmv/booking-service/internal/api"
	"github.com/antnmxmv/booking-service/internal/api/handlers"
	"github.com/antnmxmv/booking-service/internal/booking"
	"github.com/antnmxmv/booking-service/internal/booking/jobs"
	"github.com/antnmxmv/booking-service/internal/config"
	"github.com/antnmxmv/booking-service/internal/payment"
	"github.com/antnmxmv/booking-service/internal/price"
	"github.com/antnmxmv/booking-service/internal/storage/inmemory"
	"github.com/antnmxmv/booking-service/pkg/queue"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func createApp() *fx.App {
	isReady := &atomic.Bool{}

	return fx.New(
		fx.WithLogger(func() fxevent.Logger {
			return &logger{onStartedFn: func() { isReady.Store(true) }}
		}),

		fx.Provide(
			inmemory.NewStorage().WithRoomAvailability(data.RoomAvailability).Build,

			func() handlers.ReadinessMonitor {
				return isReady
			},

			config.NewConfig,
			config.NewLoader,

			payment.NewCardSource,
			payment.NewCashSource,
			func(cash *payment.CashSource, card *payment.CardSource) *payment.Provider {
				return payment.NewPaymentProvider(cash)
			},

			price.NewExampleProvider,

			AsReservationJob(jobs.NewPriceJob, `name:"price-job"`),
			AsReservationJob(jobs.NewPaymentJob, `name:"payment-job"`),
			AsReservationJob(jobs.NewNotificationJob, `name:"notification-job"`),

			fx.Annotate(
				booking.NewReservationOrchestrator,
				fx.ParamTags(
					`name:""`,
					`group:"reservation-jobs"`,
				),
			),
			fx.Annotate(queue.NewDelayedQueue[string], fx.As(new(booking.DelayedQueue))),

			booking.NewBookingService,
			api.NewController,
		),

		fx.Invoke(
			AsHook[*config.Loader],
			AsHook[*payment.CardSource],
			AsHook[*payment.Provider],

			AsHook[*booking.BookingService],
			AsHook[*api.Controller],
		),
	)
}

func main() {

	app := createApp()
	app.Run()
}
