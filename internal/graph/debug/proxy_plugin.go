package debug

import (
	"context"
	"net/http"
	"net/http/httputil"

	"github.com/herzult/porte/internal/graph"
	"github.com/herzult/porte/internal/graph/proxy"
)

type debugState struct {
	GraphErr string `json:"graphError,omitempty"`
	GraphReq string `json:"graphRequest,omitemtpy"`
	GraphRes string `json:"graphResponse,omitempty"`
}

func NewProxyPlugin() (*proxy.Plugin, error) {
	return &proxy.Plugin{
		InitContext: func(ctx context.Context) context.Context {
			return context.WithValue(ctx, debugState{}, &debugState{})
		},
		ReadProxyRequest: func(next proxy.ReadProxyRequest) proxy.ReadProxyRequest {
			return func(r *http.Request) (*graph.Request, error) {
				return next(r)
			}
		},
		SendGraphRequest: func(next http.RoundTripper) http.RoundTripper {
			return proxy.SendGraphRequest(func(req *http.Request) (*http.Response, error) {
				ds := req.Context().Value(debugState{}).(*debugState)
				reqDump, _ := httputil.DumpRequest(req, true)
				ds.GraphReq = string(reqDump)
				res, err := next.RoundTrip(req)
				if res != nil {
					resDump, _ := httputil.DumpResponse(res, true)
					ds.GraphRes = string(resDump)
				}
				if err != nil {
					ds.GraphErr = err.Error()
				}
				return res, err
			})
		},
		WriteProxyResponse: func(next proxy.WriteProxyResponse) proxy.WriteProxyResponse {
			return func(ctx context.Context, w http.ResponseWriter, graphRes *graph.Response, graphErr error) {
				ds := ctx.Value(debugState{}).(*debugState)
				graphRes.SetExtension("debug", ds)
				next(ctx, w, graphRes, graphErr)
			}
		},
	}, nil
}
