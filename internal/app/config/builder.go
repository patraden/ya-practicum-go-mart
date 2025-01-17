package config

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"

	e "github.com/patraden/ya-practicum-go-mart/internal/app/domain/errors"
)

type builder struct {
	cfg *Config
}

func newBuilder() *builder {
	return &builder{
		cfg: DefaultConfig(),
	}
}

func (b *builder) loadEnv() {
	if err := env.Parse(b.cfg); err != nil {
		log.Fatal(e.ErrConfigEnvParse)
	}
}

func (b *builder) loadFlags() {
	flag.StringVar(&b.cfg.RunAddress, "a", b.cfg.RunAddress, "server address {host}:{port}")
	flag.StringVar(&b.cfg.DatabaseURI, "d", b.cfg.DatabaseURI, "postgres uri")
	flag.StringVar(&b.cfg.AccrualAddress, "r", b.cfg.AccrualAddress, "accrual system address {host}:{port}")
	flag.Parse()
}

func (b *builder) getConfig() *Config {
	b.loadFlags()
	b.loadEnv()

	if b.cfg.DatabaseURI == `` {
		log.Fatal(e.ErrConfigDBURI)
	}

	return b.cfg
}
