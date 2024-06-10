package unleash

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Unleash/unleash-client-go/v4"
	unleashContext "github.com/Unleash/unleash-client-go/v4/context"
	"github.com/google/uuid"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	SchemeHTTP      = "http"
	SchemeHTTPS     = "https"
	DefaultInterval = 10
	RequestHeader   = "request"
	ResponseHeader  = "response"
)

type LogEntry struct {
	Message string    `json:"message"`
	Date    time.Time `json:"date"`
}

type Config struct {
	Url      string `yaml:"url"`
	App      string `yaml:"app"`
	Interval *int   `yaml:"interval"`
	Metrics  *struct {
		Interval *int `yaml:"interval"`
	} `json:"metrics"`
	Toggles []struct {
		Headers *[]struct {
			Key     string `yaml:"key"`
			Value   string `yaml:"value"`
			Context string `yaml:"context"`
		} `json:"headers"`
		Path *struct {
			Value   string `yaml:"value"`
			Rewrite string `yaml:"rewrite"`
		} `json:"path"`
		Host *struct {
			Value   string `yaml:"value"`
			Rewrite string `yaml:"rewrite"`
		} `json:"host"`
		Feature string `yaml:"feature"`
	} `json:"toggles"`
}

func CreateConfig() *Config {
	return &Config{}
}

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

func (toggle *FeatureToggle) enabled(r *http.Request) bool {
	userId := r.Header.Get("X-Unleash-User-Id")
	if userId != "" {
		ctx := unleashContext.Context{
			UserId: userId,
		}
		return unleash.IsEnabled(toggle.feature, unleash.WithContext(ctx))
	}
	return unleash.IsEnabled(toggle.feature)
}

type Unleash struct {
	next           http.Handler
	name           string
	featureToggles []FeatureToggle
}

func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	u := uuid.New()
	err := unleash.Initialize(
		unleash.WithRefreshInterval(intervalFrom(config.Interval)),
		unleash.WithMetricsInterval(intervalFrom(config.Metrics.Interval)),
		unleash.WithAppName(config.App),
		unleash.WithUrl(config.Url),
		unleash.WithInstanceId(u.String()),
	)
	if err != nil {
		_ = unleash.Close()
		return nil, err
	}
	unleash.WaitForReady()
	var toggles []FeatureToggle
	for _, t := range config.Toggles {
		var path *Path
		if t.Path != nil {
			path = &Path{
				value:   regexp.MustCompile(t.Path.Value),
				rewrite: t.Path.Rewrite,
			}
		}
		var host *Host
		if t.Host != nil {
			host = &Host{
				value:   regexp.MustCompile(t.Host.Value),
				rewrite: t.Host.Rewrite,
			}
		}
		var headersCollection []*Header
		if t.Headers != nil {
			for _, h := range *t.Headers {
				headersCollection = append(headersCollection, &Header{
					key:     h.Key,
					value:   h.Value,
					context: h.Context,
				})
			}
		}
		toggles = append(toggles, FeatureToggle{
			path:    path,
			host:    host,
			feature: t.Feature,
			headers: headersCollection,
		})
	}
	return &Unleash{
		next:           next,
		name:           name,
		featureToggles: toggles,
	}, nil
}

func (u *Unleash) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	fmt.Println(jsonMessageFrom("Executing unleash plugin"))
	for _, toggle := range u.featureToggles {
		fmt.Println(jsonMessageFrom(fmt.Sprintf("Evaluating feature flag: %s", toggle.feature)))
		if evaluateFeatureToggle(toggle, req) {
			fmt.Println(jsonMessageFrom(fmt.Sprintf("Executing feature flag: %s", toggle.feature)))
			evaluateHeadersFromToggle(rw, req, toggle)
			evaluatePathFromToggle(toggle, req)
			evaluateHostFromToggle(rw, req, toggle)
			break
		}
	}
	u.next.ServeHTTP(rw, req)
}

func evaluateHeadersFromToggle(rw http.ResponseWriter, req *http.Request, toggle FeatureToggle) {
	if toggle.headers != nil {
		for _, header := range toggle.headers {
			switch header.context {
			case RequestHeader:
				fmt.Println(jsonMessageFrom(fmt.Sprintf("Toggle with feature flag: %s set request header: %s with value: %s", toggle.feature, header.key, header.value)))
				req.Header.Set(header.key, header.value)
			case ResponseHeader:
				fmt.Println(jsonMessageFrom(fmt.Sprintf("Toggle with feature flag: %s set response header: %s with value: %s", toggle.feature, header.key, header.value)))
				rw.Header().Set(header.key, header.value)
			}
		}
	}
}

func evaluatePathFromToggle(toggle FeatureToggle, req *http.Request) {
	if toggle.path != nil {
		fmt.Println(jsonMessageFrom(fmt.Sprintf("Toggle with feature flag: %s rewrite current path with value: %s for: %s", toggle.feature, req.URL.Path, toggle.path.rewrite)))
		req.URL.Path = replaceNamedParams(toggle.path.value, req.URL.Path, toggle.path.rewrite)
		req.RequestURI = req.URL.RequestURI()
	}
}

func evaluateHostFromToggle(rw http.ResponseWriter, req *http.Request, toggle FeatureToggle) {
	if toggle.host != nil {
		fmt.Println(jsonMessageFrom(fmt.Sprintf("Toggle with feature flag: %s rewrite current host with value: %s for: %s", toggle.feature, req.Host, toggle.host.rewrite)))
		var redirectUrl = &url.URL{
			Host:   hostFrom(toggle.host.rewrite),
			Scheme: schemeFrom(toggle.host.rewrite),
		}
		fmt.Println(jsonMessageFrom(fmt.Sprintf("Redirect url with value: %s", redirectUrl.String())))
		req.Host = redirectUrl.Host
		var nextHandler = httputil.NewSingleHostReverseProxy(redirectUrl)
		nextHandler.ServeHTTP(rw, req)
	}
}

func intervalFrom(interval *int) time.Duration {
	if interval != nil {
		return time.Duration(*interval) * time.Second
	}
	return time.Second * DefaultInterval
}

func jsonMessageFrom(message string) string {
	result, _ := json.Marshal(LogEntry{Message: message, Date: time.Now()})
	return string(result)
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

func evaluateFeatureToggle(toggle FeatureToggle, req *http.Request) bool {
	return (toggle.host == nil || toggle.host.value.MatchString(req.Host)) && (toggle.path == nil || toggle.path.value.MatchString(req.URL.Path)) && toggle.enabled(req)
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
	return rewrite
}
