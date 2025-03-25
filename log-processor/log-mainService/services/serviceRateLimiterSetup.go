// services/rateLimit.go

package services

import (
	"LOGProcessor/shared/types"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ClientRateLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiterMiddleware struct {
	clients         map[string]*ClientRateLimiter
	rateLimit       rate.Limit
	burstLimit      int
	ipLimitOpts     types.IPRateLimitOptions
	mu              sync.Mutex
	cleanupInterval time.Duration
}

/******************************************************************************
* FUNCTION:        NewRateLimiterMiddleware
*
* DESCRIPTION:     This function is used initalize RateLimiterMiddleware
* INPUT:
* RETURNS:
******************************************************************************/
func NewRateLimiterMiddleware(rateLimit rate.Limit, burstLimit int, ipLimitOpts types.IPRateLimitOptions) *RateLimiterMiddleware {
	limiter := &RateLimiterMiddleware{
		clients:         make(map[string]*ClientRateLimiter),
		rateLimit:       rateLimit,
		burstLimit:      burstLimit,
		ipLimitOpts:     ipLimitOpts,
		cleanupInterval: 5 * time.Minute,
	}

	go limiter.cleanupExpiredClients()

	return limiter
}

/******************************************************************************
* FUNCTION:        cleanupExpiredClients
*
* DESCRIPTION:     This method is used to clear the clients aftercleanup time
* INPUT:					 gin context
* RETURNS:         void
******************************************************************************/
func (m *RateLimiterMiddleware) cleanupExpiredClients() {
	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		for ip, client := range m.clients {
			if time.Since(client.lastSeen) > m.ipLimitOpts.ClientTimeout {
				delete(m.clients, ip)
				log.Printf("Removed expired rate limiter for client: %s", ip)
			}
		}
		m.mu.Unlock()
	}
}

/******************************************************************************
* FUNCTION:        getLimiter
*
* DESCRIPTION:     This function is used to get limiter
* INPUT:
* RETURNS:
******************************************************************************/
func (m *RateLimiterMiddleware) getLimiter(clientIP string) *rate.Limiter {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, exists := m.clients[clientIP]
	if !exists {
		client = &ClientRateLimiter{
			limiter:  rate.NewLimiter(m.rateLimit, m.burstLimit),
			lastSeen: time.Now(),
		}
		m.clients[clientIP] = client
		log.Printf("Created new rate limiter for client: %s", clientIP)
	} else {
		client.lastSeen = time.Now()
	}

	return client.limiter
}

/******************************************************************************
* FUNCTION:        RateLimit
*
* DESCRIPTION:     This function is used to limit api hit
* INPUT:
* RETURNS:
******************************************************************************/
func (m *RateLimiterMiddleware) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		limiter := m.getLimiter(clientIP)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"status":  http.StatusTooManyRequests,
				"message": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			log.Printf("Rate limit exceeded for client: %s", clientIP)
			return
		}
		c.Next()
	}
}

/******************************************************************************
* FUNCTION:        GlobalRateLimit
*
* DESCRIPTION:     This function is used to set global rate limit
* INPUT:					 gin context
* RETURNS:         void
******************************************************************************/
func GlobalRateLimit(limit rate.Limit, burst int) gin.HandlerFunc {
	limiter := rate.NewLimiter(limit, burst)
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"status":  http.StatusTooManyRequests,
				"message": "Server is currently handling too many requests. Please try again later.",
			})
			c.Abort()
			log.Printf("Global rate limit exceeded. Request from: %s", c.ClientIP())
			return
		}
		c.Next()
	}
}

/******************************************************************************
* FUNCTION:        RouteSpecificRateLimit
*
* DESCRIPTION:     This function is used to set route specific limit
* INPUT:					 gin context
* RETURNS:         void
******************************************************************************/
func RouteSpecificRateLimit(routeName string, limit rate.Limit, burst int) gin.HandlerFunc {
	limiters := make(map[string]*rate.Limiter)
	var mu sync.Mutex

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("%s:%s", routeName, clientIP)

		mu.Lock()
		limiter, exists := limiters[key]
		if !exists {
			limiter = rate.NewLimiter(limit, burst)
			limiters[key] = limiter
		}
		mu.Unlock()

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"status":  http.StatusTooManyRequests,
				"message": fmt.Sprintf("Rate limit exceeded for route %s. Please try again later.", routeName),
			})
			c.Abort()
			log.Printf("Route-specific rate limit exceeded for route: %s, client: %s", routeName, clientIP)
			return
		}

		c.Next()
	}
}
