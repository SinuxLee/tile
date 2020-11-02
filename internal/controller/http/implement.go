package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/sinuxlee/tile/internal/controller"
	"github.com/sinuxlee/tile/internal/service"
)

// Response ...
type Response struct {
	ErrCode int    `json:"errCode"`
	ErrDesc string `json:"errDesc"`
}

func NewHttpController(uc service.UseCase) controller.Controller {
	return &ControllerOnHttp{
		useCase: uc,
	}
}

type ControllerOnHttp struct {
	useCase service.UseCase
}

// ResponseWithData ...
func (c *ControllerOnHttp) ResponseWithData(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, gin.H{"errCode": 0, "errDesc": "success", "data": data})
}

// ResponseWithCode ...
func (c *ControllerOnHttp) ResponseWithCode(ctx *gin.Context, code int) {
	resp := &Response{ErrCode: code}
	desc, ok := codeText[code]
	if ok {
		resp.ErrDesc = desc
	} else {
		resp.ErrDesc = "unknown codeText"
	}

	ctx.JSON(http.StatusOK, resp)
	if code != CodeSuccess {
		c.ErrorLog(ctx, resp)
	}
}

// ResponseWithDesc ...
func (c *ControllerOnHttp) ResponseWithDesc(ctx *gin.Context, code int, desc string) {
	resp := &Response{
		ErrCode: code,
		ErrDesc: desc,
	}

	ctx.JSON(http.StatusOK, resp)
	if code != CodeSuccess {
		c.ErrorLog(ctx, resp)
	}
}

func (c *ControllerOnHttp) ErrorLog(ctx *gin.Context, resp *Response) {
	raw, _ := ctx.GetRawData()
	log.Error().Str("path", ctx.Request.URL.Path).
		Str("query", ctx.Request.URL.RawQuery).
		Str("request", string(raw)).
		Interface("response", resp).Send()
}
