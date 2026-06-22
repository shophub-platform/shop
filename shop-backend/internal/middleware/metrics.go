package middleware

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shophub/shop/internal/metrics"
)

var (
	visitorMu    sync.RWMutex
	seenVisitors = make(map[string]struct{})
)

// PrometheusMetrics records HTTP metrics for each request.
// Must be registered after chi's routing so RoutePattern() is populated.
func PrometheusMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		iw := &instrumentedWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(iw, r)

		// chi.RouteContext gives the normalized pattern (e.g. /api/v1/items/{id})
		// which prevents high-cardinality labels from raw UUIDs in the path.
		pattern := routePattern(r)
		dur := time.Since(start).Seconds()
		statusStr := strconv.Itoa(iw.status)

		metrics.HTTPRequestsTotal.WithLabelValues(r.Method, pattern, statusStr).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(r.Method, pattern).Observe(dur)
		metrics.HTTPResponseBytesTotal.WithLabelValues(r.Method, pattern).Add(float64(iw.bytesWritten))

		trackVisitor(r)
	})
}

func routePattern(r *http.Request) string {
	rctx := chi.RouteContext(r.Context())
	if rctx != nil && rctx.RoutePattern() != "" {
		return rctx.RoutePattern()
	}
	return r.URL.Path
}

// trackVisitor increments HTTPUniqueVisitorsTotal once per IP+UA+day combination.
func trackVisitor(r *http.Request) {
	ip := strings.Split(r.RemoteAddr, ":")[0]
	ua := r.Header.Get("User-Agent")
	day := time.Now().UTC().Format("2006-01-02")
	raw := fmt.Sprintf("%s|%s|%s", ip, ua, day)
	key := fmt.Sprintf("%x", sha256.Sum256([]byte(raw)))

	visitorMu.RLock()
	_, seen := seenVisitors[key]
	visitorMu.RUnlock()

	if !seen {
		visitorMu.Lock()
		if _, seen = seenVisitors[key]; !seen {
			seenVisitors[key] = struct{}{}
			metrics.HTTPUniqueVisitorsTotal.Inc()
		}
		visitorMu.Unlock()
	}
}

// instrumentedWriter captures status code and bytes written so the metrics
// middleware can record them after the handler returns.
type instrumentedWriter struct {
	http.ResponseWriter
	status       int
	bytesWritten int
}

func (iw *instrumentedWriter) WriteHeader(status int) {
	iw.status = status
	iw.ResponseWriter.WriteHeader(status)
}

func (iw *instrumentedWriter) Write(b []byte) (int, error) {
	n, err := iw.ResponseWriter.Write(b)
	iw.bytesWritten += n
	return n, err
}
