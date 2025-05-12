package internal

import (
	"github.com/caarlos0/env/v11"
	"log"
	"sync"
)

type Env struct {
	ReporterPort       int    `env:"REPORTER_PORT" envDefault:"3000"`
	GameAddress        string `env:"GAME_A2S_ADDRESS"`
	GamePort           int    `env:"GAME_A2S_PORT"`
	QueryInterval      int    `env:"QUERY_INTERVAL" envDefault:"10000"`
	QueryTimeout       int    `env:"QUERY_TIMEOUT" envDefault:"3000"`
	QueryMaxPacketSize int    `env:"QUERY_MAX_PACKET_SIZE" envDefault:"14000"` // Some engine does not follow the protocol spec, and may require bigger packet buffer
}

var (
	envVar *Env = &Env{}
	once   sync.Once
)

func GetEnvironmentVars() Env {
	once.Do(func() {
		if err := env.Parse(envVar); err != nil {
			log.Fatal(err)
		}
	})
	return *envVar
}
