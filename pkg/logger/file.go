package logger

import (
	"fmt"
	"io"
	"os"
	"path"
	"time"
)

const (
	nameFormat = "2006_01_02"
)

func newFileWriter(dir, name, ext string) io.Writer {
	return &fileWriter{
		dir:  dir,
		name: name,
		ext:  ext,
	}
}

type fileWriter struct {
	dir          string
	name         string
	ext          string
	markDay      int
	file         *os.File
	fullFileName string
}

func (w *fileWriter) Write(data []byte) (n int, err error) {
	if f := w.getFile(); f != nil {
		return f.Write(data)
	}

	return 0, nil
}

func (w *fileWriter) Close() error {
	f := w.getFile()
	if f != nil {
		f.Sync()
		f.Close()
	}
	return nil
}

func (w *fileWriter) getFile() *os.File {
	now := time.Now()
	markDay := now.YearDay()
	if w.markDay != markDay {
		w.markDay = markDay
		w.fullFileName = w.fileName(&now)
	}
	if w.file != nil && w.isExist(w.fullFileName) {
		return w.file
	}
	f, err := os.OpenFile(w.fullFileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("cacheWriter create file error:%v \n", err)
		return nil
	}
	w.file, f = f, w.file
	if f != nil {
		go func() {
			f.Sync()
			f.Close()
		}()
	}
	return w.file
}

func (w *fileWriter) fileName(tm *time.Time) string {
	return path.Join(w.dir, fmt.Sprintf("%v_%v", w.name, tm.Format(nameFormat)+w.ext))
}

func (w *fileWriter) isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}
