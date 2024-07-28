package traefik_unleash_plugin

import (
	"context"
	"fmt"
	"github.com/Unleash/unleash-client-go/v4"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"net/url"
	"os"
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
	UserIdHeader    = "X-Unleash-User-Id"
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

type Unleash struct {
	next           http.Handler
	name           string
	featureToggles []FeatureToggle
}

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

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

	logger.Info("Executing unleash plugin")
	for _, toggle := range u.featureToggles {
		logger.Info(fmt.Sprintf("Evaluating feature flag: %s", toggle.feature))
		if toggle.appliesToRequest(req) {
			logger.Info(fmt.Sprintf("Executing feature flag: %s", toggle.feature))
			toggle.setHeaders(rw, req)
			toggle.rewritePath(req)
			if toggle.rewriteHost(rw, req) {
				return
			}
			break
		}
	}
	u.next.ServeHTTP(rw, req)
}

func intervalFrom(interval *int) time.Duration {
	if interval != nil {
		return time.Duration(*interval) * time.Second
	}
	return time.Second * DefaultInterval
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
