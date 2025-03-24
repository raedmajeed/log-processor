package main

import (
	"LOGProcessor/log-mainService/services"
	"LOGProcessor/log-mainService/tasks"
	"LOGProcessor/shared/types"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/hibiken/asynq"
	"golang.org/x/time/rate"
)

/******************************************************************************
* FUNCTION:        main
* DESCRIPTION:     Entry point for the application. Starts the Gin router and
*                  listens for system signals.
* INPUT:           None
* RETURNS:         VOID
******************************************************************************/
func main() {
	sigChan := make(chan os.Signal, 1)
	router := createNewRouter()
	go router.Run(":" + types.CmnGlblCfg.RUNNING_PORT)
	go runMuxAsynqServer()
	initRateLimitOptions()
	InitInspector()

	go signalHandler(sigChan)
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

	if types.RateLimit.Enabled {
		r.Use(services.GlobalRateLimit(rate.Limit(types.RateLimit.GlobalLimit), types.RateLimit.GlobalBurst))
	}

	ipLimitOpts := types.IPRateLimitOptions{
		ClientTimeout: 5 * time.Minute,
	}
	ipRateLimiter := services.NewRateLimiterMiddleware(rate.Limit(5), 10, ipLimitOpts)
	r.Use(ipRateLimiter.RateLimit())

	for _, route := range apiRoutes {
		// routeHandlers := []gin.HandlerFunc{}

		if route.IsAuthReq {
			r.Use(AuthMiddleware)
		}

		// if route.UseRateLimit {
		// 	routeHandlers = append(routeHandlers, services.RouteSpecificRateLimit(
		// 		route.Pattern,
		// 		rate.Limit(route.RateLimitPerSec),
		// 		route.RateLimitBurst,
		// 	))
		// }

		// routeHandlers = append(routeHandlers, route.Handler)

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
* FUNCTION:        AuthMiddleware
*
* DESCRIPTION:     Middleware function that authorizes users and restricts
*									 unauthorized users
* INPUT:           None
* RETURNS:         VOID
******************************************************************************/
func AuthMiddleware(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
		c.Abort()
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(types.CmnGlblCfg.JWT_SECRET), nil
	})

	if err != nil || !token.Valid {
		fmt.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		c.Abort()
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if claims["aud"] != "authenticated" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			c.Abort()
			return
		}
		c.Set("user_id", claims["sub"])
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
		c.Abort()
		return
	}

	c.Set("token", tokenString)
	c.Next()
}

/******************************************************************************
* FUNCTION:        signalHandler
*
* DESCRIPTION:     Listens for OS interrupt signals and gracefully shuts down
*                 the application.
* INPUT:           None
* RETURNS:         VOID
******************************************************************************/
func signalHandler(sigChan chan os.Signal) {
	signal.Notify(sigChan, os.Interrupt)
	sig := <-sigChan
	types.ExitChan <- fmt.Errorf("%+v signal", sig)
}

/******************************************************************************
* FUNCTION:        runMuxAsynqServer
*
* DESCRIPTION:     runs mux server for handling asynq queue
* INPUT:           None
* RETURNS:         VOID
******************************************************************************/
func runMuxAsynqServer() {
	srv := asynq.NewServer(asynq.RedisClientOpt{Addr: types.CmnGlblCfg.REDIS_ADDR},
		asynq.Config{
			Concurrency: 4,
			Queues: map[string]int{
				"high": 3,
				"low":  1,
			},
		})
	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeLogProcess, services.HandleAsyncTaskMethod)
	fmt.Println("muxx servr started")
	if err := srv.Run(mux); err != nil {
		fmt.Println("ERR", err)
	}
	srv.Shutdown()
}

/******************************************************************************
* FUNCTION:        initRateLimitOptions
* DESCRIPTION:     Initialize rate limit options
* INPUT:           None
* RETURNS:         VOID
******************************************************************************/

func initRateLimitOptions() {
	types.RateLimit = types.RateLimitConfig{
		Enabled:     true,
		GlobalLimit: 100,
		GlobalBurst: 200,
		PerIPLimit:  10,
		PerIPBurst:  20,
		PerRouteOpts: map[string]types.PerRouteLimit{
			"/upload-logs": {
				Limit: 5,
				Burst: 10,
			},
		},
	}
}
