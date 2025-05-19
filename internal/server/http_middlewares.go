package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type LoggerMiddleware struct {
	next http.Handler
	out  io.Writer
}

func NewLoggerMiddleware(next http.Handler, out io.Writer) *LoggerMiddleware {
	// set os.Stdout by default
	if out == nil {
		out = os.Stdout
	}
	return &LoggerMiddleware{next, out}
}

func (l LoggerMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	l.next.ServeHTTP(w, r)

	fmt.Fprintf(l.out, "[%s]: %s %s\n", r.Method, r.URL, time.Since(startTime).String())
}
