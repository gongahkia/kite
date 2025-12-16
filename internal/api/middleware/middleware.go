package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gongahkia/kite/internal/observability"
	"github.com/gongahkia/kite/pkg/errors"
)

// RequestID adds a unique request ID to each request
func RequestID() fiber.Handler {
	return requestid.New(requestid.Config{
		Header: "X-Request-ID",
		Generator: func() string {
			return time.Now().Format("20060102150405") + "-" + randString(8)
		},
	})
}

// Logger logs each request
func Logger(logger *observability.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		requestID := c.GetRespHeader("X-Request-ID")

		// Process request
		err := c.Next()

		// Log request
		duration := time.Since(start)
		status := c.Response().StatusCode()

		logger.WithFields(map[string]interface{}{
			"request_id": requestID,
			"method":     c.Method(),
			"path":       c.Path(),
			"status":     status,
			"duration":   duration.Milliseconds(),
			"ip":         c.IP(),
		}).Infof("%s %s %d %dms", c.Method(), c.Path(), status, duration.Milliseconds())

		return err
	}
}

// CORS handles Cross-Origin Resource Sharing
func CORS() fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization,X-Request-ID",
		MaxAge:       300,
	})
}

// Recovery recovers from panics
func Recovery(logger *observability.Logger) fiber.Handler {
	return recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			logger.WithField("panic", e).Error("Panic recovered")
		},
	})
}

// Metrics records metrics for each request
func Metrics(metrics *observability.Metrics) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Increment in-flight requests
		metrics.HTTPRequestsInFlight.Inc()
		defer metrics.HTTPRequestsInFlight.Dec()

		// Process request
		err := c.Next()

		// Record metrics
		duration := time.Since(start)
		status := c.Response().StatusCode()
		method := c.Method()
		path := c.Path()

		metrics.RecordHTTPRequest(method, path, fiber.StatusMessage(status), duration)

		return err
	}
}

// ErrorHandler is a custom error handler
func ErrorHandler(logger *observability.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		message := "Internal Server Error"

		// Check for Fiber errors
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
			message = e.Message
		}

		// Check for custom Kite errors
		if kiteErr, ok := err.(*errors.KiteError); ok {
			message = kiteErr.Message

			switch kiteErr.Code {
			case "NOT_FOUND":
				code = fiber.StatusNotFound
			case "VALIDATION_ERROR":
				code = fiber.StatusBadRequest
			case "AUTH_ERROR":
				code = fiber.StatusUnauthorized
			default:
				code = fiber.StatusInternalServerError
			}
		}

		// Log error
		requestID := c.GetRespHeader("X-Request-ID")
		logger.WithFields(map[string]interface{}{
			"request_id": requestID,
			"method":     c.Method(),
			"path":       c.Path(),
			"error":      err.Error(),
		}).Error(message)

		// Send error response
		return c.Status(code).JSON(fiber.Map{
			"error":      message,
			"request_id": requestID,
			"path":       c.Path(),
		})
	}
}

// randString generates a random string
func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
