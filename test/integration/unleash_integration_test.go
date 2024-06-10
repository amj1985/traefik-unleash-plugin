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

func TestIntegrationRewriteHostAndPathWhenToggleIsActiveUsingUserId(t *testing.T) {
	conf := `
url: "http://localhost:4242/api/"
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
`
	testIntegrationRewrite(t, conf, "http://localhost/foo", "12345", "whoami2", "/bar")
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
	testIntegrationRewrite(t, conf, "http://localhost/bar", "", "whoami2", "/foo")
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
	testIntegrationRewrite(t, conf, "http://localhost/john", "", "localhost", "/doe")
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
	testIntegrationRewrite(t, conf, "http://localhost", "", "whoami1", "")
}

func testIntegrationRewrite(t *testing.T, conf, url, userId, expectedHost, expectedPath string) {
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
}
