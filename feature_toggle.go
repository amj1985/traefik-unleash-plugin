package traefik_unleash_plugin

import (
	"context"
	"fmt"
	"github.com/Unleash/unleash-client-go/v4"
	uctx "github.com/Unleash/unleash-client-go/v4/context"
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

type PathRewrite struct {
	pathMatcher *regexp.Regexp
	rewriteRule string
}

type HostRewrite struct {
	hostMatcher *regexp.Regexp
	rewriteRule string
}

type HeaderModifier struct {
	headerName  string
	headerValue string
	context     string
}

type FeatureToggle struct {
	pathRewrite     *PathRewrite
	feature         string
	hostRewrite     *HostRewrite
	headerModifiers []*HeaderModifier
}

func (t *FeatureToggle) enabled(r *http.Request) bool {
	userId := r.Header.Get(UserIdHeader)
	if userId != "" {
		ctx := uctx.Context{
			UserId: userId,
		}
		return unleash.IsEnabled(t.feature, unleash.WithContext(ctx))
	}
	return unleash.IsEnabled(t.feature)
}

func (t *FeatureToggle) rewriteHost(next http.Handler, req *http.Request) (http.Handler, *http.Request) {
	if t.hostRewrite != nil {
		logger.Info(fmt.Sprintf("Toggle with feature flag: %s rewriteRule current hostRewrite with pathMatcher: %s for: %s", t.feature, req.Host, t.hostRewrite.rewriteRule))
		redirect := &url.URL{
			Host:   parseHost(t.hostRewrite.rewriteRule),
			Scheme: parseScheme(t.hostRewrite.rewriteRule),
		}
		var rr = req.Clone(context.Background())
		rr.Host = redirect.Host
		return httputil.NewSingleHostReverseProxy(redirect), rr
	}
	return next, req
}

func (t *FeatureToggle) setHeaders(rw http.ResponseWriter, req *http.Request) {
	if t.headerModifiers != nil {
		for _, header := range t.headerModifiers {
			switch header.context {
			case RequestHeader:
				logger.Info(fmt.Sprintf("Toggle with feature flag: %s set request header: %s with pathMatcher: %s", t.feature, header.headerName, header.headerValue))
				req.Header.Set(header.headerName, header.headerValue)
			case ResponseHeader:
				logger.Info(fmt.Sprintf("Toggle with feature flag: %s set response header: %s with pathMatcher: %s", t.feature, header.headerName, header.headerValue))
				rw.Header().Set(header.headerName, header.headerValue)
			}
		}
	}
}

func (t *FeatureToggle) rewritePath(req *http.Request) {
	if t.pathRewrite != nil {
		logger.Info(fmt.Sprintf("Toggle with feature flag: %s rewriteRule current pathRewrite with pathMatcher: %s for: %s", t.feature, req.URL.Path, t.pathRewrite.rewriteRule))
		req.URL.Path = replaceNamedParams(t.pathRewrite.pathMatcher, req.URL.Path, t.pathRewrite.rewriteRule)
		req.RequestURI = req.URL.RequestURI()
	}
}

func (t *FeatureToggle) appliesToRequest(req *http.Request) bool {
	return (t.hostRewrite == nil || t.hostRewrite.hostMatcher.MatchString(req.Host)) &&
		(t.pathRewrite == nil || t.pathRewrite.pathMatcher.MatchString(req.URL.Path)) &&
		(t.headerModifiers == nil || len(t.headerModifiers) > 0) &&
		t.enabled(req)
}

func parseHost(rewrite string) string {
	parsedURL, _ := url.Parse(rewrite)
	if parsedURL.Host == "" {
		return rewrite
	}
	return parsedURL.Host
}

func parseScheme(rewrite string) string {
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
