package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

var ErrGraphQLServiceNotAvailable = errors.New("graphql service not available")

type Graph interface {
	ID() string
	Execute(context.Context, *Request, http.RoundTripper) (*Response, error)
}

type GraphConfig struct {
	ServiceURL *url.URL
}

func NewGraph(cfg *GraphConfig) (Graph, error) {
	if cfg.ServiceURL == nil {
		return nil, errors.New("graph config must have a service URL")
	}

	return &graph{
		serviceURL: cfg.ServiceURL,
	}, nil
}

var _ Graph = (*graph)(nil)

type graph struct {
	serviceURL *url.URL
}

func (g *graph) ID() string {
	return g.serviceURL.String()
}

func (g *graph) Execute(ctx context.Context, graphReq *Request, transport http.RoundTripper) (*Response, error) {
	bdy, err := json.Marshal(graphReq)
	if err != nil {
		return nil, fmt.Errorf("failed to encode graphql service request: %s", err)
	}
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		g.serviceURL.String(),
		bytes.NewReader(bdy),
	)
	httpReq.Header.Set("Content-Type", "application/json; charset=utf-8")

	if transport == nil {
		transport = http.DefaultTransport
	}
	httpRes, err := transport.RoundTrip(httpReq)
	if err != nil {
		// TODO: check if this is always the case
		return nil, ErrGraphQLServiceNotAvailable
	}
	if httpRes.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("graphql service responded with error: %s", httpRes.Status)
	}

	gqlRes := new(Response)
	if err := json.NewDecoder(httpRes.Body).Decode(gqlRes); err != nil {
		return nil, fmt.Errorf("failed to decode graphql service response: %s", err)
	}

	return gqlRes, nil
}
