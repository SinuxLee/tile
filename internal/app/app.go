package app

import (
	"net"
	"net/http"
	"os"

	"tile/internal/config"
	"tile/internal/controller"
	"tile/internal/service"
	"tile/internal/store"
	"tile/pkg/logger"
	"tile/pkg/nid"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// App ...
type App interface {
	Run(chan os.Signal) error
	Stop() error
}

// New ...
func New(options ...Option) (App, error) {
	logger.Init(logger.Console(false),
		logger.File(true))

	svc := &app{}

	// init app component
	for _, opt := range options {
		if err := opt(svc); err != nil {
			return nil, err
		}
	}

	return svc, nil
}

type app struct {
	serviceName string
	serviceID   int
	nodeID      int
	localIP     string
	router      *gin.Engine
	httpSrv     *http.Server
	conf        config.Conf
	ctrl        controller.Controller
	useCase     service.UseCase
	dao         store.Dao
	named       nid.NodeNamed
}

func (s *app) GetServiceID() int {
	return s.serviceID
}

func (s *app) GetLocalIP() string {
	if s.localIP == "" {
		s.localIP = s.intranetIP()
	}
	return s.localIP
}

// Run ...
func (s *app) Run(ch chan os.Signal) error {
	// Run server
	go func() {
		if err := s.httpSrv.ListenAndServe(); err != nil {
			log.Error().Err(err).Msg("http app exit")
			close(ch)
		}
	}()

	return nil
}

// Stop ...
func (s *app) Stop() error {
	return nil
}

// intranetIP 找到第一个10、172、192开头的ip
func (s *app) intranetIP() (ip string) {
	addr, err := net.InterfaceAddrs()
	if err != nil {
		return
	}
	for _, addr := range addr {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}

	return
}
