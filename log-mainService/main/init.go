package main

import (
	"log-mainService/services"
	"shared/db"
	"shared/types"
	"os"

	"github.com/hibiken/asynq"
	_ "github.com/lib/pq"
)

const (
	baseUrl = "api/"
)

var apiRoutes = types.ApiRoutes{
	{
		Method:    "GET",
		Pattern:   "/upload-logs",
		Handler:   services.HandleUploadFileToQueue,
		IsAuthReq: true,
	},
	{
		Method:    "GET",
		Pattern:   "/queue-status",
		Handler:   services.HandleGetQueueCurrentStatus,
		IsAuthReq: false,
	},
	{
		Method:    "GET",
		Pattern:   "/stats",
		Handler:   services.HandleGetAggregatedTasks,
		IsAuthReq: true,
	},
	{
		Method:    "GET",
		Pattern:   "/stats/:jobId",
		Handler:   services.HandleGetStatsByJobId,
		IsAuthReq: true,
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
}

/******************************************************************************
* FUNCTION:        loadEnvVariables
* DESCRIPTION:     Function to load env variables and assign to global variables
* INPUT:           None
* RETURNS:         VOID
******************************************************************************/
func loadEnvVariables() {
	types.CmnGlblCfg.RUNNING_PORT = getEnv("MAIN_SERVICE_PORT", "")
	types.CmnGlblCfg.JWT_SECRET = getEnv("JWT_SECRET", "")
	types.CmnGlblCfg.SUPEBASE_API = getEnv("SUPEBASE_API", "")
	types.CmnGlblCfg.SUPEBASE_API_KEY = getEnv("SUPEBASE_API_KEY", "")
	types.CmnGlblCfg.SUPEBASE_BUCKET = getEnv("SUPEBASE_BUCKET", "")
	types.CmnGlblCfg.SUPEBASE_STORAGE_BASE = getEnv("SUPEBASE_STORAGE_BASE", "")
	types.CmnGlblCfg.SUPEBASE_REST_BASE = getEnv("SUPEBASE_REST_BASE", "")
	types.CmnGlblCfg.DB_USER = getEnv("DB_USER", "")
	types.CmnGlblCfg.DB_PASSWORD = getEnv("DB_PASSWORD", "")
	types.CmnGlblCfg.DB_DATABASE = getEnv("DB_DATABASE", "")
	types.CmnGlblCfg.DB_PORT = getEnv("DB_PORT", "")
	types.CmnGlblCfg.DB_HOST = getEnv("DB_HOST", "")
	types.CmnGlblCfg.REDIS_ADDR = getEnv("REDIS_ADDR", "")
	types.CmnGlblCfg.KEYWORD_CONFIG = getEnv("KEYWORD_CONFIG", "")
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

/******************************************************************************
* FUNCTION:        createAsynqRedisClient
* DESCRIPTION:     Initializes a Asynq Redis Client
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
* FUNCTION:        InitInspector
* DESCRIPTION:     Initializes  a new Asynq task inspector for queue monitoring
* INPUT:           None
* RETURNS:         VOID
******************************************************************************/
func InitInspector() {
	types.AsynqClient.AsynqInspector = asynq.NewInspector(asynq.RedisClientOpt{Addr: types.CmnGlblCfg.REDIS_ADDR})
	if types.AsynqClient.AsynqInspector == nil {
		os.Exit(0)
	}
}
