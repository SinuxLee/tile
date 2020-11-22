package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
	"github.com/rs/zerolog/log"
)

const (
	timeFormat = "2006-01-02 15:04:05.000"
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
	// 配置日志
	setLevel(options)
	zerolog.MessageFieldName = "msg"
	zerolog.LevelFieldName = "lvl"
	zerolog.TimeFieldFormat = timeFormat

	// 指定日志输出的位置
	writers := make([]io.Writer, 0, 2)
	setConsole(options, &writers)
	diodeCloser = setFile(options, &writers)
	multiWriter := zerolog.MultiLevelWriter(writers...)

	// 设置日志中常用属性
	ctx := zerolog.New(multiWriter).With().Timestamp()
	if options.pid {
		ctx = ctx.Int("pid", os.Getpid())
	}

	if options.caller {
		ctx = ctx.Caller()
	}

	// 修改全局日志
	log.Logger = ctx.Logger()
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
		console := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: timeFormat}
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
