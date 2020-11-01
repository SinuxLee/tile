package app

import (
	"fmt"
	"net/http"
	"reflect"
	"tile/pkg/logger"
	"tile/pkg/middleware/ginx"
	"time"

	"tile/internal/config"
	"tile/internal/controller"
	httpCtrl "tile/internal/controller/http"
	"tile/internal/service"
	"tile/internal/store"
	"tile/pkg/nid"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// Option ...
type Option func(*app) error

// ServiceName ...
func ServiceName(name string) Option {
	return func(a *app) (err error) {
		a.serviceName = name
		return
	}
}

func ServiceID(ID int) Option {
	return func(a *app) (err error) {
		a.serviceID = ID
		return
	}
}

// Conf ...
func Conf() Option {
	return func(a *app) (err error) {
		a.conf = config.Instance()
		value := reflect.ValueOf(a.conf)
		if value.IsNil() {
			err = errors.New("config init failed")
			return
		}

		return
	}
}

// LogLevel ...
func LogLevel() Option {
	return func(a *app) (err error) {
		return logger.SetLevel(a.conf.GetLogLevel())
	}
}

// UseCase ...
func UseCase() Option {
	return func(a *app) (err error) {
		a.useCase = service.NewUseCase(a.dao)
		if a.useCase == nil {
			return errors.New("create UseCase failed")
		}

		return
	}
}

// Controller ...
func Controller() Option {
	return func(a *app) (err error) {
		a.ctrl = httpCtrl.NewHttpController(a.useCase)
		if a.ctrl == nil {
			return errors.New("create Controller failed")
		}

		return
	}
}

// Dao ...
func Dao() Option {
	return func(a *app) (err error) {
		a.dao = store.NewDao(a.named)
		if a.dao == nil {
			return errors.New("create dao failed")
		}
		return
	}
}

// Router ...
func Router() Option {
	return func(a *app) (err error) {
		isDebug := a.conf.IsDebugMode()
		if isDebug {
			a.router = gin.Default()
		} else {
			gin.SetMode(gin.ReleaseMode)
			a.router = gin.New()
			a.router.Use(gin.Recovery(), cors.Default())
		}
		a.router.Use(ginx.NewRateLimiter(time.Second, 5000))
		if a.router == nil {
			return errors.New("gin router is nil")
		}

		//健康检查
		a.router.Any("/health", func(ctx *gin.Context) {
			ctx.String(http.StatusOK, "It is OK\n")
		})

		controller.RegisterHandler(a.router, a.ctrl, isDebug)

		return
	}
}

// PProf ...
func PProf() Option {
	return func(a *app) (err error) {
		pprof.Register(a.router)
		return
	}
}

// HTTPServer ...
func HTTPServer() Option {
	return func(a *app) (err error) {
		a.httpSrv = &http.Server{
			Addr:    fmt.Sprintf(":%d", a.conf.GetHTTPPort()),
			Handler: a.router,
		}
		//web.NewService() //使用micro的web
		return
	}
}

func Named() Option {
	return func(a *app) (err error) {
		a.named, err = nid.NewConsulNamed(a.conf.GetConsulAddr())
		if err != nil {
			return err
		}
		return nil
	}
}
