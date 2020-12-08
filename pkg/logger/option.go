package logger

import (
	"os"
	"path/filepath"
	"strings"
)

var (
	defaultOptions = Options{
		dir:       filepath.Dir(os.Args[0]),
		name:      strings.TrimRight(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0])),
		ext:       ".log",
		console:   true,
		file:      false,
		level:     "trace",
		expireDay: 7,
	}
)

type Options struct {
	dir       string // 日志目录
	name      string // 日志文件名
	ext       string // 后缀名
	level     string // 日志等级
	console   bool   // 开启 console 日志
	file      bool   // 开启文件日志
	pid       bool   // 显示进程id
	caller    bool   // 显示源文件及行数
	expireDay int    // 日志文件保留天数
}

// Option ...
type Option func(*Options)

// Dir ...
func Dir(dir string) Option {
	return func(o *Options) {
		o.dir = dir
	}
}

// Name ...
func Name(name string) Option {
	return func(o *Options) {
		o.name = name
	}
}

// Ext ...
func Ext(ext string) Option {
	return func(o *Options) {
		o.ext = ext
	}
}

func Console(open bool) Option {
	return func(o *Options) {
		o.console = open
	}
}

func File(open bool) Option {
	return func(o *Options) {
		o.file = open
	}
}

func Pid(open bool) Option {
	return func(o *Options) {
		o.pid = open
	}
}

func Caller(open bool) Option {
	return func(o *Options) {
		o.caller = open
	}
}

func ExpireDay(expireDay int) Option {
	return func(o *Options) {
		o.expireDay = expireDay
	}
}
