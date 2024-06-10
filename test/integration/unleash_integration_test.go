package unleash_integration

import (
	"context"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"net/http"
	"net/http/httptest"
	"testing"
	unleash "traefik_unleash"
	fixture "traefik_unleash/test"
)

func TestMain(m *testing.M) {
	fixture.Setup(m)
}

func TestIntegrationRewriteHostAndHeadersAndPathWhenToggleIsActiveUsingUserId(t *testing.T) {
	conf := `url: "http://localhost:4242/api/"
app: "test-app"
interval: 10
metrics:
  interval: 10
toggles:
  - feature: "test-toggle-user-id"
    path:
      value: "/foo"
      rewrite: "/bar"
    host:
      value: "localhost"
      rewrite: "whoami2"
    headers:
      - key: "X-Foo"
        value: "Bar"
        context: "request"
      - key: "X-Served-By"
        value: "whoami2"
        context: "response"`

	expectedRequestHeaders := map[string]string{
		"X-Foo": "Bar",
	}
	expectedResponseHeaders := map[string]string{
		"X-Served-By": "whoami2",
	}
	testIntegrationRewrite(t, conf, "http://localhost/foo", "12345", "whoami2", "/bar", expectedRequestHeaders, expectedResponseHeaders)
}

func TestIntegrationRewriteHostAndPathWhenToggleIsActive(t *testing.T) {
	conf := `
url: "http://localhost:4242/api/"
app: "test-app"
interval: 10
metrics:
  interval: 10
toggles:
  - feature: "test-toggle"
    path:
      value: "/bar"
      rewrite: "/foo"
    host:
      value: "localhost"
      rewrite: "whoami2"
`
	testIntegrationRewrite(t, conf, "http://localhost/bar", "", "whoami2", "/foo", nil, nil)
}

func TestIntegrationRewritePathWhenToggleIsActive(t *testing.T) {
	conf := `
url: "http://localhost:4242/api/"
app: "test-app"
interval: 10
metrics:
  interval: 10
toggles:
  - feature: "test-toggle-path"
    path:
      value: "/john"
      rewrite: "/doe"
`
	testIntegrationRewrite(t, conf, "http://localhost/john", "", "localhost", "/doe", nil, nil)
}

func TestIntegrationRewriteHostWhenToggleIsActive(t *testing.T) {
	conf := `
url: "http://localhost:4242/api/"
app: "test-app"
interval: 10
metrics:
  interval: 10
toggles:
  - feature: "test-toggle-host"
    host:
      value: "localhost"
      rewrite: "whoami1"
`
	testIntegrationRewrite(t, conf, "http://localhost", "", "whoami1", "", nil, nil)
}

func testIntegrationRewrite(t *testing.T, conf, url, userId, expectedHost, expectedPath string, expectedRequestHeaders, expectedResponseHeaders map[string]string) {
	var config unleash.Config
	err := yaml.Unmarshal([]byte(conf), &config)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	unleash, err := unleash.New(ctx, next, &config, "unleash-plugin")
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}

	if userId != "" {
		req.Header.Add("X-Unleash-User-Id", userId)
	}

	unleash.ServeHTTP(recorder, req)

	if expectedHost != "" {
		assert.Equal(t, expectedHost, req.Host)
	}
	if expectedPath != "" {
		assert.Equal(t, expectedPath, req.URL.Path)
	}

	for key, value := range expectedRequestHeaders {
		assert.Equal(t, value, req.Header.Get(key))
	}
	for key, value := range expectedResponseHeaders {
		assert.Equal(t, value, recorder.Header().Get(key))
	}
}
