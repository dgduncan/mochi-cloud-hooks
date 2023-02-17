package mochicloudhooks

import "net/http"

// Transport represents everything required for adding to the roundtripper interface
type Transport struct {
	OriginalTransport http.RoundTripper
}

// NewTransport creates a new Transport object with any passed in information
func NewTransport(rt http.RoundTripper) *http.Client {
	if rt == nil {
		rt = &Transport{
			OriginalTransport: http.DefaultTransport,
		}
	}

	return &http.Client{
		Transport: rt,
	}
}

// RoundTrip goes through the HTTP RoundTrip implementation and attempts to add ASAP if not passed it
func (st *Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	return st.OriginalTransport.RoundTrip(r)
}
