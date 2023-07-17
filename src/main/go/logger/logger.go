package logger

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

func InitConsoleLogger(debug bool) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC822})
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

func ExitIfError(err error) {
	if err != nil {
		log.Error().Msg(err.Error())
		os.Exit(1)
	}
}

func ExitIfErrorWithMsg(err error, msg string) {
	if err != nil {
		log.Error().Msg(err.Error())
		os.Exit(1)
	}
}
