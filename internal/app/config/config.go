package config

import "time"

const (
	defaultReadHeaderTimeout = 10 * time.Second
	defaultWriteTimeout      = 10 * time.Second
	defaultIdleTimeout       = 120 * time.Second
	defaultShutdownTimeOut   = 5 * time.Second
)

type Config struct {
	RunAddress            string `env:"RUN_ADDRESS"`
	DatabaseURI           string `env:"DATABASE_URI"`
	AccrualAddress        string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	JWTSecret             string `env:"JWT_SECRET"`
	HTTPReadHeaderTimeout time.Duration
	HTTPWriteTimeout      time.Duration
	HTTPIdleTimeout       time.Duration
	HTTPShutdownTimeOut   time.Duration
}

func DefaultConfig() *Config {
	return &Config{
		RunAddress:            `localhost:8080`,
		AccrualAddress:        `accrual:8081`,
		DatabaseURI:           ``,
		JWTSecret:             `d1a58c288a0226998149277b14993f6c73cf44ff9df3de548df4df25a13b251a`,
		HTTPReadHeaderTimeout: defaultReadHeaderTimeout,
		HTTPWriteTimeout:      defaultWriteTimeout,
		HTTPIdleTimeout:       defaultWriteTimeout,
		HTTPShutdownTimeOut:   defaultShutdownTimeOut,
	}
}

func LoadConfig() *Config {
	builder := newBuilder()
	cfg := builder.getConfig()

	return cfg
}
