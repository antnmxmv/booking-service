package middlewares

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/antnmxmv/booking-service/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Prometheus contains the metrics gathered by the instance and its path
type Prometheus struct {
	reqDur *prometheus.HistogramVec
	router *gin.Engine
	cnf    *config.Config
	server *http.Server
}

// NewPrometheus generates a new set of metrics with a certain subsystem name
func NewPrometheus(cnf *config.Config) *Prometheus {
	p := &Prometheus{
		cnf: cnf,
	}

	p.registerMetrics()

	return p
}

func (p *Prometheus) registerMetrics() {
	p.reqDur = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "Histogram request latencies",
			Buckets:   []float64{.005, .01, .02, 0.04, .06, 0.08, .1, 0.15, .25, 0.4, .6, .8, 1, 1.5, 2, 3, 5},
		},
		[]string{"code", "handler"},
	)

	err := prometheus.Register(p.reqDur)
	if err != nil {
		fmt.Println(err.Error())
	}
}

// HandlerFunc defines handler function for middleware
func (p *Prometheus) Middleware(label string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		elapsed := float64(time.Since(start)) / float64(time.Second)
		p.reqDur.WithLabelValues(status, label).Observe(elapsed)
	}
}

func (p *Prometheus) Start(ctx context.Context) error {
	h := promhttp.Handler()

	r := gin.New()
	r.Any(p.cnf.Prometheus.Path, gin.WrapH(h))
	p.server = &http.Server{Addr: "0.0.0.0:" + p.cnf.Prometheus.Port, Handler: r}
	go func() {

		err := p.server.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatalf("http server shutdown with error: %s", err.Error())
		}
	}()
	return nil
}

func (p *Prometheus) Stop(ctx context.Context) error {
	return p.server.Shutdown(ctx)
}
