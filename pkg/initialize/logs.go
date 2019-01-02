package initialize

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func initLogs() {
	zerolog.TimeFieldFormat = ""

	log.Logger = log.With().Caller().Logger()
}
