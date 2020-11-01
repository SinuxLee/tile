package controller

import (
	"github.com/gin-gonic/gin"
)

type Controller interface {
	GetNodeID(*gin.Context)
}

func RegisterHandler(engine *gin.Engine, ctrl Controller, debugMode bool) {
	group1 := engine.Group("/named/v1")
	group1.GET("/:serverName/nodeid", ctrl.GetNodeID)
	group1.POST("/:serverName/nodeid", ctrl.GetNodeID)
}
