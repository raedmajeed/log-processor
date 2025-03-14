package types

import (
	"LOGProcessor/shared/models"

	"github.com/gin-gonic/gin"
)

type ServiceApiRoute struct {
	Method    string
	Pattern   string
	Handler   gin.HandlerFunc
	IsAuthReq bool
}

type ApiRoutes []ServiceApiRoute

var (
	ExitChan   chan error
	CmnGlblCfg models.SvcConfig
)
