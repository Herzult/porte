package playground

import (
	"context"
	"net/http"

	"github.com/herzult/porte/internal/graph"
	"github.com/herzult/porte/internal/graph/proxy"
)

// NewProxyPlugin returns a new proxy plugin configured to render a GraphQL
// playground when requests to the Proxy are made from a web browser.
func NewProxyPlugin() (*proxy.Plugin, error) {
	return &proxy.Plugin{
		InitContext: func(ctx context.Context) context.Context {
			return context.WithValue(
				ctx,
				proxyPluginState{},
				&proxyPluginState{},
			)
		},
		ReadProxyRequest: func(inner proxy.ReadProxyRequest) proxy.ReadProxyRequest {
			return func(r *http.Request) (*graph.Request, error) {
				if isPlaygroundRequest(r) {
					state, _ := r.Context().Value(proxyPluginState{}).(*proxyPluginState)
					state.active = true
					state.path = r.URL.Path
					state.host = r.Host
					return nil, nil
				}
				return inner(r)
			}
		},
		WriteProxyResponse: func(inner proxy.WriteProxyResponse) proxy.WriteProxyResponse {
			return func(ctx context.Context, w http.ResponseWriter, graphRes *graph.Response, graphErr error) {
				state, _ := ctx.Value(proxyPluginState{}).(*proxyPluginState)
				if state.active {
					renderPlayground(w, state.path, state.host)
					return
				}
				inner(ctx, w, graphRes, graphErr)
			}
		},
	}, nil
}

type proxyPluginState struct {
	active bool
	path   string
	host   string
}
