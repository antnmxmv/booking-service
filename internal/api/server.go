package api

import (
	"context"
	"log"
	"net/http"

	"github.com/antnmxmv/booking-service/internal/api/handlers"
	"github.com/antnmxmv/booking-service/internal/api/middlewares"
	"github.com/antnmxmv/booking-service/internal/booking"
	"github.com/antnmxmv/booking-service/internal/config"
	"github.com/antnmxmv/booking-service/internal/payment"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Controller struct {
	cfg     *config.Config
	s       *booking.BookingService
	p       *payment.Provider
	isReady handlers.ReadinessMonitor
	server  *http.Server
}

func NewController(conf *config.Config, s *booking.BookingService, p *payment.Provider, readinessMonitor handlers.ReadinessMonitor) *Controller {
	return &Controller{
		s:       s,
		cfg:     conf,
		p:       p,
		isReady: readinessMonitor,
	}
}

func (c *Controller) buildServer() {
	router := chi.NewRouter()

	if c.cfg.Server.Debug {
		router.Use(middlewares.Logger())
	}

	router.Use(middleware.Recoverer)
	router.Use(middleware.AllowContentType("application/json"))

	router.Get("/reservation/", handlers.NewGetReservationsHandler(c.s))
	router.Post("/reservation/", handlers.NewCreateReservationHandler(c.s, c.p))
	router.Get("/hotel/{hotelID}/", handlers.NewGetRoomsHandler(c.s))

	router.Handle("/readyz", handlers.NewReadyzHandler(c.isReady))

	c.server = &http.Server{Addr: "0.0.0.0:" + c.cfg.Server.Port, Handler: router}
}

func (c *Controller) Start(ctx context.Context) error {
	c.buildServer()
	go func() {
		err := c.server.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatalf("http server shutdown with error: %s", err.Error())
		}
	}()
	return nil
}

func (c *Controller) Stop(ctx context.Context) error {
	return c.server.Shutdown(ctx)
}
