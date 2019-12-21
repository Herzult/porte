package execlog

import (
	"context"
	"net/http"
	"time"

	"github.com/herzult/porte/internal/graph"
	"github.com/herzult/porte/internal/graph/proxy"
)

// NewProxyPlugin returns a new proxy plugin instance configured to write
// every execution of the graph using the provided ExecLogEntryWriter.
func NewProxyPlugin(entryWriter EntryWriter) (*proxy.Plugin, error) {
	return &proxy.Plugin{
		InitContext: func(ctx context.Context) context.Context {
			graph := proxy.GetGraph(ctx)
			return context.WithValue(ctx, stateKey{}, &Entry{
				ID:      proxy.GetExecID(ctx),
				GraphID: graph.ID(),
			})
		},
		ReadProxyRequest: func(next proxy.ReadProxyRequest) proxy.ReadProxyRequest {
			return func(req *http.Request) (*graph.Request, error) {
				entry, _ := req.Context().Value(stateKey{}).(*Entry)
				entry.ClientName = req.Header.Get("Client-Name")
				entry.ClientVersion = req.Header.Get("Client-Version")

				nextReq, nextErr := next(req)
				if nextReq != nil {
					entry.Request = nextReq
				}
				return nextReq, nextErr
			}
		},
		SendGraphRequest: func(next http.RoundTripper) http.RoundTripper {
			return proxy.SendGraphRequest(func(req *http.Request) (*http.Response, error) {
				entry, _ := req.Context().Value(stateKey{}).(*Entry)
				entry.StartTime = time.Now()
				nextRes, nextErr := next.RoundTrip(req)
				entry.Duration = time.Since(entry.StartTime)
				return nextRes, nextErr
			})
		},
		WriteProxyResponse: func(next proxy.WriteProxyResponse) proxy.WriteProxyResponse {
			return func(ctx context.Context, w http.ResponseWriter, graphRes *graph.Response, graphErr error) {
				next(ctx, w, graphRes, graphErr)
				entry, _ := ctx.Value(stateKey{}).(*Entry)
				entry.Response = graphRes
				if graphErr != nil {
					entry.Error = graphErr.Error()
				}
				entryWriter.Write(entry)
			}
		},
	}, nil
}

type stateKey struct{}
