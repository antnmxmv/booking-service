package config

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

const ChangeWatcherDelay = time.Minute

type Loader struct {
	filePath string
	doneCh   chan struct{}
	data     *Config
}

func (c *Loader) GetData() *Config {
	return c.data
}

func NewConfig() *Config {
	return &Config{}
}

func NewLoader(cnf *Config) *Loader {
	res := &Loader{
		doneCh: make(chan struct{}),
		data:   cnf,
	}
	flag.StringVar(&res.filePath, "c", "./config.yml", "override path to config.yml file")

	return res
}

func (c *Loader) loadFile() error {
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

	if duration, err := time.ParseDuration(c.data.Payment.Card.TimeoutStr); err != nil {
		c.data.Payment.Card.Timeout = time.Minute * 30
		c.data.Payment.Card.TimeoutStr = c.data.Payment.Card.Timeout.String()
	} else {
		c.data.Payment.Card.Timeout = duration
	}

	if _, err := strconv.ParseUint(c.data.Prometheus.Port, 10, 32); err != nil {
		c.data.Prometheus.Port = "2112"
	}

	if c.data.Prometheus.Path == "" || c.data.Prometheus.Path[0] != '/' {
		c.data.Prometheus.Path = "/metrics"
	}
	return nil
}

func (c *Loader) Start(ctx context.Context) error {
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

func (c *Loader) Stop(ctx context.Context) error {
	close(c.doneCh)
	return nil
}
