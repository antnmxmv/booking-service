package config

import "time"

type Config struct {
	Server     Server     `yaml:"server"`
	Booking    Booking    `yaml:"booking"`
	Payment    Payment    `yaml:"payment"`
	Prometheus Prometheus `yaml:"prometheus"`
}

type Server struct {
	Port  string `yaml:"port"`
	Debug bool   `yamls:"debug"`
}

type Booking struct {
	IdleReservationTimeoutStr string        `yaml:"idleReservationTimeout"`
	IdleReservationTimeout    time.Duration `yaml:"-"`
}

type Payment struct {
	Card Card `yaml:"card"`
}

type Prometheus struct {
	Port string `yaml:"port"`
	Path string `yaml:"path"`
}

type Card struct {
	TimeoutStr string        `yaml:"timeout"`
	Timeout    time.Duration `yaml:"-"`
}
