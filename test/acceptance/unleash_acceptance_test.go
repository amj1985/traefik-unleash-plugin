package unleash_acceptance

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"strings"
	"testing"
	fixture "traefik_unleash/test"
)

func TestMain(m *testing.M) {
	fixture.Setup(m)
}

func TestAcceptanceRewriteHostAndHeadersAndPathWhenToggleIsActiveUsingUserId(t *testing.T) {
	headers := map[string]string{
		"X-Unleash-User-Id": "12345",
	}
	expectedRequestHeaders := map[string]string{
		"X-Foo": "Bar",
	}
	expectedResponseHeaders := map[string]string{
		"X-Served-By": "whoami2",
	}
	testAcceptanceRewrite(t, http.MethodGet, "/foo", headers, "localhost", "GET /bar HTTP/1.1", "Host: whoami2", expectedRequestHeaders, expectedResponseHeaders)
}

func TestAcceptanceRewriteHostAndPathWhenToggleIsActive(t *testing.T) {
	testAcceptanceRewrite(t, http.MethodGet, "/bar", nil, "localhost", "GET /foo HTTP/1.1", "Host: whoami2", nil, nil)
}

func TestAcceptanceRewritePathWhenToggleIsActive(t *testing.T) {
	testAcceptanceRewrite(t, http.MethodGet, "/john", nil, "localhost", "GET /doe HTTP/1.1", "Host: localhost", nil, nil)
}

func TestAcceptanceRewriteHostWhenToggleIsActive(t *testing.T) {
	testAcceptanceRewrite(t, http.MethodGet, "", nil, "localhost", "", "Host: whoami1", nil, nil)
}

func testAcceptanceRewrite(t *testing.T, method, path string, headers map[string]string, hostname, expectedPath, expectedHost string, expectedRequestHeaders, expectedResponseHeaders map[string]string) {
	response, body, _ := doRequest(method, path, headers, hostname)

	responseBody := string(body)
	hasPathRedirected := strings.Contains(responseBody, expectedPath)
	hasHostRedirected := strings.Contains(responseBody, expectedHost)

	assert.True(t, hasPathRedirected)
	assert.True(t, hasHostRedirected)
	for key, value := range expectedRequestHeaders {
		var existsRequestHeader = strings.Contains(responseBody, fmt.Sprintf("%s: %s", key, value))
		assert.True(t, existsRequestHeader)
	}
	for key, value := range expectedResponseHeaders {
		var headerValue = response.Header.Get(key)
		assert.Equal(t, value, headerValue)
	}
}

func doRequest(method, path string, headers map[string]string, hostname string) (*http.Response, []byte, error) {
	request, err := http.NewRequest(method, fmt.Sprintf("http://%s%s", hostname, path), nil)
	if err != nil {
		return nil, nil, err
	}
	for k, v := range headers {
		request.Header.Add(k, v)
	}
	request.Host = hostname

	httpClient := &http.Client{}
	response, err := httpClient.Do(request)
	if err != nil {
		return response, nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, nil, err
	}

	return response, body, err
}
