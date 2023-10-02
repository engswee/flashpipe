package logger

import (
	"github.com/go-errors/errors"
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
		// Display stack trace based on type of error
		switch err.(type) {
		case *errors.Error:
			log.Fatal().Msg(err.(*errors.Error).ErrorStack())
		default:
			log.Fatal().Msg(err.Error())
		}
	}
}
