[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=amj1985_traefik-unleash-plugin&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=amj1985_traefik-unleash-plugin)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=amj1985_traefik-unleash-plugin&metric=coverage)](https://sonarcloud.io/summary/new_code?id=amj1985_traefik-unleash-plugin)
[![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=amj1985_traefik-unleash-plugin&metric=code_smells)](https://sonarcloud.io/summary/new_code?id=amj1985_traefik-unleash-plugin)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=amj1985_traefik-unleash-plugin&metric=bugs)](https://sonarcloud.io/summary/new_code?id=amj1985_traefik-unleash-plugin)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=amj1985_traefik-unleash-plugin&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=amj1985_traefik-unleash-plugin)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=amj1985_traefik-unleash-plugin&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=amj1985_traefik-unleash-plugin)
[![Duplicated Lines (%)](https://sonarcloud.io/api/project_badges/measure?project=amj1985_traefik-unleash-plugin&metric=duplicated_lines_density)](https://sonarcloud.io/summary/new_code?id=amj1985_traefik-unleash-plugin)

# Traefik Unleash Middleware

This Traefik middleware validates requests against an Unleash (feature flag) server and rewrites the path or host of the request based on input parameters defined in the YAML configuration file.

## Table of Contents

- [Installation](#installation)
- [Configuration](#configuration)
    - [Input Parameters](#input-parameters)
- [Usage](#usage)
- [Example](#example)
- [Contributing](#contributing)
- [License](#license)

## Installation

1. Clone this repository:
    ```bash
    git clone https://github.com/amj1985/traefik-unleash-plugin.git
    cd traefik-unleash-plugin
    ```

2. Follow the [Traefik instructions for installing plugins](https://doc.traefik.io/traefik/plugins/overview/).

## Configuration

To configure this middleware, you need to define the parameters in the `dynamic.yml` file.

### Input Parameters

| Parameter              | Type   | Required   | Description                                                          |
|------------------------|--------|------------|----------------------------------------------------------------------|
| `url`                  | string | Yes        | URL of the Unleash server                                            |
| `app`                  | string | Yes        | Name of the application in Unleash                                   |
| `interval`             | int    | No         | Update interval in seconds                                           |
| `metrics.interval`     | int    | No         | Metrics reporting interval in seconds                                |
| `toggles`              | list   | Yes        | List of feature flag toggles                                         |
| `toggles[].feature`    | string | Yes        | Name of the feature flag                                             |
| `toggles[].path.value` | string | No         | Path to be validated                                                 |
| `toggles[].path.rewrite` | string | No       | Path to redirect to if the feature flag is active                    |
| `toggles[].host.value` | string | No         | Host to be validated                                                 |
| `toggles[].host.rewrite` | string | No       | Host to redirect to if the feature flag is active                    |

## Usage

1. Define the configuration in the `dynamic.yml` file:

    ```yaml
    unleash:
      url: "http://unleash:4242/api/"
      app: "test-app"
      interval: 10
      metrics:
        interval: 10
      toggles:
        - feature: "test-toggle-user-id"
          path:
            value: "/foo"
            rewrite: "/bar"
          host:
            value: "localhost"
            rewrite: "whoami2"
        - feature: "test-toggle"
          path:
            value: "/bar"
            rewrite: "/foo"
          host:
            value: "localhost"
            rewrite: "whoami2"
        - feature: "test-toggle-path"
          path:
            value: "/john"
            rewrite: "/doe"
        - feature: "test-toggle-host"
          host:
            value: "localhost"
            rewrite: "whoami1"
    ```

2. Apply the middleware to your routers in the Traefik configuration:

    ```yaml
    http:
      routers:
        my-router:
          rule: "Host(`example.com`)"
          service: "my-service"
          middlewares: 
            - "unleash"
    ```

## Example

Here is a complete configuration example:

```yaml
http:
  routers:
    my-router:
      rule: "Host(`example.com`)"
      service: "my-service"
      middlewares:
        - "unleash"

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
            - feature: "test-toggle-user-id"
              path:
                value: "/foo"
                rewrite: "/bar"
              host:
                value: "localhost"
                rewrite: "whoami2"
            - feature: "test-toggle"
              path:
                value: "/bar"
                rewrite: "/foo"
              host:
                value: "localhost"
                rewrite: "whoami2"
            - feature: "test-toggle-path"
              path:
                value: "/john"
                rewrite: "/doe"
            - feature: "test-toggle-host"
              host:
                value: "localhost"
                rewrite: "whoami1"
```

## Contributing
Contributions are welcome! If you want to contribute, please open an issue or a pull request in this repository.

## License
This project is licensed under the MIT License.