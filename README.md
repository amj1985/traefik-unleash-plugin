[![Build status](https://github.mpi-internal.com/scmspain/traefik-unleash-plugin/actions/workflows/main.yml/badge.svg)](https://github.mpi-internal.com/scmspain/traefik-unleash-plugin/actions/workflows/main.yml)
[![Sonarqube status](https://sonarqube-enterprise.ets.mpi-internal.com/api/project_badges/measure?project=github.mpi-internal.com%3Ascmspain%3Atraefik-unleash-plugin&metric=alert_status&token=sqb_255be217fd8b59bf82d57a4934e36d207860dd29)](https://sonarqube-enterprise.ets.mpi-internal.com/dashboard?id=github.mpi-internal.com%3Ascmspain%3Atraefik-unleash-plugin)
[![Test coverage](https://sonarqube-enterprise.ets.mpi-internal.com/api/project_badges/measure?project=github.mpi-internal.com%3Ascmspain%3Atraefik-unleash-plugin&metric=coverage&token=sqb_255be217fd8b59bf82d57a4934e36d207860dd29)](https://sonarqube-enterprise.ets.mpi-internal.com/dashboard?id=github.mpi-internal.com%3Ascmspain%3Atraefik-unleash-plugin)
[![Code Smells](https://sonarqube-enterprise.ets.mpi-internal.com/api/project_badges/measure?project=github.mpi-internal.com%3Ascmspain%3Atraefik-unleash-plugin&metric=code_smells&token=sqb_255be217fd8b59bf82d57a4934e36d207860dd29)](https://sonarqube-enterprise.ets.mpi-internal.com/dashboard?id=github.mpi-internal.com%3Ascmspain%3Atraefik-unleash-plugin)
[![Vulnerabilities](https://sonarqube-enterprise.ets.mpi-internal.com/api/project_badges/measure?project=github.mpi-internal.com%3Ascmspain%3Atraefik-unleash-plugin&metric=vulnerabilities&token=sqb_255be217fd8b59bf82d57a4934e36d207860dd29)](https://sonarqube-enterprise.ets.mpi-internal.com/dashboard?id=github.mpi-internal.com%3Ascmspain%3Atraefik-unleash-plugin)
[![Security Rating](https://sonarqube-enterprise.ets.mpi-internal.com/api/project_badges/measure?project=github.mpi-internal.com%3Ascmspain%3Atraefik-unleash-plugin&metric=security_rating&token=sqb_255be217fd8b59bf82d57a4934e36d207860dd29)](https://sonarqube-enterprise.ets.mpi-internal.com/dashboard?id=github.mpi-internal.com%3Ascmspain%3Atraefik-unleash-plugin)
[![Latest delivery](https://badger.engprod-pro.mpi-internal.com/badge/delivery/scmspain/traefik-unleash-plugin)](https://badger.engprod-pro.mpi-internal.com/redirect/delivery/scmspain/traefik-unleash-plugin)
[![Badger](https://badger.engprod-pro.mpi-internal.com/badge/engprod/scmspain/traefik-unleash-plugin)](https://badger.engprod-pro.mpi-internal.com/redirect/engprod/scmspain/traefik-unleash-plugin)

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
git clone https://github.mpi-internal.com/scmspain/traefik-unleash-plugin.git
```

1. Start the services with Docker 

```bash
docker compose up -d
```
