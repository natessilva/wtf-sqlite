package sqlite

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)
)

func init() {
	prometheus.MustRegister(requestDuration)
}

type instrumentedResponseWriter struct {
	http.ResponseWriter
	statusCode int
	path       string
}

func (rec *instrumentedResponseWriter) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

func (rec *instrumentedResponseWriter) WritePath(path string) {
	rec.path = path
}

func instrumentedHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &instrumentedResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		path := r.URL.Path
		timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
			if rec.path != "" {
				path = rec.path
			}
			requestDuration.WithLabelValues(r.Method, path, http.StatusText(rec.statusCode)).Observe(v)
		}))
		defer timer.ObserveDuration()
		handler.ServeHTTP(rec, r)
	})
}

type InstrumentedRouter struct {
	*httprouter.Router
}

func NewInstrumentedRouter() *InstrumentedRouter {
	return &InstrumentedRouter{
		Router: httprouter.New(),
	}
}

func (r *InstrumentedRouter) GET(path string, handle httprouter.Handle) {
	r.Handle(http.MethodGet, path, handle)
}

func (r *InstrumentedRouter) HEAD(path string, handle httprouter.Handle) {
	r.Handle(http.MethodHead, path, handle)
}

func (r *InstrumentedRouter) OPTIONS(path string, handle httprouter.Handle) {
	r.Handle(http.MethodOptions, path, handle)
}

func (r *InstrumentedRouter) POST(path string, handle httprouter.Handle) {
	r.Handle(http.MethodPost, path, handle)
}

func (r *InstrumentedRouter) PUT(path string, handle httprouter.Handle) {
	r.Handle(http.MethodPut, path, handle)
}

func (r *InstrumentedRouter) PATCH(path string, handle httprouter.Handle) {
	r.Handle(http.MethodPatch, path, handle)
}

func (r *InstrumentedRouter) DELETE(path string, handle httprouter.Handle) {
	r.Handle(http.MethodDelete, path, handle)
}

func (r *InstrumentedRouter) Handle(method, path string, handle httprouter.Handle) {
	r.Router.HandlerFunc(method, path, func(w http.ResponseWriter, r *http.Request) {
		if rec, ok := w.(*instrumentedResponseWriter); ok {
			rec.WritePath(path)
		}
		handle(w, r, httprouter.ParamsFromContext(r.Context()))
	})
}

func (r *InstrumentedRouter) Handler(method, path string, handler http.Handler) {
	r.Handle(method, path,
		func(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
			handler.ServeHTTP(w, req)
		},
	)
}

func (r *InstrumentedRouter) HandlerFunc(method, path string, handler http.HandlerFunc) {
	r.Handler(method, path, handler)
}
