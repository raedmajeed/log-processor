package types

import (
	"LOGProcessor/shared/models"
	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

type ServiceApiRoute struct {
	Method    string
	Pattern   string
	Handler   gin.HandlerFunc
	IsAuthReq bool
}

type Data interface {
}

type ApiResponse struct {
	Status     int    `json:"status"`
	Message    string `json:"message"`
	DbRecCount int64  `json:"dbRecCount"`
	Data       Data   `json:"data"`
}

var (
	Db          DbHandler
	AsynqClient AsynqHdlr
)

type DbHandler struct {
	DbConn *sql.DB
}

type AsynqHdlr struct {
	AsynqClient    *asynq.Client
	AsynqInspector *asynq.Inspector
}

type ApiRoutes []ServiceApiRoute

var (
	ExitChan   chan error
	CmnGlblCfg models.SvcConfig
)
