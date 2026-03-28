package common

import (
	"bytes"
	"net/http"
)

// HttpClient is an interface for making HTTP Requests.
type HttpClient interface {
	Get(url string) (*http.Response, error)
	Post(uri string, contentType string, body *bytes.Reader) (*http.Response, error)
}

type DefaultHttpClient struct{}

func (client *DefaultHttpClient) Get(url string) (*http.Response, error) {
	return http.Get(url)
}

func (client *DefaultHttpClient) Post(uri string, contentType string, body *bytes.Reader) (*http.Response, error) {
	return http.Post(uri, contentType, body)
}
