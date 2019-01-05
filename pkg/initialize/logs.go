package initialize

import (
	"os"
	"runtime"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

func initLogs() {
	if runtime.GOOS == "darwin" {
		// Use ConsoleWriter locally in development
		zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	zerolog.TimeFieldFormat = ""

	zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stdout})
}
