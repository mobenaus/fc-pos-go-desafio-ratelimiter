package middleware

import (
	"net/http"

	"github.com/mobenaus/fc-pos-go-desafio-ratelimiter/internal/ratelimit"
	"github.com/mobenaus/fc-pos-go-desafio-ratelimiter/internal/util"
)

type RateLimitMiddleWareConfig struct {
	IPRateLimitConfig    ratelimit.RateLimitConfig
	TokenRateLimitConfig ratelimit.RateLimitConfig
}

func NewRateLimitConfig(
	ipconfig ratelimit.RateLimitConfig,
	tokenconfig ratelimit.RateLimitConfig,
) *RateLimitMiddleWareConfig {
	return &RateLimitMiddleWareConfig{
		IPRateLimitConfig:    ipconfig,
		TokenRateLimitConfig: tokenconfig,
	}
}

func (rl *RateLimitMiddleWareConfig) RateLimitMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var limit *ratelimit.RateLimit
			apiKey := r.Header.Get("API_KEY")
			if apiKey != "" {
				limit = rl.TokenRateLimitConfig.GetRateLimit(apiKey)
			} else {
				limit = rl.IPRateLimitConfig.GetRateLimit(util.GetIpFromAddress(r.RemoteAddr))
			}
			if limit.UseToken() != nil {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte("you have reached the maximum number of requests or actions allowed within a certain time frame"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
