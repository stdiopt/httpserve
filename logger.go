package httpserve

import (
	"bufio"
	"log"
	"net"
	"net/http"
)

// LogHelper struct to handle write logs
type LogHelper struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader hijack write header to track httpStatus
func (l *LogHelper) WriteHeader(code int) {
	l.statusCode = code
	l.ResponseWriter.WriteHeader(code)
}

// Hijack hihack wrapper for hijacker users (websocket?)
func (l *LogHelper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker := l.ResponseWriter.(http.Hijacker)
	return hijacker.Hijack()
}

// Logger middleware
func Logger(log *log.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := &LogHelper{w, 200}
			if next != nil {
				next.ServeHTTP(l, r)
			}
			raddr := r.RemoteAddr
			log.Printf("(%s) %s %s - [%d %s]", raddr, r.Method, r.URL.Path, l.statusCode, http.StatusText(l.statusCode))
		})
	}
}
