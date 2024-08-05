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
	OfflineMode bool `yaml:"offline_mode"`
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

	return toggles
}

func intervalFrom(interval *int) time.Duration {
	if interval != nil {
		return time.Duration(*interval) * time.Second
	}
	return time.Second * DefaultInterval
}
