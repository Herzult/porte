package prometheus

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/herzult/porte/internal/graph"
	"github.com/herzult/porte/internal/graph/proxy"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ProxyPluginConfig defines the configuration of the proxy plugin
type ProxyPluginConfig struct {
	Namespace string
	Subsystem string
}

// NewProxyPlugin returns a new proxy instance that records metrics using
// prometheus.
func NewProxyPlugin(cfg ProxyPluginConfig) (*proxy.Plugin, error) {

	graphGraphqlErrorsTotal := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: cfg.Namespace,
		Subsystem: cfg.Subsystem,
		Name:      "graphql_errors_total",
	})

	graphHTTPRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: cfg.Namespace,
			Subsystem: cfg.Subsystem,
			Name:      "graph_http_requests_total",
		},
		[]string{"code", "method"},
	)

	graphHTTPRequestDuration := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: cfg.Namespace,
			Subsystem: cfg.Subsystem,
			Name:      "graph_http_request_duration_seconds",
		},
		[]string{"code", "method"},
	)

	graphHTTPRequestsInFlight := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: cfg.Namespace,
			Subsystem: cfg.Subsystem,
			Name:      "graph_http_requests_in_flight",
		},
	)

	prometheus.MustRegister(
		graphHTTPRequestsTotal,
		graphHTTPRequestDuration,
		graphHTTPRequestsInFlight,
	)

	return &proxy.Plugin{
		InitContext: func(ctx context.Context) context.Context {
			return ctx
		},
		ReadProxyRequest: func(next proxy.ReadProxyRequest) proxy.ReadProxyRequest {
			return func(r *http.Request) (*graph.Request, error) {
				return next(r)
			}
		},
		SendGraphRequest: func(next http.RoundTripper) http.RoundTripper {
			next = proxy.SendGraphRequest(promhttp.InstrumentRoundTripperCounter(
				graphHTTPRequestsTotal,
				next,
			))

			next = proxy.SendGraphRequest(promhttp.InstrumentRoundTripperDuration(
				graphHTTPRequestDuration,
				next,
			))

			next = proxy.SendGraphRequest(promhttp.InstrumentRoundTripperInFlight(
				graphHTTPRequestsInFlight,
				next,
			))

			return next
		},
		WriteProxyResponse: func(next proxy.WriteProxyResponse) proxy.WriteProxyResponse {
			return func(ctx context.Context, w http.ResponseWriter, graphRes *graph.Response, graphErr error) {
				if graphRes != nil {
					graphGraphqlErrorsTotal.Add(float64(len(graphRes.Errors)))
				}

				next(ctx, w, graphRes, graphErr)
			}
		},
	}, nil
}
