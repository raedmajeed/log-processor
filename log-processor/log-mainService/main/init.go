package main

import (
	"LOGProcessor/log-mainService/services"
	"LOGProcessor/shared/db"
	"LOGProcessor/shared/types"
	"os"

	"github.com/hibiken/asynq"
	_ "github.com/lib/pq"
)

const (
	baseUrl = "api/"
)

var apiRoutes = types.ApiRoutes{
	{
		Method:           "GET",
		Pattern:          "/upload-logs",
		Handler:          services.HandleUploadFileToQueue,
		IsAuthReq:        true,
		UseRateLimit:     true,
		RateLimitPerSec:  1,
		RateLimitBurst:   2,
		RateLimitMessage: "too many requests",
	},
	{
		Method:           "GET",
		Pattern:          "/queue-status",
		Handler:          services.HandleGetQueueCurrentStatus,
		IsAuthReq:        false,
		UseRateLimit:     true,
		RateLimitPerSec:  1,
		RateLimitBurst:   2,
		RateLimitMessage: "too many requests",
	},
	{
		Method:           "GET",
		Pattern:          "/stats",
		Handler:          services.HandleGetAggregatedTasks,
		IsAuthReq:        true,
		UseRateLimit:     true,
		RateLimitPerSec:  1,
		RateLimitBurst:   2,
		RateLimitMessage: "too many requests",
	},
	{
		Method:           "GET",
		Pattern:          "/stats/:jobId",
		Handler:          services.HandleGetStatsByJobId,
		IsAuthReq:        true,
		UseRateLimit:     true,
		RateLimitPerSec:  1,
		RateLimitBurst:   2,
		RateLimitMessage: "too many requests",
	},
	{
		Method:    "GET",
		Pattern:   "/live-stats",
		Handler:   services.WebSocketHandler,
		IsAuthReq: true,
	},
}

func init() {
	var (
		err error
	)

	loadEnvVariables()
	err = db.InitDbConnection()
	if err != nil {
		return
	}
	createAsynqRedisClient()
	// InitInspector()
}

/******************************************************************************
* FUNCTION:        loadEnvVariables
* DESCRIPTION:     Function to load env variables and assign to global variables
* INPUT:           None
* RETURNS:         VOID
******************************************************************************/
func loadEnvVariables() {
	types.CmnGlblCfg.RUNNING_PORT = getEnv("MAIN_SERVICE_PORT", "0001")
	types.CmnGlblCfg.JWT_SECRET = getEnv("JWT_SECRET", "0001")
	types.CmnGlblCfg.SUPEBASE_API = getEnv("SUPEBASE_API", "0001")
	types.CmnGlblCfg.SUPEBASE_API_KEY = getEnv("SUPEBASE_API_KEY", "0001")
	types.CmnGlblCfg.SUPEBASE_BUCKET = getEnv("SUPEBASE_BUCKET", "0001")
	types.CmnGlblCfg.SUPEBASE_STORAGE_BASE = getEnv("SUPEBASE_STORAGE_BASE", "0001")
	types.CmnGlblCfg.SUPEBASE_REST_BASE = getEnv("SUPEBASE_REST_BASE", "0001")
	types.CmnGlblCfg.DB_USER = getEnv("DB_USER", "0001")
	types.CmnGlblCfg.DB_PASSWORD = getEnv("DB_PASSWORD", "0001")
	types.CmnGlblCfg.DB_DATABASE = getEnv("DB_DATABASE", "0001")
	types.CmnGlblCfg.DB_PORT = getEnv("DB_PORT", "0001")
	types.CmnGlblCfg.DB_HOST = getEnv("DB_HOST", "0001")
	types.CmnGlblCfg.REDIS_ADDR = getEnv("REDIS_ADDR", "0001")
	types.CmnGlblCfg.KEYWORD_CONFIG = getEnv("KEYWORD_CONFIG", "0001")
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

/******************************************************************************
* FUNCTION:        loadEnvVariables
* DESCRIPTION:     Function to load env variables and assign to global variables
* INPUT:           None
* RETURNS:         VOID
******************************************************************************/
func createAsynqRedisClient() {
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: types.CmnGlblCfg.REDIS_ADDR})
	types.AsynqClient.AsynqClient = asynqClient
	if asynqClient == nil {
		os.Exit(0)
	}
}

/******************************************************************************
* FUNCTION:        loadEnvVariables
* DESCRIPTION:     Function to load env variables and assign to global variables
* INPUT:           None
* RETURNS:         VOID
******************************************************************************/
func InitInspector() {
	types.AsynqClient.AsynqInspector = asynq.NewInspector(asynq.RedisClientOpt{Addr: types.CmnGlblCfg.REDIS_ADDR})
	if types.AsynqClient.AsynqInspector == nil {
		os.Exit(0)
	}
}
