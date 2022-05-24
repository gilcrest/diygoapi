package logger_test

import (
	"os"

	"github.com/gilcrest/diy-go-api/domain/logger"

	"github.com/rs/zerolog"
)

func ExampleNewLogger() {
	lgr := logger.NewLogger(os.Stdout, zerolog.DebugLevel, false)
	lgr.Trace().Msg("Trace is lower than Debug, this message is filtered out")
	lgr.Debug().Msg("This is a log at the Debug level")

	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	lgr.Debug().Msg("Logging level raised to Error, Debug is lower than Error, this message is filtered out")
	lgr.Error().Msg("This is a log at the Error level")

	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	lgr.Trace().Msg("Setting Global level will not impact minimum set to logger, this trace message will still be filtered out")
	lgr.Debug().Msg("Logging level raised all the way down to Trace level, Debug is higher than Trace, this will log")

	// Output:
	// {"level":"debug","severity":"DEBUG","message":"This is a log at the Debug level"}
	// {"level":"error","severity":"ERROR","message":"This is a log at the Error level"}
	// {"level":"debug","severity":"DEBUG","message":"Logging level raised all the way down to Trace level, Debug is higher than Trace, this will log"}
}
