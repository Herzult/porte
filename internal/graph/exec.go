package graph

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type Request struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	Extensions    map[string]interface{} `json:"extensions,omitempty"`
}

var ErrHTTPMethodNotAllowed = errors.New("HTTP method not allowed for graph request")

func NewRequestFromHTTP(r *http.Request) (*Request, error) {
	if r.Method != http.MethodPost {
		return nil, ErrHTTPMethodNotAllowed
	}

	gr := new(Request)
	if err := json.NewDecoder(r.Body).Decode(gr); err != nil {
		return nil, fmt.Errorf("Failed to decode graph request from HTTP request's body: %s", err)
	}

	return gr, nil
}

type Response struct {
	Data       interface{}            `json:"data,omitempty"`
	Errors     []*Error               `json:"errors,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

func (r *Response) SetExtension(key string, val interface{}) {
	if r.Extensions == nil {
		r.Extensions = make(map[string]interface{})
	}
	r.Extensions[key] = val
}

type Error struct {
	Message   string `json:"message"`
	Locations []*struct {
		Line   int64 `json:"line"`
		Column int64 `json:"column"`
	} `json:"locations,omitempty"`
	Path []interface{} `json:"path,omitempty"`
}
