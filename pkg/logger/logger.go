package logger

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
	"github.com/rs/zerolog/log"
)

var (
	diodeCloser io.Closer
)

func Init(opt ...Option) {
	options := &Options{}
	*options = defaultOptions

	for _, o := range opt {
		o(options)
	}

	setLevel(options)

	writers := make([]io.Writer, 0, 2)
	setConsole(options, &writers)
	diodeCloser = setFile(options, &writers)
	multiWriter := zerolog.MultiLevelWriter(writers...)

	log.Logger = zerolog.New(multiWriter).With().Int("pid", os.Getpid()).Timestamp().Caller().Logger()
	log.Info().Msg("log init success")
}

func Close() error {
	if diodeCloser != nil {
		return diodeCloser.Close()
	}

	return nil
}

func SetLevel(l string) error {
	level, err := zerolog.ParseLevel(l)
	if err != nil {
		return nil
	}
	zerolog.SetGlobalLevel(level)
	return nil
}

func setLevel(options *Options) {
	level, err := zerolog.ParseLevel(options.level)
	if err != nil {
		level = zerolog.TraceLevel
	}
	zerolog.SetGlobalLevel(level)
}

func setConsole(options *Options, writers *[]io.Writer) {
	if options.console {
		console := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		*writers = append(*writers, console)
	}
}

func setFile(options *Options, writers *[]io.Writer) io.Closer {
	if options.file {
		diodeWriter := diode.NewWriter(newFileWriter(options.dir, options.name, options.ext),
			500000, 0, func(missed int) {
				fmt.Fprintf(os.Stderr, "Logger Dropped %d messages", missed)
			})

		*writers = append(*writers, &diodeWriter)
		return &diodeWriter
	}

	return nil
}
