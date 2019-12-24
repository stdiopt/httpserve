package httpserve

import (
	"net/http"
)

// Middleware type for middleware chain
type Middleware func(http.Handler) http.Handler

// Middlewares chain
type Middlewares []Middleware

// Apply middlewares and return the final handler
func (mws Middlewares) Apply(next http.Handler) http.Handler {
	if len(mws) == 0 {
		return next
	}
	last := len(mws) - 1
	return mws[:last].Apply(mws[last](next))
	// Change order
	//return mws[1:].apply(mws[0](next))
}
