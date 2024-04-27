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
	"github.com/gin-gonic/gin"
)

type Controller struct {
	cfg              *config.Config
	s                *booking.BookingService
	p                *payment.Provider
	isReady          handlers.ReadinessMonitor
	server           *http.Server
	prometheusServer *middlewares.Prometheus
}

func NewController(conf *config.Config, s *booking.BookingService, p *payment.Provider, readinessMonitor handlers.ReadinessMonitor, prometheus *middlewares.Prometheus) *Controller {
	return &Controller{
		s:                s,
		cfg:              conf,
		p:                p,
		isReady:          readinessMonitor,
		prometheusServer: prometheus,
	}
}

func (c *Controller) buildServer() {

	interceptors := []gin.HandlerFunc{
		gin.Recovery(),
	}

	if c.cfg.Server.Debug {
		gin.SetMode(gin.ReleaseMode)
		interceptors = append(interceptors, middlewares.Logger)
	}

	r := gin.New()

	r.Use(interceptors...)

	r.GET("/reservation/", handlers.NewGetReservationsHandler(c.s))
	r.POST("/reservation/", c.prometheusServer.Middleware("create_reservation"), handlers.NewCreateReservationHandler(c.s, c.p))
	r.GET("/hotel/:hotelID/", handlers.NewGetRoomsHandler(c.s))

	r.Handle(http.MethodGet, "/readyz", handlers.NewReadyzHandler(c.isReady))

	c.server = &http.Server{Addr: "0.0.0.0:" + c.cfg.Server.Port, Handler: r}
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
