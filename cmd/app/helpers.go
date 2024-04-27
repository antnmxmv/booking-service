package main

import (
	"context"
	"log"
	"reflect"
	"time"

	"github.com/antnmxmv/booking-service/internal/booking"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func AsReservationJob(f any, extraTags ...string) any {
	return fx.Annotate(
		f,
		fx.As(new(booking.Job)),
		fx.ResultTags(append([]string{`group:"reservation-jobs"`}, extraTags...)...),
	)
}

func AsPaymentSource(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`group:"payment-sources"`),
	)
}

type Module interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

func AsHook[M Module](m M, lc fx.Lifecycle) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			startTime := time.Now()
			if err := m.Start(ctx); err == nil {
				log.Printf("[%s] started in %s", reflect.TypeOf(m).String(), time.Since(startTime).String())
				return nil
			} else {
				log.Printf("[%s] start failed after %s! err: %s", reflect.TypeOf(m).String(), time.Since(startTime).String(), err.Error())
				return err
			}
		},
		OnStop: func(ctx context.Context) error {
			startTime := time.Now()
			if err := m.Stop(ctx); err == nil {
				log.Printf("[%s] stopped in %s", reflect.TypeOf(m).String(), time.Since(startTime).String())
				return nil
			} else {
				log.Printf("[%s] stop failed after %s! err: %s", reflect.TypeOf(m).String(), time.Since(startTime).String(), err.Error())
				return err
			}
		},
	})
}

type logger struct {
	onStartedFn func()
	startTime   time.Time
	stopTime    time.Time
}

func (o *logger) LogEvent(e fxevent.Event) {
	switch e.(type) {
	case *fxevent.LoggerInitialized:
		o.startTime = time.Now()
	case *fxevent.Stopping:
		o.stopTime = time.Now()
	case *fxevent.Started:
		o.onStartedFn()
		log.Println("[app] started in", time.Since(o.startTime).String())
	case *fxevent.Stopped:
		log.Printf("[app] stopped in %s total uptime: %s", time.Since(o.stopTime).String(), time.Since(o.startTime).String())
	}

	// check if event contains non-nil Err field

	rv := reflect.Indirect(reflect.ValueOf(e))
	rt := rv.Type()
	for i, limit := 0, rt.NumField(); i < limit; i++ {
		fld := rt.Field(i)

		nv, ok := rv.FieldByName(fld.Name).Interface().(error)
		if ok {
			log.Printf("[app] fx_event: %T err: %s", e, nv.Error())
		}
	}
}
