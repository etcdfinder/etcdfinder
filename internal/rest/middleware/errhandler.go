package middleware

import (
	"encoding/json"
	"maps"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/etcdfinder/etcdfinder/internal/customerrors"
	"github.com/gin-gonic/gin"
)

// ErrorHandler middleware handles error responses
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			// Get display message from hints
			display := getDisplayMessage(err)

			// Get safe details
			details := getSafeDetails(err)

			response := customerrors.ErrorResponse{
				Success: false,
				Error: customerrors.ErrorDetail{
					Display:       display,
					InternalError: err.Error(),
					Details:       details,
				},
			}

			status := customerrors.HTTPStatusFromErr(err)
			c.JSON(status, response)
		}
	}
}

func getDisplayMessage(err error) string {
	if hints := errors.GetAllHints(err); len(hints) > 0 {
		// Get the first non-empty hint - GetAllHints is post-order traversal
		for _, hint := range hints {
			if hint = strings.TrimSpace(hint); hint != "" {
				return hint
			}
		}
	}

	// fallback to the error message
	return "An unexpected error occurred"
}

func getSafeDetails(err error) map[string]any {
	details := make(map[string]any)

	allSafeDetails := errors.GetAllSafeDetails(err)
	for _, sdp := range allSafeDetails {
		if len(sdp.SafeDetails) == 0 {
			continue
		}

		for _, payload := range sdp.SafeDetails {
			if len(payload) > 9 && strings.HasPrefix(payload, "__json__:") {
				jsonStr := payload[9:]
				var jsonDetails map[string]any
				if err := json.Unmarshal([]byte(jsonStr), &jsonDetails); err == nil {
					maps.Copy(details, jsonDetails)
				}
			}
		}
	}

	return details
}
