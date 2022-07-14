package main

import (
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"net/http"
	"os"
)

func main() {
	logger := log.NewLogfmtLogger(os.Stderr)

	var svc StringService
	// service
	svc = stringService{}
	// wrap service with logging middleware
	svc = loggingMiddleware{logger, svc}

	uppercaseHandler := httptransport.NewServer(
		makeUppercaseEndpoint(svc), // endpoint
		decodeUppercaseRequest,     // request
		encodeResponse,             // response
	)

	countHandler := httptransport.NewServer(
		makeCountEndpoint(svc), // endpoint
		decodeCountRequest,     // request
		encodeResponse,         // response
	)

	// route
	http.Handle("/uppercase", uppercaseHandler)
	http.Handle("/count", countHandler)
	http.ListenAndServe(":8080", nil)
}
