package responders

import (
	"net/http"

	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

type RateLimitResponder struct {
}

func (r *RateLimitResponder) ServeHTTP(w http.ResponseWriter, req *http.Request, next caddyhttp.Handler) error {
	req.Header.Set("X-RateLimit-Apply", "true")

	// Continue with the handler chain
	return next.ServeHTTP(w, req)
}
