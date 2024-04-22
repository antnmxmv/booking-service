package container

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"reflect"
	"sync/atomic"
	"syscall"
	"time"
)

const StartTimeout = time.Second * 15
const StopTimeout = time.Second * 15

type module interface {
	Start(context.Context) error
	Stop(context.Context) error
}

type App struct {
	modules []module
	IsReady *atomic.Bool
}

func NewApp() *App {
	return &App{
		modules: []module{},
		IsReady: &atomic.Bool{},
	}
}

func (app *App) AddContainer(module module) {
	app.modules = append(app.modules, module)
}

func (app *App) stopFromEndToStart(lastIndex int) error {
	ctx, cancel := context.WithTimeout(context.Background(), StopTimeout)
	defer cancel()

	go func() {
		<-ctx.Done()
		if ctx.Err() == context.DeadlineExceeded {
			log.Fatal("graceful shutdown timed out.. forcing exit.")
		}
	}()

	if lastIndex == -1 {
		lastIndex = len(app.modules)
	}

	for j := lastIndex - 1; j >= 0; j-- {
		err := app.modules[j].Stop(ctx)
		if err != nil {
			// TODO log error and go on
		}
	}
	return nil
}

func getModuleName(module module) string {
	if module == nil {
		return "nil"
	}
	return reflect.TypeOf(module).String()
}

func (app *App) Stop() {
	app.stopFromEndToStart(-1)
}

func (app *App) Run() {
	ctx, cancel := context.WithTimeout(context.Background(), StartTimeout)
	defer cancel()
	logPrefix := "[booking-service] "
	for i, module := range app.modules {
		moduleLogPrefix := logPrefix + fmt.Sprintf("[%v] ", getModuleName(module))
		log.Println(moduleLogPrefix + "starting...")
		err := module.Start(ctx)
		if err != nil {
			log.Printf(logPrefix+"error while starting app: %s . shutdown previously started modules and exit\n", err.Error())
			log.Println(moduleLogPrefix + "start error!")
			app.stopFromEndToStart(i)
			return
		}
		log.Println(moduleLogPrefix + "started!")
	}

	log.Println(logPrefix + "started")

	app.IsReady.Store(true)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-sig
	app.Stop()
	fmt.Println("gracefuly stopped")
}
