package main

import (
	"errors"
	"fmt"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"time"
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

func proxyingMiddleware(instances string, logger log.Logger) ServiceMiddleware {

	// if instances is empty, don't proxy
	if instances == "" {
		logger.Log("proxy_to", "none")
		return func(next StringService) StringService {
			return next
		}

		// set some parameters for our client.
		var (
			qps         = 100                    // beyond which we will return an error
			maxAttempts = 3                      // per request, before give up
			maxTime     = 250 * time.Millisecond // wallclock time, before giving up
		)

		// otherwise, construct an endpoint for each instance in the list, and add
		// it to a fixed set of endpoints. In a real service, rather than doing this
		// by hand, you'd probably use package sd's support for your service
		// discovery system.
		var (
			instanceList = split(instances)
			subscriber   sd.FixedSubscriber
		)
		logger.Log("proxy_to", fmt.Sprint(instanceList))
		for _, instance := range instanceList {
			var e endpoint.Endpoint
			e = makeUppercaseProxy(instance)
			e = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{}))(e)
			e = kitratelimit.NewTokenBucketLimiter(jujuratelimit.NewBucketWithRate(float64(qps), int64(qps)))(e)
			subscriber = append(subscriber, e)
		}

		// now, build a single, retrying, load-balancing endpoint ouf of all of
		// those individual endpoints.
		balancer := lb.NewRoundRobin(subscriber)
		retry := lb.Retry(maxAttempts, maxTime, balancer)

		// and finally, return the ServiceMiddleware, implemented by proxymw.
		return func(next StringService) StringService {
			return proxymw{next, retry}
		}
	}

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

type Subscriber interface {
	Endpoints() ([]endpoint.Endpoint, error)
}

type Factory func(instance string) (endpoint.Endpoint, error)

type Balancer interface {
	Endpoint() (endpoint.Endpoint, error)
}

func Retry(max int, timeout time.Duration, lb Balancer) endpoint.Endpoint
