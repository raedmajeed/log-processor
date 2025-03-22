package types

import (
	"LOGProcessor/shared/models"
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

type ServiceApiRoute struct {
	Method           string
	Pattern          string
	Handler          gin.HandlerFunc
	IsAuthReq        bool
	UseRateLimit     bool
	RateLimitPerSec  float64
	RateLimitBurst   int
	RateLimitMessage string
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
	RateLimit  RateLimitConfig
)

type PerRouteLimit struct {
	Limit float64
	Burst int
}

type RateLimitConfig struct {
	Enabled      bool                     `json:"enabled"`
	GlobalLimit  float64                  `json:"globalLimit"`
	GlobalBurst  int                      `json:"globalBurst"`
	PerIPLimit   float64                  `json:"perIPLimit"`
	PerIPBurst   int                      `json:"perIPBurst"`
	PerRouteOpts map[string]PerRouteLimit `json:"perRouteOpts"`
}

type IPRateLimitOptions struct {
	ClientTimeout time.Duration
}
