package main

import (
	"LOGProcessor/log-mainService/services"
	"LOGProcessor/shared/types"
	"fmt"
	"os"
	"os/signal"

	"github.com/gin-gonic/gin"
)

const (
	baseUrl = "api/"
)

var apiRoutes = types.ApiRoutes{
	{
		Method:    "GET",
		Pattern:   "/test-endpoint",
		Handler:   services.TestService,
		IsAuthReq: false,
	},
}

/******************************************************************************
* FUNCTION:        main
* DESCRIPTION:     Entry point for the application. Starts the Gin router and
*                  listens for system signals.
* INPUT:           None
* RETURNS:         VOID
******************************************************************************/
func main() {
	router := createNewRouter()
	go router.Run(":" + types.CmnGlblCfg.RUNNING_PORT)

	go signalHandler()
	err := <-types.ExitChan

	// !REPLACE WITH LOGGER HERE
	fmt.Println("error", err)
}

/******************************************************************************
* FUNCTION:         createNewRouter
*
* DESCRIPTION:      Initializes a new Gin router and sets up API routes.
* INPUT:          	None
* RETURNS:          *gin.Engine - Configured Gin router
******************************************************************************/
func createNewRouter() *gin.Engine {
	r := gin.Default()

	for _, route := range apiRoutes {
		endpoint := baseUrl + route.Pattern
		switch route.Method {
		case "GET":
			r.GET(endpoint, route.Handler)
		case "POST":
			r.POST(endpoint, route.Handler)
		case "PUT":
			r.PUT(endpoint, route.Handler)
		case "DELETE":
			r.DELETE(endpoint, route.Handler)
		default:
			panic("Unsupported HTTP method: " + route.Method)
		}
	}

	return r
}

/******************************************************************************
* FUNCTION:        signalHandler
*
* DESCRIPTION:     Listens for OS interrupt signals and gracefully shuts down
*                 the application.
* INPUT:           None
* RETURNS:         VOID
******************************************************************************/
func signalHandler() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	sig := <-sigChan
	types.ExitChan <- fmt.Errorf("%+v signal", sig)
}
