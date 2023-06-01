package inputhttp_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	inputhttp "github.com/castmetal/cliquefarma-analize-redirect-csv/http"
	"github.com/castmetal/cliquefarma-analize-redirect-csv/metadata"
)

func TestHTTPInput(t *testing.T) {
	testCases := []struct {
		desc string

		errAssertionFunc require.ErrorAssertionFunc
		metadata         metadata.Map
		assertRequest    func(t *testing.T) *httptest.Server
	}{
		{
			desc:             "new http input doing request",
			errAssertionFunc: require.NoError,
			assertRequest: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					expectedQueryString := url.Values{
						"hello": []string{"world"},
					}
					expectedHeaders := http.Header{
						"Content-Type":    []string{"application/json"},
						"Accept-Encoding": []string{"gzip"},
						"User-Agent":      []string{"cliquefarmabot v1.0.0"},
					}

					require.Equal(t, expectedQueryString, r.URL.Query())
					require.Equal(t, expectedHeaders, r.Header)
					fmt.Fprint(w, "OK")
				}))
			},
			metadata: metadata.Map{
				"method": "get",
				"headers": map[string]interface{}{
					"content-type": "application/json",
				},
				"query": map[string]interface{}{
					"hello": "world",
				},
			},
		},
		{
			desc:             "new http input doing request",
			errAssertionFunc: require.NoError,
			assertRequest: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					expectedQueryString := url.Values{
						"hello": []string{"world"},
					}
					expectedHeaders := http.Header{
						"Content-Type":    []string{"application/json"},
						"Accept-Encoding": []string{"gzip"},
						"User-Agent":      []string{"cliquefarmabot v1.0.0"},
					}

					require.Equal(t, expectedQueryString, r.URL.Query())
					require.Equal(t, expectedHeaders, r.Header)
					fmt.Fprint(w, "OK")
				}))
			},
			metadata: metadata.Map{
				"method": "get",
				"headers": map[string]interface{}{
					"content-type": "application/json",
				},
				"query": map[string]interface{}{
					"hello": "world",
				},
			},
		},
		{
			desc:             "new http input doing request, using now() generated time",
			errAssertionFunc: require.NoError,
			assertRequest: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					expectedHeaders := http.Header{
						"Content-Type":    []string{"application/json"},
						"Accept-Encoding": []string{"gzip"},
						"User-Agent":      []string{"cliquefarmabot v1.0.0"},
					}

					qs := r.URL.Query()
					log.Println(qs.Get("t"))
					require.Equal(t, "world", qs.Get("hello"))
					require.NotEmpty(t, qs.Get("t"))
					require.Equal(t, expectedHeaders, r.Header)
					fmt.Fprint(w, "OK")
				}))
			},
			metadata: metadata.Map{
				"method": "get",
				"headers": map[string]interface{}{
					"content-type": "application/json",
				},
				"query": map[string]interface{}{
					"hello": "world",
					"t":     "{{ now }}",
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			testServer := tC.assertRequest(t)
			defer testServer.Close()

			tC.metadata["targetURL"] = testServer.URL
			client, err := inputhttp.New(context.Background(), tC.metadata)
			tC.errAssertionFunc(t, err)
			require.NotNil(t, client)

			data, err := client.Data(context.Background())
			require.NoError(t, err)

			var res []byte
			for msg := range data {

				func() {
					defer msg.Close()
					res, err = ioutil.ReadAll(msg)
					require.NoError(t, err)
				}()
			}
			require.Equal(t, "OK", string(res))
		})
	}
}
