package traefik_unleash_plugin

import (
	"context"
	"fmt"
	"github.com/Unleash/unleash-client-go/v4"
	unleashContext "github.com/Unleash/unleash-client-go/v4/context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
)

const (
	SchemeHTTP     = "http"
	SchemeHTTPS    = "https"
	RequestHeader  = "request"
	ResponseHeader = "response"
	UserIdHeader   = "X-Unleash-User-Id"
)

type Path struct {
	value   *regexp.Regexp
	rewrite string
}

type Host struct {
	value   *regexp.Regexp
	rewrite string
}

type Header struct {
	key     string
	value   string
	context string
}

type FeatureToggle struct {
	path    *Path
	feature string
	host    *Host
	headers []*Header
}

func (t *FeatureToggle) enabled(r *http.Request) bool {
	userId := r.Header.Get(UserIdHeader)
	if userId != "" {
		ctx := unleashContext.Context{
			UserId: userId,
		}
		return unleash.IsEnabled(t.feature, unleash.WithContext(ctx))
	}
	return unleash.IsEnabled(t.feature)
}

func (t *FeatureToggle) rewriteHost(rw http.ResponseWriter, req *http.Request) bool {
	if t.host != nil {
		logger.Info(fmt.Sprintf("Toggle with feature flag: %s rewrite current host with value: %s for: %s", t.feature, req.Host, t.host.rewrite))
		var redirectUrl = &url.URL{
			Host:   hostFrom(t.host.rewrite),
			Scheme: schemeFrom(t.host.rewrite),
		}
		var newRequest = req.Clone(context.Background())
		newRequest.Host = redirectUrl.Host
		var nextHandler = httputil.NewSingleHostReverseProxy(redirectUrl)
		nextHandler.ServeHTTP(rw, newRequest)
		return true
	}
	return false
}

func (t *FeatureToggle) setHeaders(rw http.ResponseWriter, req *http.Request) {
	if t.headers != nil {
		for _, header := range t.headers {
			switch header.context {
			case RequestHeader:
				logger.Info(fmt.Sprintf("Toggle with feature flag: %s set request header: %s with value: %s", t.feature, header.key, header.value))
				req.Header.Set(header.key, header.value)
			case ResponseHeader:
				logger.Info(fmt.Sprintf("Toggle with feature flag: %s set response header: %s with value: %s", t.feature, header.key, header.value))
				rw.Header().Set(header.key, header.value)
			}
		}
	}
}

func (t *FeatureToggle) rewritePath(req *http.Request) {
	if t.path != nil {
		logger.Info(fmt.Sprintf("Toggle with feature flag: %s rewrite current path with value: %s for: %s", t.feature, req.URL.Path, t.path.rewrite))
		req.URL.Path = replaceNamedParams(t.path.value, req.URL.Path, t.path.rewrite)
		req.RequestURI = req.URL.RequestURI()
	}
}

func (t *FeatureToggle) appliesToRequest(req *http.Request) bool {
	return (t.host == nil || t.host.value.MatchString(req.Host)) &&
		(t.path == nil || t.path.value.MatchString(req.URL.Path)) &&
		(t.headers == nil || len(t.headers) > 0) &&
		t.enabled(req)
}

func hostFrom(rewrite string) string {
	parsedURL, _ := url.Parse(rewrite)
	if parsedURL.Host == "" {
		return rewrite
	}
	return parsedURL.Host
}

func schemeFrom(rewrite string) string {
	parsedURL, _ := url.Parse(rewrite)
	if isValidScheme(parsedURL.Scheme) {
		return parsedURL.Scheme
	}
	return SchemeHTTP
}

func isValidScheme(scheme string) bool {
	return scheme != "" && (scheme == SchemeHTTP || scheme == SchemeHTTPS)
}

func replaceNamedParams(r *regexp.Regexp, path string, rewrite string) string {
	m := r.FindStringSubmatch(path)
	if len(m) > 0 {
		for i, name := range r.SubexpNames() {
			if len(name) > 0 {
				rewrite = strings.Replace(rewrite, ":"+name, m[i], -1)
			}
		}
	}
	remainingPath := r.ReplaceAllString(path, "")
	return rewrite + remainingPath
}
