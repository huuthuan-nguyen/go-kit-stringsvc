package main

import (
	"errors"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
)

// proxymw implements StringService, forwarding Uppercase requests to the
// provided endpoint, and serving all other (i.e Count) requests via the
// next StringService.
type proxymw struct {
	next      StringService     // serve most requests via this service...
	uppercase endpoint.Endpoint // ...except Uppercase, which gets served by this endpoint
}

func (mw proxymw) Uppercase(s string) (string, error) {
	response, err := mw.uppercase(nil, uppercaseRequest{S: s})
	if err != nil {
		return "", err
	}
	resp := response.(uppercaseResponse)
	if resp.Err != "" {
		return resp.V, errors.New(resp.Err)
	}
	return resp.V, nil
}

func proxyingMiddleware(proxyURL string) ServiceMiddleware {
	return func(next StringService) StringService {
		return proxymw{next, makeUppercaseEndpoint(proxyURL)}
	}
}

func makeUppercaseProxy(proxyURL string) endpoint.Endpoint {
	return httptransport.NewClient(
		"GET",
		mustParseURL(proxyURL),
		encodeUppercaseRequest,
		decodeUppercaseResponse,
	).Endpoint()
}
