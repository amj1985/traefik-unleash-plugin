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
	} `yaml:"metrics"`
	Toggles []struct {
		HeaderModifiers *[]struct {
			HeaderName  string `yaml:"headerName"`
			HeaderValue string `yaml:"headerValue"`
			Context     string `yaml:"context"`
		} `yaml:"headerModifiers"`
		PathRewrite *struct {
			PathMatcher string `yaml:"pathMatcher"`
			RewriteRule string `yaml:"rewriteRule"`
		} `yaml:"pathRewrite"`
		HostRewrite *struct {
			HostMatcher string `yaml:"hostMatcher"`
			RewriteRule string `yaml:"rewriteRule"`
		} `yaml:"hostRewrite"`
		Feature string `yaml:"feature"`
	} `yaml:"toggles"`
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
		if t.PathRewrite != nil {
			path = &PathRewrite{
				pathMatcher: regexp.MustCompile(t.PathRewrite.PathMatcher),
				rewriteRule: t.PathRewrite.RewriteRule,
			}
		}
		var host *HostRewrite
		if t.HostRewrite != nil {
			host = &HostRewrite{
				hostMatcher: regexp.MustCompile(t.HostRewrite.HostMatcher),
				rewriteRule: t.HostRewrite.RewriteRule,
			}
		}
		var headersCollection []*HeaderModifier
		if t.HeaderModifiers != nil {
			for _, h := range *t.HeaderModifiers {
				headersCollection = append(headersCollection, &HeaderModifier{
					headerName:  h.HeaderName,
					headerValue: h.HeaderValue,
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
