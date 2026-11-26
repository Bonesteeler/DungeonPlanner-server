package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

const (
	authRateLimitBurst          = 10
	authRateLimitRetryAfterSecs = 6
	authRateLimitExpiry         = 15 * time.Minute
)

var authRateLimit = rate.Every(6 * time.Second)

type RateLimiterConfig struct {
	Store             echomiddleware.RateLimiterStore
	RetryAfterSeconds int
}

func NewRateLimiter() echo.MiddlewareFunc {
	return NewRateLimiterWithConfig(RateLimiterConfig{
		Store: echomiddleware.NewRateLimiterMemoryStoreWithConfig(
			echomiddleware.RateLimiterMemoryStoreConfig{
				Rate:      authRateLimit,
				Burst:     authRateLimitBurst,
				ExpiresIn: authRateLimitExpiry,
			},
		),
		RetryAfterSeconds: authRateLimitRetryAfterSecs,
	})
}

func NewRateLimiterWithConfig(config RateLimiterConfig) echo.MiddlewareFunc {
	if config.RetryAfterSeconds <= 0 {
		config.RetryAfterSeconds = authRateLimitRetryAfterSecs
	}

	return echomiddleware.RateLimiterWithConfig(echomiddleware.RateLimiterConfig{
		Store: config.Store,
		IdentifierExtractor: func(c echo.Context) (string, error) {
			return c.RealIP(), nil
		},
		DenyHandler: func(c echo.Context, _ string, _ error) error {
			c.Response().Header().Set(echo.HeaderRetryAfter, strconv.Itoa(config.RetryAfterSeconds))
			return c.JSON(http.StatusTooManyRequests, struct {
				Error string `json:"error"`
			}{Error: "Too many requests"})
		},
	})
}
