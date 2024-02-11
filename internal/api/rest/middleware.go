package rest

import (
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"strconv"
	"time"
)

func Metrics(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		wrr := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()
		defer func() {
			HttpRequestLatency.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(wrr.Status())).Observe(float64(time.Since(start).Milliseconds()))
		}()
		h.ServeHTTP(wrr, r)
	}

	return http.HandlerFunc(fn)
}
