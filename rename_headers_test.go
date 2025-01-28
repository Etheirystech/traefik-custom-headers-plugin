package traefik_custom_headers_plugin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServeHTTP(t *testing.T) {
	tests := []struct {
		desc          string
		reqHeader     http.Header
		expReqHeader  http.Header
	}{
		{
			desc: "Should copy CF-Connecting-IP to X-Forwarded-For",
			reqHeader: map[string][]string{
				"CF-Connecting-IP": {"203.0.113.42"},
			},
			expReqHeader: map[string][]string{
				"X-Forwarded-For": {"203.0.113.42"},
			},
		},
		{
			desc: "Should not modify X-Forwarded-For if CF-Connecting-IP is missing",
			reqHeader: map[string][]string{
				"X-Forwarded-For": {"198.51.100.24"},
			},
			expReqHeader: map[string][]string{
				"X-Forwarded-For": {"198.51.100.24"},
			},
		},
		{
			desc: "Should overwrite X-Forwarded-For with CF-Connecting-IP if both exist",
			reqHeader: map[string][]string{
				"CF-Connecting-IP": {"203.0.113.42"},
				"X-Forwarded-For":  {"198.51.100.24"},
			},
			expReqHeader: map[string][]string{
				"X-Forwarded-For": {"203.0.113.42"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			// Create the plugin configuration
			config := &Config{}

			// Create the next handler
			next := func(rw http.ResponseWriter, req *http.Request) {
				// Pass the request to the next handler without modifying it
				for k, v := range req.Header {
					for _, h := range v {
						rw.Header().Add(k, h)
					}
				}
				rw.WriteHeader(http.StatusOK)
			}

			// Create the plugin instance
			plugin, err := New(context.Background(), http.HandlerFunc(next), config, "customHeader")
			if err != nil {
				t.Fatal(err)
			}

			// Create a test request
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			// Add the test headers to the request
			for k, v := range test.reqHeader {
				for _, h := range v {
					req.Header.Add(k, h)
				}
			}

			// Pass the request through the plugin
			plugin.ServeHTTP(recorder, req)

			// Check the resulting request headers
			for k, expected := range test.expReqHeader {
				values := req.Header.Values(k)

				if !testEq(values, expected) {
					t.Errorf("Header mismatch for %q: expected %+v, got %+v", k, expected, values)
				}
			}
		})
	}
}

// Helper function to compare two slices
func testEq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
