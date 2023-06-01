package httpclient

import (
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	timeoutQueryKey             = "httpclient-timeout"
	maxIdleConnsQueryKey        = "httpclient-maxidleconns"
	maxIdleConnsPerHostQueryKey = "httpclient-maxidleconnsperhost"
)

var configurationKeys = [...]string{timeoutQueryKey, maxIdleConnsQueryKey, maxIdleConnsPerHostQueryKey}

const DefaultTimeOutInterval = 210 * time.Second

type HTTPClient struct {
	*http.Client
	Target *url.URL
}

func New(baseURL string) (*HTTPClient, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	return NewWithTransport(baseURL, transport)
}

func NewWithTransport(baseURL string, transport http.RoundTripper) (*HTTPClient, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	httpClient := newDefaultHttpClient(transport)

	queryValues := parsedURL.Query()
	for _, key := range configurationKeys {
		value := queryValues.Get(key)
		if value == "" {
			continue
		}
		switch key {
		case timeoutQueryKey:
			httpClient.Timeout, err = time.ParseDuration(value)
		case maxIdleConnsQueryKey:
			httpClient.Transport.(*http.Transport).MaxIdleConns, err = strconv.Atoi(value)
		case maxIdleConnsPerHostQueryKey:
			httpClient.Transport.(*http.Transport).MaxIdleConnsPerHost, err = strconv.Atoi(value)
		}
		if err != nil {
			return nil, err
		}
	}

	if len(configurationKeys) < 1 {
		httpClient.Timeout = time.Duration(2)*time.Minute + time.Duration(30)*time.Second
	}

	parsedURL = newUrlWithoutConfiguration(parsedURL)

	return &HTTPClient{
		Client: httpClient,
		Target: parsedURL,
	}, nil
}

func newDefaultHttpClient(transport http.RoundTripper) *http.Client {
	return &http.Client{
		Transport: transport,
		Timeout:   DefaultTimeOutInterval,
	}
}

func newUrlWithoutConfiguration(url *url.URL) *url.URL {
	newUrl := url
	queryValues := newUrl.Query()
	for _, key := range configurationKeys {
		delete(queryValues, key)
	}
	newUrl.RawQuery = queryValues.Encode()
	return newUrl
}

func (c *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	req.URL = c.Target.ResolveReference(req.URL)
	urlValues := req.URL.Query()
	mergeValues(urlValues, c.Target.Query())
	req.URL.RawQuery = urlValues.Encode()
	return c.Client.Do(req)
}

func mergeValues(dst, src url.Values) {
	for key, values := range src {
		dst[key] = append(dst[key], values...)
	}
}
