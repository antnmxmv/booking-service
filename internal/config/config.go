package config

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config interface {
	GetData() Data
}

const ChangeWatcherDelay = time.Minute

type Data struct {
	Server  Server  `yaml:"server"`
	Booking Booking `yaml:"booking"`
}

type Server struct {
	Port  string `yaml:"port"`
	Debug bool   `yamls:"debug"`
}

type Booking struct {
	IdleReservationTimeoutStr string `yaml:"idleReservationTimeout"`
	IdleReservationTimeout    time.Duration
}

type configWatcher struct {
	filePath string
	doneCh   chan struct{}
	data     Data
}

func (c *configWatcher) GetData() Data {
	return c.data
}

func NewConfig() *configWatcher {
	res := &configWatcher{
		doneCh: make(chan struct{}),
	}
	flag.StringVar(&res.filePath, "c", "./config.yml", "override path to config.yml file")
	return res
}

func (c *configWatcher) loadFile() error {
	configBytes, err := os.ReadFile(c.filePath)
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(configBytes, &c.data); err != nil {
		return err
	}
	if duration, err := time.ParseDuration(c.data.Booking.IdleReservationTimeoutStr); err != nil {
		c.data.Booking.IdleReservationTimeout = time.Minute * 30
		c.data.Booking.IdleReservationTimeoutStr = c.data.Booking.IdleReservationTimeout.String()
	} else {
		c.data.Booking.IdleReservationTimeout = duration
	}
	return nil
}

func (c *configWatcher) Start(ctx context.Context) error {
	if err := c.loadFile(); err != nil {
		return fmt.Errorf("loading config file: %s", err.Error())
	}
	go func() {
		t := time.NewTicker(ChangeWatcherDelay)

		for {
			select {
			case <-t.C:
				if err := c.loadFile(); err != nil {
					log.Println("error while reloading config file")
				}
			case <-c.doneCh:
				return
			}
		}
	}()

	return nil
}

func (c *configWatcher) Stop(ctx context.Context) error {
	close(c.doneCh)
	return nil
}
