package http_helper

import (
	"bytes"
	"context"
	"net/http"
)

type HttpHeader map[string]string

type HttpClient interface {
	DoGet(ctx context.Context, headerValues HttpHeader, url string) (*http.Response, error)
	DoPost(ctx context.Context, headerValues HttpHeader, url string, body []byte) (*http.Response, error)
}

type HttpCaller struct {
	client *http.Client
}

func NewHTTPCaller() HttpClient {
	return &HttpCaller{
		client: &http.Client{Transport: http.DefaultTransport},
	}
}

func (r *HttpCaller) DoGet(ctx context.Context, headerValues HttpHeader, url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return nil, err
	}

	for key, value := range headerValues {
		req.Header.Set(key, value)
	}
	return r.client.Do(req.WithContext(ctx))
}

func (r *HttpCaller) DoPost(ctx context.Context, headerValues HttpHeader, url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))

	if err != nil {
		return nil, err
	}

	for key, value := range headerValues {
		req.Header.Set(key, value)
	}
	return r.client.Do(req.WithContext(ctx))
}
