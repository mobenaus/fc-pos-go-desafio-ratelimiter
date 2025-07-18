package middleware

import (
	"net/http"

	"github.com/mobenaus/fc-pos-go-desafio-ratelimiter/internal/ratelimit"
	"github.com/mobenaus/fc-pos-go-desafio-ratelimiter/internal/util"
)

type RateLimitMiddleWare struct {
	IPRateLimitConfig    *ratelimit.RateLimit
	TokenRateLimitConfig *ratelimit.RateLimit
}

func NewRateLimit(
	ipconfig *ratelimit.RateLimit,
	tokenconfig *ratelimit.RateLimit,
) *RateLimitMiddleWare {
	return &RateLimitMiddleWare{
		IPRateLimitConfig:    ipconfig,
		TokenRateLimitConfig: tokenconfig,
	}
}

func (rl *RateLimitMiddleWare) RateLimitMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var error error
			apiKey := r.Header.Get("API_KEY")
			if apiKey != "" {
				error = rl.TokenRateLimitConfig.UseToken(apiKey)
			} else {
				error = rl.IPRateLimitConfig.UseToken(util.GetIpFromAddress(r.RemoteAddr))
			}
			if error != nil {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(error.Error()))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
