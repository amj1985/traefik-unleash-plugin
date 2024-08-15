package traefik_unleash_plugin

import (
	"context"
	"fmt"
	"github.com/Unleash/unleash-client-go/v4"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"time"
)

const DefaultInterval = 10

type Config struct {
	Url      string `yaml:"url"`
	App      string `yaml:"app"`
	Interval *int   `yaml:"interval"`
	Metrics  *struct {
		Interval *int `yaml:"interval"`
	} `json:"metrics"`
	Toggles []struct {
		Headers *[]struct {
			Key     string `yaml:"headerName"`
			Value   string `yaml:"pathMatcher"`
			Context string `yaml:"context"`
		} `json:"headerModifiers"`
		Path *struct {
			Value   string `yaml:"pathMatcher"`
			Rewrite string `yaml:"rewriteRule"`
		} `json:"pathRewrite"`
		Host *struct {
			Value   string `yaml:"pathMatcher"`
			Rewrite string `yaml:"rewriteRule"`
		} `json:"hostRewrite"`
		Feature string `yaml:"feature"`
	} `json:"toggles"`
	OfflineMode bool `yaml:"offlineMode"`
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

	if config.OfflineMode {
		return &Unleash{
			next:           next,
			name:           name,
			featureToggles: []FeatureToggle{},
		}, nil
	}
	u := uuid.New()

	if err := unleash.Initialize(
		unleash.WithRefreshInterval(intervalFrom(config.Interval)),
		unleash.WithMetricsInterval(intervalFrom(config.Metrics.Interval)),
		unleash.WithAppName(config.App),
		unleash.WithUrl(config.Url),
		unleash.WithInstanceId(u.String()),
	); err != nil {
		_ = unleash.Close()
		return nil, err
	}

	unleash.WaitForReady()

	return &Unleash{
		next:           next,
		name:           name,
		featureToggles: readConfig(config),
	}, nil
}

func (u *Unleash) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	logger.Info("Executing unleash plugin")
	next := u.next
	for _, toggle := range u.featureToggles {
		logger.Info(fmt.Sprintf("Evaluating feature flag: %s", toggle.feature))
		if toggle.appliesToRequest(req) {
			logger.Info(fmt.Sprintf("Executing feature flag: %s", toggle.feature))
			toggle.setHeaders(rw, req)
			toggle.rewritePath(req)
			next, req = toggle.rewriteHost(next, req)
			break
		}
	}
	next.ServeHTTP(rw, req)
}

func readConfig(config *Config) []FeatureToggle {
	var toggles []FeatureToggle
	for _, t := range config.Toggles {
		var path *PathRewrite
		if t.Path != nil {
			path = &PathRewrite{
				pathMatcher: regexp.MustCompile(t.Path.Value),
				rewriteRule: t.Path.Rewrite,
			}
		}
		var host *HostRewrite
		if t.Host != nil {
			host = &HostRewrite{
				hostMatcher: regexp.MustCompile(t.Host.Value),
				rewriteRule: t.Host.Rewrite,
			}
		}
		var headersCollection []*HeaderModifier
		if t.Headers != nil {
			for _, h := range *t.Headers {
				headersCollection = append(headersCollection, &HeaderModifier{
					headerName:  h.Key,
					headerValue: h.Value,
					context:     h.Context,
				})
			}
		}
		toggles = append(toggles, FeatureToggle{
			pathRewrite:     path,
			hostRewrite:     host,
			feature:         t.Feature,
			headerModifiers: headersCollection,
		})
	}

	return toggles
}

func intervalFrom(interval *int) time.Duration {
	if interval != nil {
		return time.Duration(*interval) * time.Second
	}
	return time.Second * DefaultInterval
}
