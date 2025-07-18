package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mobenaus/fc-pos-go-desafio-ratelimiter/internal/util"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%v Request received: %s %s from %s\n", time.Now().Format(time.RFC3339Nano), r.Method, r.URL.Path, util.GetIpFromAddress(r.RemoteAddr))
		next.ServeHTTP(w, r)
	})
}
