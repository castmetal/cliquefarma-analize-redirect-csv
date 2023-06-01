package inputhttp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/template"
	"time"

	"go.uber.org/zap"

	"github.com/castmetal/cliquefarma-analize-redirect-csv/http/httpclient"
	"github.com/castmetal/cliquefarma-analize-redirect-csv/logger"
	"github.com/castmetal/cliquefarma-analize-redirect-csv/metadata"
)

type HTTP struct {
	Client  *httpclient.HTTPClient
	Method  string
	Headers map[string]string
	Query   map[string]string
}

func New(ctx context.Context, meta metadata.Map) (*HTTP, error) {
	targetURL := meta.AsString("targetURL", "")
	if targetURL == "" {
		return nil, errors.New("could not create source with empty target url")
	}
	method := meta.AsString("method", "")
	if method == "" {
		return nil, errors.New("could not create source with empty http method")
	}
	client, err := httpclient.New(targetURL)
	if err != nil {
		return nil, fmt.Errorf("could not create httpclient with this target url: %w", err)
	}

	hdrs := map[string]string{
		"User-Agent": "cliquefarmabot v1.0.0",
	}
	headers := meta.AsMap("headers")
	for k, v := range headers {
		hdrs[k] = fmt.Sprintf("%s", v)
	}

	qs := make(map[string]string)
	// Template applies over the value of the query strings.
	// This allows us to dynamically set some query strings values, allowing
	// us to force a cache bypass, for example.
	// Right now only a now() function is supported, as it generates a unix timestamp
	// that changes everytime now() is called, bypassing caches.
	tmpl := template.New("queryStringTmpl").Funcs(template.FuncMap{
		"now": func() int64 { return time.Now().UnixNano() },
	})
	query := meta.AsMap("query")
	for k, v := range query {
		var builder strings.Builder

		t, err := tmpl.Parse(fmt.Sprintf("%s", v))
		if err != nil {
			return nil, fmt.Errorf("could not create http client because of invalid query string template: %w", err)
		}
		if err := t.Execute(&builder, ""); err != nil {
			return nil, fmt.Errorf("could not execute query string template: %w", err)
		}

		qs[k] = builder.String()
	}

	return &HTTP{
		Client:  client,
		Method:  method,
		Headers: hdrs,
		Query:   qs,
	}, nil
}

func (h *HTTP) Do(req *http.Request) (*http.Response, error) {
	for k, v := range h.Headers {
		req.Header.Set(k, v)
	}
	query := req.URL.Query()
	for k, v := range h.Query {
		query.Set(k, v)
	}
	req.URL.RawQuery = query.Encode()
	return h.Client.Do(req)
}

func (h *HTTP) Data(ctx context.Context) (chan io.ReadCloser, error) {
	out := make(chan io.ReadCloser, 1)
	defer close(out)
	req, err := http.NewRequestWithContext(ctx, h.Method, h.Client.Target.Path, nil)
	if err != nil {
		return nil, fmt.Errorf("input/http: could not create new request: %w", err)
	}
	res, err := h.Do(req)
	if err != nil {
		return nil, fmt.Errorf("input/http: could not complete request: %w", err)
	}
	fields := []zap.Field{
		zap.Int("statusCode", res.StatusCode),
		zap.String("targetURL", h.Client.Target.String()),
		zap.Int64("contentLength", res.ContentLength),
	}

	if res.StatusCode > 299 {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, res.Body)
		if err != nil {
			logger.Error(ctx, err, "could not read body when status code > 299")
		}

		logger.Error(ctx, errors.New(
			"we have received a unexpected status code",
		), "input/http: unexpected status code received", append(fields, zap.ByteString("body", buf.Bytes()))...)

		// modify - todo no return here as we know if there is any error parser/formatter will catch it

		return nil, fmt.Errorf("could not read body when status code > 299")
	}

	logger.Info(ctx, "input/http: request completed successfully", fields...)

	out <- res.Body
	return out, nil
}
