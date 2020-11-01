package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type nodeRequest struct {
	LocalPath  string `json:"path"`
	InternalIP string `json:"ip"`
}

func (c *ControllerOnHttp) GetNodeID(ctx *gin.Context) {
	service := ctx.Param("serverName")
	if service == "" {
		c.ResponseWithCode(ctx, CodeLackParam)
		return
	}

	var internalIp, localPath string
	if ctx.Request.Method == http.MethodGet {
		localPath = ctx.Query("path")
		internalIp = ctx.Query("ip")
	} else if ctx.Request.Method == http.MethodPost {
		req := &nodeRequest{}
		err := ctx.ShouldBind(req)
		if err != nil {
			c.ResponseWithCode(ctx, CodeInvalidParam)
			return
		}
		localPath = req.LocalPath
		internalIp = req.InternalIP
	}

	if internalIp == "" {
		c.ResponseWithCode(ctx, CodeLackParam)
		return
	}

	id, err := c.useCase.GetNodeID(localPath, internalIp, service)
	if err != nil {
		c.ResponseWithDesc(ctx, CodeNodeID, err.Error())
		return
	}

	c.ResponseWithData(ctx, gin.H{"nodeId": id})
}
