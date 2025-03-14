package main

import (
	"LOGProcessor/shared/types"
	"os"
)

func init() {
	//* SET UP supebase setup
	//* SET UP Logging conf

	loadEnvVariables()
}

func loadEnvVariables() {
	types.CmnGlblCfg.RUNNING_PORT = getEnv("MAIN_SERVICE_PORT", "0001")
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
