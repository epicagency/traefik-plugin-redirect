# Traefik Plugin - Redirect

`Redirect` is a Traefik plugin to redirect a list with status code.

Based on :
- [Traefik documentation](https://doc.traefik.io/traefik-pilot/plugins/overview/)
- [Traefik plugin example](https://github.com/traefik/plugindemo)
- [Traefik internal redirect plugin](https://github.com/traefik/traefik/blob/master/pkg/middlewares/redirect/redirect.go)

## Installation

Into Traefik static configuration

### TOML
```toml
[entryPoints]
  [entryPoints.web]
    address = ":80"

[pilot]
  token = "xxxxxxxxx"

[experimental.plugins]
  [experimental.plugins.redirect]
    modulename = "gitlab.com/evolves-fr/traefik-plugin-redirect"
    version = "v1.0.0"
```

### YAML
```yaml
entryPoints:
  web:
    address: :80

pilot:
    token: xxxxxxxxx

experimental:
  plugins:
    redirect:
      modulename: gitlab.com/evolves-fr/traefik-plugin-redirect
      version: v1.0.0
```

### CLI
```shell
--entryPoints.web.address=:80
--pilot.token=xxxxxxxxx
--experimental.plugins.redirect.modulename=gitlab.com/evolves-fr/traefik-plugin-redirect
--experimental.plugins.redirect.version=v1.0.0
```

## Configuration

Into Traefik dynamic configuration

### Docker
```yaml
labels:
  - "traefik.http.middlewares.my-redirect.plugin.redirect.redirects[0].regex=/301"
  - "traefik.http.middlewares.my-redirect.plugin.redirect.redirects[0].replacement=/moved-permanently"
  - "traefik.http.middlewares.my-redirect.plugin.redirect.redirects[0].statusCode=301"
  - "traefik.http.middlewares.my-redirect.plugin.redirect.redirects[1].regex=/not-found"
  - "traefik.http.middlewares.my-redirect.plugin.redirect.redirects[1].statusCode=404"
```

### Kubernetes
```yaml
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: my-redirect
spec:
  plugin:
    redirect:
      redirects:
        - regex: /301
          replacement: /moved-permanently
          statusCode: 301
        - regex: /not-found
          statusCode: 404
```

### TOML
```toml
[http.middlewares]
  [http.middlewares.my-redirect.plugin.redirect]
    [[http.middlewares.my-redirect.plugin.redirect.redirects]]
      regex = "/301"
      replacement = "/moved-permanently"
      statusCode = "301"
    [[http.middlewares.my-redirect.plugin.redirect.redirects]]
      regex = "/not-found"
      statusCode = "404"
```

### YAML
```yaml
http:
  middlewares:
    my-redirect:
      plugin:
        redirect:
          redirects:
            - regex: /301
              replacement: /moved-permanently
              statusCode: 301
            - regex: /not-found
              statusCode: 404
```
