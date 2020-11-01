package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/rs/zerolog/log"
)

const timespan = 1 //1秒检测一次文件变更状态

type watchCallback func(string, *config) bool //检测到文件变更后的回调

var conf *config
var once sync.Once

// Conf ...
type Conf interface {
	AppConf
}

type config struct {
	appConfig
}

// Instance ...
func Instance() Conf {
	once.Do(func() {
		conf = new(config)
		confList := make(map[string]watchCallback)

		//需要检测的文件列表
		confList[serverConfFile] = loadServerConf

		//应用启动后，立即执行一次加载
		for file, fun := range confList {
			if !fun(file, conf) {
				conf = nil
				return
			}
		}

		//定时检测加载
		go checkConfigUpdate(confList)
	})

	return conf
}

func checkConfigUpdate(configList map[string]watchCallback) {
	t := time.NewTicker(timespan * time.Second)
	fileMap := make(map[string]int64)

	defer func() {
		if err := recover(); err != nil {
			log.Error().Str("stackInfo", string(debug.Stack())).Msg("painc and recover")
		}
		t.Stop()
	}()

	for range t.C {
		checkFileStat(configList, fileMap)
	}
}

func checkFileStat(configList map[string]watchCallback, fileMap map[string]int64) {
	for file, fun := range configList {
		if fi, err := os.Stat(file); err == nil {
			curModifyTime := fi.ModTime().Unix()

			if lastModifyTime, ok := fileMap[file]; ok {
				if curModifyTime > lastModifyTime {
					tempConf := *conf
					if fun(file, &tempConf) {
						fileMap[file] = curModifyTime
						// todo 保证业务中使用配置的原子性
						atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&conf)), unsafe.Pointer(&tempConf))
					} else {
						log.Error().Str("confName", file).Msg("load file failed.")
					}
				}
			} else {
				fileMap[file] = curModifyTime
				log.Info().Str("fileName", file).Msg("checkConfigUpdate")
			}
		}
	}
}

func loadConfFromFile(filePath string, i interface{}) bool {
	bs, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Error().Err(err).Str("fileName", filePath).Msg("config read file failed")
		return false
	}

	err = json.Unmarshal(bs, i)
	if err != nil {
		log.Error().Err(err).Str("fileName", filePath).Msg("parse config error")
		return false
	}

	log.Info().Str("path", filePath).Interface("confContext", i).Msg("load config successfully")

	return true
}
