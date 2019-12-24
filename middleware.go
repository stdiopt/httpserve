package httpserve

import "net/http"

type middleware func(http.Handler) http.Handler
type middlewares []middleware

func (mws middlewares) apply(next http.Handler) http.Handler {
	if len(mws) == 0 {
		return next
	}
	last := len(mws) - 1
	return mws[:last].apply(mws[last](next))
	// Change order
	//return mws[1:].apply(mws[0](next))
}
