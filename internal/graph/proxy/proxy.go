package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/herzult/porte/internal/graph"
)

type Proxy interface {
	http.Handler
}

type Config struct {
	Graph   graph.Graph
	Plugins []*Plugin
}

type InitContext func(context.Context) context.Context
type ReadProxyRequest func(*http.Request) (*graph.Request, error)
type SendGraphRequest func(*http.Request) (*http.Response, error)
type WriteProxyResponse func(context.Context, http.ResponseWriter, *graph.Response, error)

func (f SendGraphRequest) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

type Plugin struct {
	InitContext        InitContext
	ReadProxyRequest   func(ReadProxyRequest) ReadProxyRequest
	SendGraphRequest   func(http.RoundTripper) http.RoundTripper
	WriteProxyResponse func(WriteProxyResponse) WriteProxyResponse
}

type graphKey struct{}
type execIDKey struct{}

func GetGraph(ctx context.Context) graph.Graph {
	return ctx.Value(graphKey{}).(graph.Graph)
}

func GetExecID(ctx context.Context) string {
	return ctx.Value(execIDKey{}).(string)
}

func New(cfg *Config) (Proxy, error) {
	initContext := func(ctx context.Context) context.Context {
		ctx = context.WithValue(ctx, graphKey{}, cfg.Graph)
		ctx = context.WithValue(ctx, execIDKey{}, uuid.Must(uuid.NewRandom()).String())
		for _, plugin := range cfg.Plugins {
			if plugin.InitContext != nil {
				ctx = plugin.InitContext(ctx)
			}
		}
		return ctx
	}
	readProxyRequest := graph.NewRequestFromHTTP
	sendGraphRequest := http.DefaultTransport
	writeProxyResponse := defaultWriteProxyResponse

	for _, plugin := range cfg.Plugins {
		if plugin.ReadProxyRequest != nil {
			readProxyRequest = plugin.ReadProxyRequest(readProxyRequest)
		}
		if plugin.SendGraphRequest != nil {
			sendGraphRequest = plugin.SendGraphRequest(sendGraphRequest)
		}
		if plugin.WriteProxyResponse != nil {
			writeProxyResponse = plugin.WriteProxyResponse(writeProxyResponse)
		}
	}

	return &proxy{
		initContext:        initContext,
		readProxyRequest:   readProxyRequest,
		writeProxyResponse: writeProxyResponse,
		graph:              cfg.Graph,
		graphTransport:     sendGraphRequest,
	}, nil
}

type proxy struct {
	initContext        InitContext
	readProxyRequest   ReadProxyRequest
	writeProxyResponse WriteProxyResponse
	graph              graph.Graph
	graphTransport     http.RoundTripper
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r = r.WithContext(p.initContext(r.Context()))

	graphReq, err := p.readProxyRequest(r)
	if err == graph.ErrHTTPMethodNotAllowed {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Method not allowed: %s", err)
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Bad request: %s", err)
		return
	}

	if nil == graphReq {
		p.writeProxyResponse(r.Context(), w, nil, nil)
		return
	}

	graphRes, graphErr := p.graph.Execute(
		r.Context(),
		graphReq,
		forwardHeadersToGraph(p.graphTransport, r.Header),
	)

	if graphRes == nil && graphErr != nil {
		graphRes = &graph.Response{
			Errors: []*graph.Error{
				&graph.Error{
					Message: "Failed to execute graph request.",
				},
			},
		}
	}

	p.writeProxyResponse(r.Context(), w, graphRes, graphErr)
}

func forwardHeadersToGraph(next http.RoundTripper, head http.Header) http.RoundTripper {
	return SendGraphRequest(func(req *http.Request) (*http.Response, error) {
		req.Header = head.Clone()
		// Remove hop-by-hop headers to the backend. Especially
		// important is "Connection" because we want a persistent
		// connection, regardless of what the client sent to us.
		for _, h := range hopHeaders {
			hv := req.Header.Get(h)
			if h == "Te" && hv == "trailers" {
				continue
			}
			req.Header.Del(h)
		}
		// removes hop-by-hop headers listed in the "Connection" header of h.
		// See RFC 7230, section 6.1
		for _, f := range req.Header["Connection"] {
			for _, sf := range strings.Split(f, ",") {
				if sf = strings.TrimSpace(sf); sf != "" {
					req.Header.Del(sf)
				}
			}
		}

		return next.RoundTrip(req)
	})
}

var hopHeaders = []string{
	"Connection",
	"Proxy-Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",
	"Trailer",
	"Transfer-Encoding",
	"Upgrade",
}

func defaultWriteProxyResponse(_ context.Context, w http.ResponseWriter, graphRes *graph.Response, graphErr error) {
	if graphRes == nil && graphErr == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err := json.NewEncoder(w).Encode(graphRes)
	if err != nil {
		log.Println("Failed to write back graph response:", err.Error())
	}
}
