package logger

import (
	"os"
	"path/filepath"
	"strings"
)

var (
	defaultOptions = Options{
		dir:     filepath.Dir(os.Args[0]),
		name:    strings.TrimRight(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0])),
		ext:     ".log",
		console: true,
		file:    false,
		level:   "trace",
	}
)

type Options struct {
	dir     string // 日志目录
	name    string // 日志文件名
	ext     string // 后缀名
	console bool   // 开启 console 日志
	file    bool   // 开启文件日志
	level   string // 日志等级
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
