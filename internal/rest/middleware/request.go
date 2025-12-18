package middleware

import (
	"context"

	"github.com/etcdfinder/etcdfinder/internal/lib"
	"github.com/gin-gonic/gin"
)

func RequestIDMiddleware(c *gin.Context) {
	// Create a new context from the request context
	ctx := c.Request.Context()

	// Add request ID
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = lib.GenerateUUID()
	}

	// Create new context with values
	ctx = context.WithValue(ctx, lib.CtxRequestID, requestID)

	// Replace request context
	c.Request = c.Request.WithContext(ctx)

	// Add headers for response
	c.Header(lib.HeaderRequestID, requestID)

	c.Next()
}
