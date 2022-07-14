package main

import (
	"errors"
	httptransport "github.com/go-kit/kit/transport/http"
	"log"
	"net/http"
)

var ErrEmpty = errors.New("empty string")

func main() {

	// make new service
	svc := stringService{}

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
	log.Fatal(http.ListenAndServe(":8080", nil))
}
