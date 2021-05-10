package logger

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	nameFormat = "2006_01_02"
)

func newFileWriter(dir, name, ext string, expireDay int) io.Writer {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil
	}

	writer := &fileWriter{
		dir:       dir,
		name:      name,
		ext:       ext,
		expireDay: expireDay,
	}

	if writer.expireDay > 0 {
		go writer.cleanExpiredFile()
	}

	return writer
}

type fileWriter struct {
	dir          string
	name         string
	ext          string
	markDay      int
	file         *os.File
	fullFileName string
	expireDay    int
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
	f, err := os.OpenFile(w.fullFileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
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

func (w *fileWriter) cleanExpiredFile() {
	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case <-ticker.C:
			rd, err := ioutil.ReadDir(w.dir)
			if err != nil {
				log.Error().Err(err).Str("dir", w.dir).Msg("failed to read directory")
				break
			}

			expiredDate := time.Now().Add(-time.Duration(w.expireDay*24) * time.Hour)
			for _, fileDir := range rd {
				if fileDir.IsDir() {
					continue
				}

				fileName := fileDir.Name()
				if !strings.HasPrefix(fileName, w.name) || !strings.HasSuffix(fileName, w.ext) {
					continue
				}

				// Format: logger_2020_11_29.log
				fileStrTime := strings.TrimRight(fileName, w.ext)
				fileStrTime = strings.TrimLeft(fileStrTime, w.name+"_")
				fileTime, err := time.Parse(nameFormat, fileStrTime)
				if err != nil || fileTime.After(expiredDate) {
					continue
				}

				filePath := path.Join(w.dir, fileName)
				if err = os.Remove(filePath); err != nil {
					log.Error().Err(err).Str("filePath", filePath).Msg("failed to remove file")
				}
			}
		}
	}
}
