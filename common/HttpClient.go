package common

import "net/http"

// HttpClient is an interface for making HTTP Requests.
type HttpClient interface {
	Get(url string) (*http.Response, error)
}

type DefaultHttpClient struct{}

func (client *DefaultHttpClient) Get(url string) (*http.Response, error) {
	return http.Get(url)
}
