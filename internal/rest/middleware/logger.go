package middleware

import (
	"net/http"
	"time"

	"github.com/etcdfinder/etcdfinder/pkg/logger"
	"github.com/gin-gonic/gin"
)

// LoggerMiddleware logs details about each request
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		if len(c.Errors) > 0 {
			// Skip error logging for 404 responses
			if c.Writer.Status() != http.StatusNotFound {
				for _, e := range c.Errors.Errors() {
					logger.Errorf(e)
				}
			}
		} else {
			logger.Debugf("%s %s %s %d %s",
				c.Request.Method,
				path,
				query,
				c.Writer.Status(),
				latency,
			)
		}
	}
}
