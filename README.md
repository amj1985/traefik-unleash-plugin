[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=amj1985_traefik-unleash-plugin&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=amj1985_traefik-unleash-plugin)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=amj1985_traefik-unleash-plugin&metric=coverage)](https://sonarcloud.io/summary/new_code?id=amj1985_traefik-unleash-plugin)
[![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=amj1985_traefik-unleash-plugin&metric=code_smells)](https://sonarcloud.io/summary/new_code?id=amj1985_traefik-unleash-plugin)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=amj1985_traefik-unleash-plugin&metric=bugs)](https://sonarcloud.io/summary/new_code?id=amj1985_traefik-unleash-plugin)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=amj1985_traefik-unleash-plugin&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=amj1985_traefik-unleash-plugin)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=amj1985_traefik-unleash-plugin&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=amj1985_traefik-unleash-plugin)
[![Duplicated Lines (%)](https://sonarcloud.io/api/project_badges/measure?project=amj1985_traefik-unleash-plugin&metric=duplicated_lines_density)](https://sonarcloud.io/summary/new_code?id=amj1985_traefik-unleash-plugin)

# Traefik Unleash Plugin Middleware

This repository contains a Traefik plugin that installs middleware to intercept requests and query the Unleash server for feature flag status. The plugin determines whether a feature is active or inactive, and it also supports feature evaluation by user ID.

## Features

- Intercepts HTTP requests and checks feature flag status from Unleash.
- Evaluates feature flags globally and by specific user ID.
- Includes a comprehensive `docker-compose` setup for acceptance testing.
- Allows for the rewriting of the host in HTTP requests.
- Allows for the rewriting of the path in HTTP requests.

### Prerequisites

- Docker
- Docker Compose

### Installation

```yaml
http:
  serversTransports:
    default:
      insecureSkipVerify: true
  services:
    unleash:
      loadBalancer:
        servers:
          - url: "http://whoami"
        passHostHeader: false
        serversTransport: default
  routers:
    unleash:
      rule: "Host(`localhost`) && Path(`/foo`)"
      entryPoints:
        - web
      middlewares:
        - unleash
      service: unleash
  middlewares:
    unleash:
      plugin:
        unleash:
          url: "http://unleash:4242/api/"
          app: "test-app"
          interval: 10
          metrics:
            interval: 10
          toggles:
            - feature: "test-toggle"
              path:
                value: "/foo"
                rewrite: "/bar"
              host:
                value: "localhost"
                rewrite: "example.com"
```

1. Clone the repository:

```bash
git clone https://github.com/amj1985/traefik-unleash-plugin.git
```

1. Start the services with Docker

```bash
docker compose up -d
```
